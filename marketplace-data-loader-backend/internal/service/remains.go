// Package service implements business logic for data synchronization
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/client"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// RemainsService manages Wildberries stock synchronization
type RemainsService struct {
	repo      domain.WbRemainRepository
	client    *client.WbRemainsClient
	logRepo   domain.SyncLogRepository
	pollDelay time.Duration
}

// NewRemainsService returns a new RemainsService instance
func NewRemainsService(
	repo domain.WbRemainRepository,
	client *client.WbRemainsClient,
	logRepo domain.SyncLogRepository,
) *RemainsService {
	return &RemainsService{
		repo:      repo,
		client:    client,
		logRepo:   logRepo,
		pollDelay: 30 * time.Second,
	}
}

// SyncRemains performs the full stock synchronization flow
func (s *RemainsService) SyncRemains(ctx context.Context) error {
	startTime := time.Now()
	var totalProcessed int
	status := "success"
	var errMsg string

	log.Println("[remains] sync started")

	defer func() {
		execTime := int(time.Since(startTime).Seconds())
		_, logErr := s.logRepo.Insert(ctx, domain.SyncLog{
			SyncAt:               startTime,
			Status:               status,
			RecordsCount:         totalProcessed,
			ExecutionTimeSeconds: execTime,
			EntityType:           "remains",
			ErrorMessage:         stringPtr(errMsg),
		})
		if logErr != nil {
			log.Printf("[remains] failed to save log: %v", logErr)
		}
		log.Printf("[remains] sync finished: status=%s records=%d duration=%ds", status, totalProcessed, execTime)
	}()

	// 1. Create report
	taskID, err := s.client.CreateRemainsReport(ctx)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("create report: %v", err)
		return fmt.Errorf("create remains report: %w", err)
	}
	log.Printf("[remains] report task created: %s", taskID)

	// 2. Wait for report completion
	if err := s.waitForReport(ctx, taskID); err != nil {
		status = "error"
		errMsg = fmt.Sprintf("wait report: %v", err)
		return fmt.Errorf("wait for remains report: %w", err)
	}

	// 3. Download report
	body, err := s.client.DownloadRemainsReport(ctx, taskID)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("download report: %v", err)
		return fmt.Errorf("download remains report: %w", err)
	}

	// 4. Normalize data
	remains, err := normalizeRemainsData(body)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("normalize data: %v", err)
		return fmt.Errorf("normalize remains: %w", err)
	}
	log.Printf("[remains] normalized %d records", len(remains))

	// 5. Save to DB
	if len(remains) > 0 {
		count, err := s.repo.UpsertBatch(ctx, remains)
		if err != nil {
			status = "error"
			errMsg = fmt.Sprintf("upsert batch: %v", err)
			return fmt.Errorf("upsert remains batch: %w", err)
		}
		totalProcessed = count
		log.Printf("[remains] saved %d records", totalProcessed)
	}

	return nil
}

// waitForReport polls the report status until it is "done" or fails
func (s *RemainsService) waitForReport(ctx context.Context, taskID string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		status, err := s.client.CheckReportStatus(ctx, taskID)
		if err != nil {
			return fmt.Errorf("check status for task %s: %w", taskID, err)
		}

		switch status {
		case "done":
			return nil
		case "error":
			return fmt.Errorf("report generation failed for task %s", taskID)
		default:
			log.Printf("[remains] report %s status: %s, waiting %s...", taskID, status, s.pollDelay)
		}

		time.Sleep(s.pollDelay)
	}
}

// rawRemainItem mirrors the WB report JSON structure for a single item
type rawRemainItem struct {
	NmID       int64  `json:"nmID"`
	Size       string `json:"size"`
	Barcode    string `json:"barcode"`
	Warehouses []struct {
		WarehouseName string `json:"warehouseName"`
		Quantity      int    `json:"quantity"`
	} `json:"warehouses"`
}

// normalizeRemainsData transforms raw API report bytes into a flat slice of WbRemain
func normalizeRemainsData(body []byte) ([]domain.WbRemain, error) {
	var raw []rawRemainItem
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal remains report: %w", err)
	}

	// Preallocate slice with estimated capacity
	remains := make([]domain.WbRemain, 0, len(raw)*2)

	for _, item := range raw {
		for _, wh := range item.Warehouses {
			if wh.Quantity == 0 {
				continue
			}
			remains = append(remains, domain.WbRemain{
				NmID:      item.NmID,
				Size:      item.Size,
				Warehouse: wh.WarehouseName,
				Quantity:  wh.Quantity,
				Barcode:   stringPtr(item.Barcode),
			})
		}
	}
	return remains, nil
}

// stringPtr returns a pointer to the string, or nil if empty
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

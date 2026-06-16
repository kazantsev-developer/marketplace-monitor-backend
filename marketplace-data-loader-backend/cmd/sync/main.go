// Package main is the entry point for WB synchronization jobs.
// Usage: go run ./cmd/sync --entity=orders|remains|cards
package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/client"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/config"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/repository"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/service"
)

func main() {
	entity := flag.String("entity", "", "entity to sync: orders, remains, cards")
	flag.Parse()

	if *entity == "" {
		log.Fatal("--entity flag is required (orders, remains, cards)")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := repository.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("create db pool: %v", err)
	}
	defer pool.Close()

	syncLogRepo := repository.NewSyncLogRepo(pool)

	switch *entity {
	case "orders":
		runOrders(ctx, cfg, pool, syncLogRepo)
	case "remains":
		runRemains(ctx, cfg, pool, syncLogRepo)
	case "cards":
		runCards(ctx, cfg, pool, syncLogRepo)
	default:
		log.Fatalf("unknown entity: %s (use orders, remains, cards)", *entity)
	}
}

func runOrders(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, logRepo *repository.SyncLogRepo) {
	wbClient := client.NewWbClient(cfg.WB)
	orderRepo := repository.NewWbOrderRepo(pool)
	svc := service.NewOrdersService(orderRepo, wbClient, logRepo, cfg.WB, cfg.Settings.APILimit)

	if err := svc.SyncOrders(ctx); err != nil {
		log.Fatalf("orders sync failed: %v", err)
	}
}

func runRemains(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, logRepo *repository.SyncLogRepo) {
	remainsClient := client.NewWbRemainsClient(cfg.WB)
	remainsRepo := repository.NewWbRemainRepo(pool)
	svc := service.NewRemainsService(remainsRepo, remainsClient, logRepo)

	if err := svc.SyncRemains(ctx); err != nil {
		log.Fatalf("remains sync failed: %v", err)
	}
}

func runCards(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, logRepo *repository.SyncLogRepo) {
	cardsClient := client.NewWbCardsClient(cfg.WB)
	cardsRepo := repository.NewWbCardRepo(pool)
	svc := service.NewCardsService(cardsRepo, cardsClient, logRepo, cfg.Settings.BatchSize, cfg.Settings.BatchDelayMs)

	if err := svc.SyncCards(ctx); err != nil {
		log.Fatalf("cards sync failed: %v", err)
	}
}

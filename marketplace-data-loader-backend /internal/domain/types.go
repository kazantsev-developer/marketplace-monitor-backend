// Package domain предоставляет основные бизнес-сущности и типы.
package domain

type DailyChartItem struct {
	Date  string
	Count int
}

type SyncCursorState struct {
	UpdatedAt string
	NmID      int64
}

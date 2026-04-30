package model

import (
	"fmt"
	"time"
)

func EnsureCurrentPartition() {
	now := time.Now()
	createMonthPartition(now)
}

func EnsureNextMonthPartition() {
	
	nextMonth := time.Now().AddDate(0, 1, 0)
	createMonthPartition(nextMonth)
}

func createMonthPartition(t time.Time) {
	year, month, _ := t.Date()
	partitionName := fmt.Sprintf("logs_%d_%02d", year, month)
	
	
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	
	end := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF logs FOR VALUES FROM ('%s') TO ('%s')",
		partitionName, start, end,
	)

	if err := DB.Exec(query).Error; err != nil {
		fmt.Printf("Error creating partition %s: %v\n", partitionName, err)
	} else {
		fmt.Printf("Partition checked/created: %s\n", partitionName)
	}
}
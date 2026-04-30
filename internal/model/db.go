package model

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DB_URL")
	DB, _ = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	
	DB.AutoMigrate(&User{}, &Service{})

	
	createPartitionedTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id BIGSERIAL,
		service_id UUID NOT NULL,
		raw_message TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		PRIMARY KEY (id, created_at)
	) PARTITION BY RANGE (created_at);`

	DB.Exec(createPartitionedTable)

	
	EnsureCurrentPartition()
	log.Println("DB connected")
}
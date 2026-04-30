package model

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL is not set in .env")
	}

	var err error

	DB, err = gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

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
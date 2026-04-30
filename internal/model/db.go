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

	
	err = DB.AutoMigrate(&User{}, &Service{}, &Log{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("DB connected")
}
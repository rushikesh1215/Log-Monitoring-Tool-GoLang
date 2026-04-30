package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	
	gin.SetMode(gin.DebugMode)

	
	r := gin.Default()

	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive"})
	})

	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
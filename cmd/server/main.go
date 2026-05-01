package main

import (
	"log"
	"log-monitor/internal/model"
	"log-monitor/internal/router"
	"log-monitor/internal/ws"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	model.InitDB()

	c := cron.New() 
	
	c.AddFunc("0 0 25 * *", func() {
		log.Println("Running cron: Creating next month's partition...")
		model.EnsureNextMonthPartition()
	})

	c.Start()

	gin.SetMode(gin.DebugMode)

	
	go ws.GlobalHub.Run() 

    r := gin.Default()
    

    r.SetTrustedProxies(nil)

	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive"})
	})

	router.SetupRouter(r)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
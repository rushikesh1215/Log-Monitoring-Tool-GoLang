package router

import (
	"log-monitor/internal/handler"
	"log-monitor/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", handler.LoginHandler)
	}
	ingest := v1.Group("/ingest")
	ingest.Use(middleware.AgentKeyMiddleware())
	{
    ingest.POST("", handler.IngestHandler)
	}
	
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware()) 
	{
		
		protected.GET("/services", handler.GetServicesHandler)
        protected.GET("/services/:id/logs", handler.GetInitialLogsHandler)
	}
}
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

	
	protected := v1.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware()) 
	{
		
		protected.GET("/services", handler.GetServicesHandler)
	}
}
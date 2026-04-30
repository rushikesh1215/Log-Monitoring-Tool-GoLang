package handler

import (
	"log-monitor/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetServicesHandler(c *gin.Context) {
	var services []model.Service


	result := model.DB.Order("created_at desc").Find(&services)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve services from database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":    len(services),
		"services": services,
	})
}
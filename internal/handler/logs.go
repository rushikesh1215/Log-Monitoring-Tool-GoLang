package handler

import (
	"log-monitor/internal/model"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ParsedLog struct {
	ID        int64     `json:"id"`
	IP        string    `json:"ip"`
	Node      string    `json:"node"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func GetInitialLogsHandler(c *gin.Context) {
	serviceID := c.Param("id")
	
	
	timeLimit := time.Now().Add(-18 * time.Hour)

	var rawLogs []model.Log
	
	err := model.DB.Where("service_id = ? AND created_at > ?", serviceID, timeLimit).
		Order("created_at ASC").
		Find(&rawLogs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	
	var parsedResponse []ParsedLog
	for _, l := range rawLogs {
		parts := strings.Split(l.RawMessage, "|")
			
		logEntry := ParsedLog{
			ID:        l.ID,
			CreatedAt: l.CreatedAt,
		}

		if len(parts) >= 4 {
			logEntry.IP = strings.TrimSpace(parts[0])
			logEntry.Node = strings.TrimSpace(parts[1])
			logEntry.Level = strings.TrimSpace(parts[2])
			logEntry.Message = strings.TrimSpace(parts[3])
		} else {
			logEntry.Message = l.RawMessage 
		}

		parsedResponse = append(parsedResponse, logEntry)
	}

	c.JSON(http.StatusOK, gin.H{
		"service_id": serviceID,
		"count":      len(parsedResponse),
		"logs":       parsedResponse,
	})
}
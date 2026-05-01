package handler

import (
	"encoding/json"
	"log-monitor/internal/model"
	"log-monitor/internal/ws"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LogBatchRequest struct {
	ServiceName string   `json:"service_name" binding:"required"`
	Logs        []string `json:"logs" binding:"required"`
}

func IngestHandler(c *gin.Context) {
	var req LogBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid batch format"})
		return
	}

	
	var service model.Service
	err := model.DB.Where("name = ?", req.ServiceName).FirstOrCreate(&service, model.Service{
		ID:   uuid.New(),
		Name: req.ServiceName,
	}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync service"})
		return
	}

	
	var logEntries []model.Log
	currentTime := time.Now()

	for _, rawLine := range req.Logs {
		logEntries = append(logEntries, model.Log{
			ServiceID:  service.ID,
			RawMessage: rawLine,
			CreatedAt:  currentTime,
		})
	}
	
	if err := model.DB.Create(&logEntries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save logs"})
		return
	}

	
	for _, entry := range logEntries {
		parts := strings.Split(entry.RawMessage, "|")
		
		
		liveLog := ParsedLog{
			ID:        entry.ID,
			CreatedAt: entry.CreatedAt,
		}

		if len(parts) >= 4 {
			liveLog.IP = strings.TrimSpace(parts[0])
			liveLog.Node = strings.TrimSpace(parts[1])
			liveLog.Level = strings.TrimSpace(parts[2])
			liveLog.Message = strings.TrimSpace(parts[3])
		} else {
			liveLog.Message = entry.RawMessage
		}

		
		messageBytes, err := json.Marshal(liveLog)
		if err == nil {
			
			ws.GlobalHub.Broadcast <- ws.BroadcastMessage{
				ServiceID: service.ID.String(),
				Message:   messageBytes,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"ingested": len(logEntries),
	})
}
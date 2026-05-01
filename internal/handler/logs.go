package handler

import (
	"log-monitor/internal/model"
	"net/http"
	"strconv"
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

func GetFilteredLogsHandler(c *gin.Context) {
	serviceID := c.Param("id")
	level := c.Query("level")
	search := c.Query("search")
	timeRange := c.Query("range")
	
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 { page = 1 }
	offset := (page - 1) * pageSize

	var rawLogs []model.Log
	db := model.DB.Where("service_id = ?", serviceID)

	
	if level != "" {
		db = db.Where("level = ?", level)
	}
	if search != "" {
		db = db.Where("raw_message ILIKE ?", "%"+search+"%")
	}

	now := time.Now()
	startTime := now.Add(-18 * time.Hour) 
	if timeRange != "" {
		switch timeRange {
		case "1h": startTime = now.Add(-1 * time.Hour)
		case "6h": startTime = now.Add(-6 * time.Hour)
		case "24h": startTime = now.Add(-24 * time.Hour)
		case "7d": startTime = now.AddDate(0, 0, -7)
		}
	}
	db = db.Where("created_at > ?", startTime)

	var total int64
	db.Model(&model.Log{}).Count(&total)

	err := db.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&rawLogs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
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
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"logs":       parsedResponse,
	})
}
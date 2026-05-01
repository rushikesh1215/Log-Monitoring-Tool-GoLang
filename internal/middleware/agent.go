package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func AgentKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedKey := os.Getenv("AGENT_SECRET_KEY")
		clientKey := c.GetHeader("X-Agent-Key")

		if clientKey == "" || clientKey != expectedKey {
			log.Println("Agent Key Unauthorized")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized Agent"})
			return
		}
		c.Next()
	}
}
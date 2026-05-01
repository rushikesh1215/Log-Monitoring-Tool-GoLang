package handler

import (
	"log-monitor/internal/ws"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, 
}

func WebSocketHandler(c *gin.Context) {
	serviceID := c.Param("id")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{Conn: conn, Send: make(chan []byte, 256)}
	ws.GlobalHub.Register <- ws.Subscription{ServiceID: serviceID, Client: client}

	
	go func() {
		defer func() {
			ws.GlobalHub.Unregister <- ws.Subscription{ServiceID: serviceID, Client: client}
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	
	go func() {
		for message := range client.Send {
			conn.WriteMessage(websocket.TextMessage, message)
		}
	}()
}
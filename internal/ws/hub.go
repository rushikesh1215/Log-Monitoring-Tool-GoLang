package ws

import (
	"sync"
	"github.com/gorilla/websocket"
)


type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

type BroadcastMessage struct {
	ServiceID string
	Message   []byte
}

type Subscription struct {
	ServiceID string
	Client    *Client
}

type Hub struct {
	// Rooms map
	Rooms      map[string]map[*Client]bool
	Broadcast  chan BroadcastMessage
	Register   chan Subscription
	Unregister chan Subscription
	mu         sync.Mutex
}

var GlobalHub = Hub{
	Rooms:      make(map[string]map[*Client]bool),
	Broadcast:  make(chan BroadcastMessage),
	Register:   make(chan Subscription),
	Unregister: make(chan Subscription),
}

func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.Register:
			h.mu.Lock()
			if h.Rooms[sub.ServiceID] == nil {
				h.Rooms[sub.ServiceID] = make(map[*Client]bool)
			}
			h.Rooms[sub.ServiceID][sub.Client] = true
			h.mu.Unlock()

		case sub := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Rooms[sub.ServiceID]; ok {
				delete(h.Rooms[sub.ServiceID], sub.Client)
				close(sub.Client.Send)
				if len(h.Rooms[sub.ServiceID]) == 0 {
					delete(h.Rooms, sub.ServiceID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.Lock()
			clients := h.Rooms[msg.ServiceID]
			for client := range clients {
				select {
				case client.Send <- msg.Message:
				default:
					close(client.Send)
					delete(clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}
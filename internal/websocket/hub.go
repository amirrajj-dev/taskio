package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	Clients     map[uuid.UUID]*Client          // All connected clients
	Rooms       map[string]map[uuid.UUID]*Client // Clients in each room
	Register    chan *Client
	Unregister  chan *Client
	Subscribe   chan *Subscription
	Unsubscribe chan *Subscription
	Broadcast   chan *Message
	mu          sync.RWMutex
}

type Subscription struct {
	ClientID uuid.UUID
	Room     string
	UserID   uuid.UUID
}

func NewHub() *Hub {
	return &Hub{
		Clients:     make(map[uuid.UUID]*Client),
		Rooms:       make(map[string]map[uuid.UUID]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Subscribe:   make(chan *Subscription),
		Unsubscribe: make(chan *Subscription),
		Broadcast:   make(chan *Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client %s (user %s) connected", client.ID, client.UserID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				
				// Remove from all rooms
				for room, clients := range h.Rooms {
					if _, ok := clients[client.ID]; ok {
						delete(h.Rooms[room], client.ID)
						if len(h.Rooms[room]) == 0 {
							delete(h.Rooms, room)
						}
					}
				}
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client %s disconnected", client.ID)

		case sub := <-h.Subscribe:
			h.mu.Lock()
			if _, ok := h.Rooms[sub.Room]; !ok {
				h.Rooms[sub.Room] = make(map[uuid.UUID]*Client)
			}
			if client, ok := h.Clients[sub.ClientID]; ok {
				h.Rooms[sub.Room][client.ID] = client
				client.Rooms[sub.Room] = true
			}
			h.mu.Unlock()
			log.Printf("Client %s subscribed to room %s", sub.ClientID, sub.Room)

		case sub := <-h.Unsubscribe:
			h.mu.Lock()
			if clients, ok := h.Rooms[sub.Room]; ok {
				delete(clients, sub.ClientID)
				if len(clients) == 0 {
					delete(h.Rooms, sub.Room)
				}
			}
			if client, ok := h.Clients[sub.ClientID]; ok {
				delete(client.Rooms, sub.Room)
			}
			h.mu.Unlock()
			log.Printf("Client %s unsubscribed from room %s", sub.ClientID, sub.Room)

		case message := <-h.Broadcast:
			h.mu.RLock()
			var targetClients []*Client

			if message.TargetID != "" {
				// Send to specific user
				targetUserID, _ := uuid.Parse(message.TargetID)
				for _, client := range h.Clients {
					if client.UserID == targetUserID {
						targetClients = append(targetClients, client)
					}
				}
			} else if message.Room != "" {
				// Send to all clients in a room
				if clients, ok := h.Rooms[message.Room]; ok {
					for _, client := range clients {
						targetClients = append(targetClients, client)
					}
				}
			} else {
				// Send to all clients
				for _, client := range h.Clients {
					targetClients = append(targetClients, client)
				}
			}
			h.mu.RUnlock()

			// Send message to target clients
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				continue
			}

			for _, client := range targetClients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					h.mu.Lock()
					delete(h.Clients, client.ID)
					h.mu.Unlock()
				}
			}
		}
	}
}
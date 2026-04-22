package websocket

import (
	"log"
	"sync"
	"time"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

type Client struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Conn      *websocket.Conn
	Hub       *Hub
	Send      chan []byte
	Rooms     map[string]bool // Rooms the client is subscribed to
	mu        sync.RWMutex
	LastPing  time.Time
}

type Message struct {
	Type      string      `json:"type"`      // event type: "task_assigned", "comment_added", etc.
	Payload   interface{} `json:"payload"`   // the actual data
	Room      string      `json:"room"`      // room to broadcast to (optional)
	TargetID  string      `json:"target_id"` // specific user or resource
}

func NewClient(userID uuid.UUID, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:       uuid.New(),
		UserID:   userID,
		Conn:     conn,
		Hub:      hub,
		Send:     make(chan []byte, 256),
		Rooms:    make(map[string]bool),
		LastPing: time.Now(),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.LastPing = time.Now()
		return nil
	})

	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle subscription messages
		switch msg.Type {
		case "subscribe":
			c.Hub.Subscribe <- &Subscription{
				ClientID: c.ID,
				Room:     msg.Room,
				UserID:   c.UserID,
			}
		case "unsubscribe":
			c.Hub.Unsubscribe <- &Subscription{
				ClientID: c.ID,
				Room:     msg.Room,
				UserID:   c.UserID,
			}
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
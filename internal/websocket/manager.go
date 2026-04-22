package websocket

import (
	"net/http"

	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development, restrict in production
		return configs.Configs.App.GO_ENV != "production" || r.Header.Get("Origin") == configs.Configs.FRONTEND_URL
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketManager struct {
	Hub *Hub
}

var WsManager *WebSocketManager

func NewWebSocketManager() *WebSocketManager {
	hub := NewHub()
	go hub.Run()
	WsManager = &WebSocketManager{Hub: hub}
	return WsManager
}

func (m *WebSocketManager) HandleWebSocket(c *gin.Context) {
	// Get JWT token from query parameter or header
	tokenString := c.Query("token")
	if tokenString == "" {
		tokenString = c.GetHeader("Sec-WebSocket-Protocol")
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(configs.Configs.JWT.JWT_SECRET), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		return
	}

	userIDStr, ok := claims["id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id format"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upgrade connection"})
		return
	}

	// Create and register client
	client := NewClient(userID, conn, m.Hub)
	m.Hub.Register <- client

	// Start pumps
	go client.WritePump()
	go client.ReadPump()
}

// BroadcastToUser sends a message to a specific user
func (m *WebSocketManager) BroadcastToUser(userID uuid.UUID, eventType string, payload interface{}) {
	m.Hub.Broadcast <- &Message{
		Type:     eventType,
		Payload:  payload,
		TargetID: userID.String(),
	}
}

// BroadcastToRoom sends a message to all clients in a room
func (m *WebSocketManager) BroadcastToRoom(room string, eventType string, payload interface{}) {
	m.Hub.Broadcast <- &Message{
		Type:    eventType,
		Payload: payload,
		Room:    room,
	}
}

// BroadcastToAll sends a message to all connected clients
func (m *WebSocketManager) BroadcastToAll(eventType string, payload interface{}) {
	m.Hub.Broadcast <- &Message{
		Type:    eventType,
		Payload: payload,
	}
}
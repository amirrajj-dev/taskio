package websocket

import (
	"log"
	"sync"

	"github.com/google/uuid"
)

type EventPublisher struct {
	manager *WebSocketManager
	queue   chan *PublishEvent
	wg      sync.WaitGroup
}

type PublishEvent struct {
	UserID    uuid.UUID
	EventType string
	Payload   interface{}
}

var (
	Publisher     *EventPublisher
	publisherOnce sync.Once
)

func InitPublisher(manager *WebSocketManager) {
	publisherOnce.Do(func() {
		Publisher = &EventPublisher{
			manager: manager,
			queue:   make(chan *PublishEvent, 1000), // Buffered queue
		}
		go Publisher.start()
	})
}

func (p *EventPublisher) start() {
	for event := range p.queue {
		if p.manager != nil {
			p.manager.BroadcastToUser(event.UserID, event.EventType, event.Payload)
		}
	}
}

// Async publish - doesn't block the main operation
func (p *EventPublisher) Publish(userID uuid.UUID, eventType string, payload interface{}) {
	select {
	case p.queue <- &PublishEvent{
		UserID:    userID,
		EventType: eventType,
		Payload:   payload,
	}:
	default:
		log.Printf("WebSocket publisher queue full, dropping event for user %s", userID)
	}
}

// Batch publish to multiple users
func (p *EventPublisher) PublishToMany(userIDs []uuid.UUID, eventType string, payload interface{}) {
	for _, userID := range userIDs {
		p.Publish(userID, eventType, payload)
	}
}
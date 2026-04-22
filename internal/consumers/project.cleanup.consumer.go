package consumers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/amirrajj-dev/taskio/internal/constants"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
)

type ProjectCleanupEvent struct {
	ProjectID     uuid.UUID `json:"project_id"`
	OrgID         uuid.UUID `json:"org_id"`
	DeletedBy     uuid.UUID `json:"deleted_by"`
	Timestamp     time.Time `json:"timestamp"`
}

func StartProjectCleanupConsumer(conn *amqp.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("project cleanup worker failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		constants.PROJECT_CLEANUP_QUEUE,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("project cleanup worker failed to declare queue: %v", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("project cleanup worker failed to set QoS: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("project cleanup worker failed to register consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			var cleanupEvent ProjectCleanupEvent
			if err := json.Unmarshal(msg.Body, &cleanupEvent); err != nil {
				log.Printf("project cleanup: bad event format: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("Starting cleanup for project: %s", cleanupEvent.ProjectID)
			
			if err := performProjectCleanup(cleanupEvent.ProjectID); err != nil {
				log.Printf("Failed to cleanup project %s: %v", cleanupEvent.ProjectID, err)
				msg.Nack(false, true) // Requeue for retry
				continue
			}

			log.Printf("Successfully cleaned up project: %s", cleanupEvent.ProjectID)
			msg.Ack(false)
		}
	}()

	log.Println("Project cleanup worker running...")
	select {}
}

func performProjectCleanup(projectID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx := database.PG.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Delete comments (they reference tasks)
	// Using subquery to delete all comments for tasks under this project
	if err := tx.WithContext(ctx).
		Where("task_id IN (SELECT id FROM tasks WHERE project_id = ?)", projectID).
		Delete(&models.Comment{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// 2. Delete all tasks (without worrying about order since no FK constraints)
	if err := tx.WithContext(ctx).
		Where("project_id = ?", projectID).
		Delete(&models.Task{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Project itself is deleted separately in the main service
	
	if err := tx.Commit().Error; err != nil {
		return err
	}
	
	return nil
}
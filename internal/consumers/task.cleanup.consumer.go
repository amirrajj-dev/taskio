package consumers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/amirrajj-dev/taskio/internal/constants"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type TaskCleanupEvent struct {
	TaskID     uuid.UUID `json:"task_id"`
	DeletedBy  uuid.UUID `json:"deleted_by"`
	ProjectID  uuid.UUID `json:"project_id"`
	Timestamp  time.Time `json:"timestamp"`
}

func StartTaskCleanupConsumer(conn *amqp.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("task cleanup worker failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		constants.TASK_CLEANUP_QUEUE,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("task cleanup worker failed to declare queue: %v", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("task cleanup worker failed to set QoS: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("task cleanup worker failed to register consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			var cleanupEvent TaskCleanupEvent
			if err := json.Unmarshal(msg.Body, &cleanupEvent); err != nil {
				log.Printf("task cleanup: bad event format: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("Starting cleanup for task: %s (and its subtasks)", cleanupEvent.TaskID)

			// Retry logic with exponential backoff
			var err error
			for retries := 0; retries < 3; retries++ {
				err = performTaskCleanup(cleanupEvent.TaskID)
				if err == nil {
					break
				}

				backoff := time.Duration(1<<uint(retries)) * time.Second
				log.Printf("Cleanup attempt %d failed for task %s: %v, retrying in %v",
					retries+1, cleanupEvent.TaskID, err, backoff)
				time.Sleep(backoff)
			}

			if err != nil {
				log.Printf("Failed to cleanup task %s after 3 attempts: %v", cleanupEvent.TaskID, err)
				msg.Nack(false, true) // Requeue for later retry
				continue
			}

			log.Printf("Successfully cleaned up task: %s and all its subtasks", cleanupEvent.TaskID)
			msg.Ack(false)
		}
	}()

	log.Println("Task cleanup worker running...")
	select {}
}

func performTaskCleanup(taskID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx := database.PG.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Get all subtask IDs recursively (to delete their comments too)
	subtaskIDs, err := getAllSubtaskIDs(ctx, tx, taskID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Combine parent task ID with all subtask IDs
	allTaskIDs := append([]uuid.UUID{taskID}, subtaskIDs...)

	// 2. Delete comments for all tasks (parent + subtasks)
	if err := deleteTaskComments(ctx, tx, allTaskIDs); err != nil {
		tx.Rollback()
		return err
	}

	// 3. Delete all subtasks (parent task will be deleted separately)
	if err := deleteSubtasks(ctx, tx, subtaskIDs); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// getAllSubtaskIDs recursively fetches all subtask IDs for a given task
func getAllSubtaskIDs(ctx context.Context, tx *gorm.DB, parentTaskID uuid.UUID) ([]uuid.UUID, error) {
	var subtaskIDs []uuid.UUID
	
	// Get direct subtasks
	var directSubtasks []models.Task
	err := tx.WithContext(ctx).
		Where("parent_task_id = ?", parentTaskID).
		Select("id").
		Find(&directSubtasks).Error
	if err != nil {
		return nil, err
	}

	for _, subtask := range directSubtasks {
		subtaskIDs = append(subtaskIDs, subtask.ID)
		
		// Recursively get nested subtasks
		nestedIDs, err := getAllSubtaskIDs(ctx, tx, subtask.ID)
		if err != nil {
			return nil, err
		}
		subtaskIDs = append(subtaskIDs, nestedIDs...)
	}

	return subtaskIDs, nil
}

func deleteTaskComments(ctx context.Context, tx *gorm.DB, taskIDs []uuid.UUID) error {
	if len(taskIDs) == 0 {
		return nil
	}
	
	return tx.WithContext(ctx).
		Where("task_id IN ?", taskIDs).
		Delete(&models.Comment{}).Error
}

func deleteSubtasks(ctx context.Context, tx *gorm.DB, subtaskIDs []uuid.UUID) error {
	if len(subtaskIDs) == 0 {
		return nil
	}
	
	return tx.WithContext(ctx).
		Where("id IN ?", subtaskIDs).
		Delete(&models.Task{}).Error
}
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
	"gorm.io/gorm"
)

type TeamCleanupEvent struct {
	TeamID     uuid.UUID `json:"team_id"`
	OrgID      uuid.UUID `json:"org_id"`
	DeletedBy  uuid.UUID `json:"deleted_by"`
	Timestamp  time.Time `json:"timestamp"`
}

func StartTeamCleanupConsumer(conn *amqp.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("team cleanup worker failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		constants.TEAM_CLEANUP_QUEUE,
		true,  
		false, 
		false, 
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("team cleanup worker failed to declare queue: %v", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("team cleanup worker failed to set QoS: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("team cleanup worker failed to register consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			var cleanupEvent TeamCleanupEvent
			if err := json.Unmarshal(msg.Body, &cleanupEvent); err != nil {
				log.Printf("team cleanup: bad event format: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("Starting cleanup for team: %s", cleanupEvent.TeamID)
			
			if err := performTeamCleanup(cleanupEvent.TeamID); err != nil {
				log.Printf("Failed to cleanup team %s: %v", cleanupEvent.TeamID, err)
				msg.Nack(false, true) // Requeue for retry
				continue
			}

			log.Printf("Successfully cleaned up team: %s", cleanupEvent.TeamID)
			msg.Ack(false)
		}
	}()

	log.Println("Team cleanup worker running...")
	select {}
}

func performTeamCleanup(teamID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx := database.PG.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete in correct order (child first, then parent)
	
	// 1. First, get all projects under this team
	var projectIDs []uuid.UUID
	if err := tx.WithContext(ctx).
		Table("projects").
		Select("id").
		Where("team_id = ?", teamID).
		Scan(&projectIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. Delete comments from tasks under these projects
	for _, projectID := range projectIDs {
		if err := deleteTeamProjectComments(ctx, tx, projectID); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 3. Delete tasks from projects under this team
	for _, projectID := range projectIDs {
		if err := deleteTeamProjectTasks(ctx, tx, projectID); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 4. Delete projects under this team
	if err := deleteTeamProjects(ctx, tx, teamID); err != nil {
		tx.Rollback()
		return err
	}

	// 5. Delete team members (should be empty by now, but just in case)
	if err := deleteTeamMembers(ctx, tx, teamID); err != nil {
		tx.Rollback()
		return err
	}

	// Note: Team itself is deleted separately in the main service
	// because you already call s.teamRepo.DeleteTeam()
	
	if err := tx.Commit().Error; err != nil {
		return err
	}
	
	return nil
}

func deleteTeamProjectComments(ctx context.Context, tx *gorm.DB, projectID uuid.UUID) error {
	// Delete comments from tasks under this project
	return tx.WithContext(ctx).
		Table("comments").
		Where("task_id IN (?)", 
			tx.Table("tasks").
				Select("id").
				Where("project_id = ?", projectID),
		).
		Delete(&models.Comment{}).Error
}

func deleteTeamProjectTasks(ctx context.Context, tx *gorm.DB, projectID uuid.UUID) error {
	// Delete all tasks for this project
	return tx.WithContext(ctx).
		Where("project_id = ?", projectID).
		Delete(&models.Task{}).Error
}

func deleteTeamProjects(ctx context.Context, tx *gorm.DB, teamID uuid.UUID) error {
	// Delete all projects under this team
	return tx.WithContext(ctx).
		Where("team_id = ?", teamID).
		Delete(&models.Project{}).Error
}

func deleteTeamMembers(ctx context.Context, tx *gorm.DB, teamID uuid.UUID) error {
	// Delete all team members (should be empty, but cleanup just in case)
	return tx.WithContext(ctx).
		Where("team_id = ?", teamID).
		Delete(&models.TeamMember{}).Error
}
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

type OrgCleanupEvent struct {
	OrgID     uuid.UUID `json:"org_id"`
	DeletedBy uuid.UUID `json:"deleted_by"`
	Timestamp time.Time `json:"timestamp"`
}

func StartOrgCleanupConsumer(conn *amqp.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("org cleanup worker failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		constants.ORG_CLEANUP_QUEUE,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("org cleanup worker failed to declare queue: %v", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("org cleanup worker failed to set QoS: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("org cleanup worker failed to register consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			var cleanupEvent OrgCleanupEvent
			if err := json.Unmarshal(msg.Body, &cleanupEvent); err != nil {
				log.Printf("org cleanup: bad event format: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("Starting cleanup for org: %s", cleanupEvent.OrgID)

			// Retry logic with exponential backoff
			var err error
			for retries := 0; retries < 3; retries++ {
				err = performOrgCleanup(cleanupEvent.OrgID)
				if err == nil {
					break
				}

				backoff := time.Duration(1<<uint(retries)) * time.Second
				log.Printf("Cleanup attempt %d failed for org %s: %v, retrying in %v",
					retries+1, cleanupEvent.OrgID, err, backoff)
				time.Sleep(backoff)
			}

			if err != nil {
				log.Printf("Failed to cleanup org %s after 3 attempts: %v", cleanupEvent.OrgID, err)
				msg.Nack(false, true) // Requeue for later retry
				continue
			}

			log.Printf("Successfully cleaned up org: %s", cleanupEvent.OrgID)
			msg.Ack(false)
		}
	}()

	log.Println("Org cleanup worker running...")
	select {}
}

func performOrgCleanup(orgID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx := database.PG.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Delete comments (through tasks)
	if err := deleteOrgComments(ctx, tx, orgID); err != nil {
		tx.Rollback()
		return err
	}

	// 2. Delete tasks
	if err := deleteOrgTasks(ctx, tx, orgID); err != nil {
		tx.Rollback()
		return err
	}

	// 3. Delete team members (must come before teams)
	if err := deleteOrgTeamMembers(ctx, tx, orgID); err != nil {
		tx.Rollback()
		return err
	}

	// 4. Delete teams
	if err := deleteOrgTeams(ctx, tx, orgID); err != nil {
		tx.Rollback()
		return err
	}

	// 5. Delete projects
	if err := deleteOrgProjects(ctx, tx, orgID); err != nil {
		tx.Rollback()
		return err
	}

	// 6. Delete organization users (members)
	if err := deleteOrgMembers(ctx, tx, orgID); err != nil {
		tx.Rollback()
		return err
	}

	// 7. Delete organization invites (if table exists)
	if err := deleteOrgInvites(ctx, tx, orgID); err != nil {
		tx.Rollback()
		return err
	}

	// Note: Organization itself is deleted separately in the main service

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func deleteOrgComments(ctx context.Context, tx *gorm.DB, orgID uuid.UUID) error {
	// Comments are linked to tasks, tasks to projects, projects to org
	return tx.WithContext(ctx).
		Table("comments").
		Where("task_id IN (?)",
			tx.Table("tasks").
				Select("tasks.id").
				Joins("JOIN projects ON projects.id = tasks.project_id").
				Where("projects.org_id = ?", orgID),
		).
		Delete(&models.Comment{}).Error
}

func deleteOrgTasks(ctx context.Context, tx *gorm.DB, orgID uuid.UUID) error {
	// Tasks are linked to projects, projects to org
	return tx.WithContext(ctx).
		Table("tasks").
		Where("project_id IN (?)",
			tx.Table("projects").
				Select("id").
				Where("org_id = ?", orgID),
		).
		Delete(&models.Task{}).Error
}

func deleteOrgTeamMembers(ctx context.Context, tx *gorm.DB, orgID uuid.UUID) error {
	// Team members are linked to teams, teams to org
	return tx.WithContext(ctx).
		Table("team_members").
		Where("team_id IN (?)",
			tx.Table("teams").
				Select("id").
				Where("org_id = ?", orgID),
		).
		Delete(&models.TeamMember{}).Error
}

func deleteOrgTeams(ctx context.Context, tx *gorm.DB, orgID uuid.UUID) error {
	return tx.WithContext(ctx).
		Where("org_id = ?", orgID).
		Delete(&models.Team{}).Error
}

func deleteOrgProjects(ctx context.Context, tx *gorm.DB, orgID uuid.UUID) error {
	return tx.WithContext(ctx).
		Where("org_id = ?", orgID).
		Delete(&models.Project{}).Error
}

func deleteOrgMembers(ctx context.Context, tx *gorm.DB, orgID uuid.UUID) error {
	return tx.WithContext(ctx).
		Where("org_id = ?", orgID).
		Delete(&models.OrganizationUser{}).Error
}

func deleteOrgInvites(ctx context.Context, tx *gorm.DB, orgID uuid.UUID) error {
	// Check if OrganizationInvite model exists first
	return tx.WithContext(ctx).
		Where("org_id = ?", orgID).
		Delete(&models.OrganizationInvite{}).Error
}

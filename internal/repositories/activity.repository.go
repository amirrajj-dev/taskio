package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityRepository interface {
	CreateActivityFromEvent(ctx context.Context, evt models.ActivityEvent) (*models.ActivityLog, error)
	GetUserActivities(ctx context.Context, userID uuid.UUID) ([]*models.ActivityLog, error)
	DeleteActivity(ctx context.Context, actID uuid.UUID) (bool, error)
}

type activityRepository struct {
	Activities *gorm.DB
}

var ActivityRepo *activityRepository

func NewActivityRepository() {
	ActivityRepo = &activityRepository{
		Activities: database.PG.Model(&models.ActivityLog{}),
	}
}

func (r *activityRepository) CreateActivityFromEvent(ctx context.Context, evt models.ActivityEvent) (*models.ActivityLog, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	log := models.ActivityLog{
		ID:        uuid.New(),
		UserID:    evt.UserID,
		OrgID:     evt.OrgID,
		TaskID:    evt.TaskID,
		ProjectID: evt.ProjectID,
		Action:    evt.EventName,
		CreatedAt: time.Now().UTC(),
	}

	if evt.Metadata != nil {
		body, _ := json.Marshal(evt.Metadata)
		log.Metadata = body
	}

	if err := r.Activities.WithContext(opCtx).Create(&log).Error; err != nil {
		return nil, fmt.Errorf("failed to create activity %s : %w", log.Action, err)
	}
	return &log, nil
}

func (r *activityRepository) GetUserActivities(ctx context.Context, userID uuid.UUID) ([]*models.ActivityLog, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var userActivities []*models.ActivityLog
	res := r.Activities.WithContext(opCtx).Where("user_id = ?", userID).Find(&userActivities)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user has  no activities")
		}
		return nil, fmt.Errorf("failed to fetch user activities")
	}
	return userActivities, nil
}

func (r *activityRepository) DeleteActivity(ctx context.Context, actID uuid.UUID) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Activities.WithContext(opCtx).Where("id = ?", actID).Delete(&models.ActivityLog{})
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("activity not found")
		}
		return false, fmt.Errorf("failed to delete activity")
	}
	if res.RowsAffected == 0 {
		return false, fmt.Errorf("failed to delete activity")
	}
	return true, nil
}

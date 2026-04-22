package services

import (
	"context"
	"log/slog"

	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/google/uuid"
)

type ActivityService interface {
	GetUserActivities(ctx context.Context, userID uuid.UUID) ([]*models.ActivityLog, error)
	DeleteActivity(ctx context.Context, actID uuid.UUID) (bool, error)
}

type activityService struct {
	activityRepo repositories.ActivityRepository
}

var ActivitiesService *activityService

func NewActivityService() {
	ActivitiesService = &activityService{
		activityRepo: repositories.ActivityRepo,
	}
}

func (s *activityService) GetUserActivities(ctx context.Context, userID uuid.UUID) ([]*models.ActivityLog, error) {
	slog.Info("fetching user activities", "user_id", userID)
	
	activities, err := s.activityRepo.GetUserActivities(ctx, userID)
	if err != nil {
		slog.Error("failed to get user activities", "user_id", userID, "error", err)
		return nil, err
	}
	
	slog.Info("user activities fetched", "user_id", userID, "count", len(activities))
	return activities, nil
}

func (s *activityService) DeleteActivity(ctx context.Context, actID uuid.UUID) (bool, error) {
	slog.Info("deleting activity", "activity_id", actID)
	
	deleted, err := s.activityRepo.DeleteActivity(ctx, actID)
	if err != nil {
		slog.Error("failed to delete activity", "activity_id", actID, "error", err)
		return false, err
	}
	
	slog.Info("activity deleted", "activity_id", actID)
	return deleted, nil
}
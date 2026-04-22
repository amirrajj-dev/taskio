package services

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/amirrajj-dev/taskio/internal/events"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/amirrajj-dev/taskio/internal/repositories"
	"github.com/amirrajj-dev/taskio/internal/websocket"
	"github.com/google/uuid"
)

type CommentService interface {
	CreateComment(ctx context.Context, taskID, userID uuid.UUID, content string) (*models.Comment, error)
	GetTaskComments(ctx context.Context, taskID uuid.UUID) ([]*models.Comment, error)
}

type commentService struct {
	commentRepo      repositories.CommentRepository
	taskRepo         repositories.TaskRepository
	userRepo         repositories.UserRepository
	activityProducer *queue.ActivityProducer
}

var CommentsService *commentService

func NewCommentService(producer *queue.ActivityProducer) {
	CommentsService = &commentService{
		commentRepo:      repositories.CommentRepo,
		taskRepo:         repositories.TaskRepo,
		userRepo:         repositories.UserRepo,
		activityProducer: producer,
	}
}

func (s *commentService) CreateComment(ctx context.Context, taskID, userID uuid.UUID, content string) (*models.Comment, error) {
	slog.Info("creating comment", "task_id", taskID, "user_id", userID)
	comment := models.Comment{
		ID:        uuid.New(),
		TaskID:    taskID,
		UserID:    userID,
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now().UTC(),
	}
	createdComment, createErr := s.commentRepo.CreateComment(ctx, comment)
	if createErr != nil {
		slog.Error("failed to create comment", "task_id", taskID, "user_id", userID, "error", createErr)
		return nil, createErr
	}

	slog.Info("comment created successfully", "comment_id", createdComment.ID, "task_id", taskID)

	if websocket.WsManager != nil {
		task, _ := s.taskRepo.FindTaskByID(ctx, taskID)
		user, _ := s.userRepo.FindUserByID(ctx, userID)

		payload := websocket.CommentAddedPayload{
			CommentID: createdComment.ID,
			TaskID:    taskID,
			TaskTitle: task.Title,
			UserID:    userID,
			UserName:  user.FullName,
			Content:   createdComment.Content,
			CreatedAt: createdComment.CreatedAt.Format(time.RFC3339),
		}
		websocket.WsManager.BroadcastToUser(task.AssingedTo, websocket.EventCommentAdded, payload)
		websocket.WsManager.BroadcastToRoom("project:"+task.ProjectID.String(), websocket.EventCommentAdded, payload)
	} else {
		slog.Warn("websocket manager not available", "task_id", taskID, "comment_id", createdComment.ID)

	}

	event := models.ActivityEvent{
		UserID:    &userID,
		TaskID:    &taskID,
		EventName: events.ActivityEvents.CreateComment,
		Metadata: map[string]any{
			"comment_id": createdComment.ID,
			"content":    createdComment.Content,
		},
	}

	if err := s.activityProducer.Publish(event); err != nil {
		slog.Error("failed to publish activity event", "event", events.ActivityEvents.CreateComment, "error", err)
	}

	return createdComment, nil
}

func (s *commentService) GetTaskComments(ctx context.Context, taskID uuid.UUID) ([]*models.Comment, error) {
	slog.Info("fetching task comments", "task_id", taskID)
	comments, err := s.commentRepo.GetTaskComments(ctx, taskID)
	if err != nil {
		slog.Error("failed to fetch task comments", "task_id", taskID, "error", err)
		return nil, err
	}
	slog.Info("task comments fetched", "task_id", taskID, "count", len(comments))
	return comments, nil
}

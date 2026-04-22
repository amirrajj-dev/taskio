package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/amirrajj-dev/taskio/internal/infrastructure/database"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentRepository interface {
	CreateComment(ctx context.Context, comment models.Comment) (*models.Comment, error)
	GetTaskComments(ctx context.Context, taskID uuid.UUID) ([]*models.Comment, error)
	FindCommentByID(ctx context.Context, commentID uuid.UUID) (*models.Comment, error)
	DeleteCommentByID(ctx context.Context, commentID uuid.UUID) (bool, error)
}

type commentRepository struct {
	Comments *gorm.DB
}

var CommentRepo *commentRepository

func NewCommentRepository() {
	CommentRepo = &commentRepository{
		Comments: database.PG.Model(&models.Comment{}),
	}
}

func (r *commentRepository) CreateComment(ctx context.Context, comment models.Comment) (*models.Comment, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := r.Comments.WithContext(opCtx).Create(&comment).Error; err != nil {
		return nil, fmt.Errorf("failed to create comment : %w", err)
	}
	return &comment, nil
}

func (r *commentRepository) GetTaskComments(ctx context.Context, taskID uuid.UUID) ([]*models.Comment, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var comments []*models.Comment
	res := r.Comments.WithContext(opCtx).Where("task_id = ?", taskID).Find(&comments)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task has  no comments")
		}
		return nil, fmt.Errorf("failed to fetch task comments")
	}
	return comments, nil
}

func (r *commentRepository) FindCommentByID(ctx context.Context, commentID uuid.UUID) (*models.Comment, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var comment *models.Comment
	res := r.Comments.WithContext(opCtx).First(&comment, commentID)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to fetch comment")
	}
	return comment, nil
}

func (r *commentRepository) DeleteCommentByID(ctx context.Context, commentID uuid.UUID) (bool, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Comments.WithContext(opCtx).Where("id = ?", commentID).Delete(&models.Comment{})
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("comment not found")
		}
		return false, fmt.Errorf("failed to delete comment")
	}
	if res.RowsAffected == 0 {
		return false, fmt.Errorf("failed to delete commen")
	}
	return true, nil
}

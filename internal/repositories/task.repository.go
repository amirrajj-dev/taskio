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

type TaskRepository interface {
	CreateTask(ctx context.Context, task models.Task) (*models.Task, error)
	FindTaskByTitle(ctx context.Context, title string) (*models.Task, error)
	FindTaskByID(ctx context.Context, taskID uuid.UUID) (*models.Task, error)
	FindProjectTasks(ctx context.Context , projectID uuid.UUID) ([]*models.Task , error)
	UpdateTaskByID(ctx context.Context, taskID uuid.UUID, task models.Task) (*models.Task, error)
	DeleteTaskByID(ctx context.Context , taskID uuid.UUID) (bool , error)
	FindTaskSubTasks(ctx context.Context , taskID uuid.UUID) ([]*models.Task , error)
}

type taskRepository struct {
	Tasks *gorm.DB
}

var TaskRepo *taskRepository

func NewTaskRepository() {
	TaskRepo = &taskRepository{
		Tasks: database.PG.Model(&models.Task{}),
	}
}

func (r *taskRepository) CreateTask(ctx context.Context , task models.Task) (*models.Task, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Tasks.WithContext(opCtx).Create(&task)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to create task : %w", res.Error)
	}
	return &task, nil
}

func (r *taskRepository) FindTaskByTitle(ctx context.Context, title string) (*models.Task, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var task *models.Task
	res := r.Tasks.WithContext(opCtx).Where("title = ?", title).Find(&task)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil , fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to find task by title : %w", res.Error)
	}
	return task, nil
}

func (r *taskRepository) FindTaskByID(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var task models.Task
	res := r.Tasks.WithContext(opCtx).Where("id = ?", taskID).First(&task)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil , fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to find task by title : %w", res.Error)
	}
	return &task, nil
}


func (r *taskRepository) FindProjectTasks(ctx context.Context , projectID uuid.UUID) ([]*models.Task , error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var tasks []*models.Task
	res := r.Tasks.WithContext(opCtx).Where("project_id = ?" , projectID).Find(&tasks)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil , fmt.Errorf("project has no tasks")
		}
		return nil , fmt.Errorf("failed to find project tasks : %w" , res.Error)
	}
	return tasks , nil
}

func (r *taskRepository) UpdateTaskByID(ctx context.Context, taskID uuid.UUID, task models.Task) (*models.Task, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := r.Tasks.WithContext(opCtx).
		Where("id = ?", taskID).
		UpdateColumns(task)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	var updatedTask models.Task
	if err := r.Tasks.WithContext(opCtx).Where("id = ?", taskID).First(&updatedTask).Error; err != nil {
		return nil, err
	}

	return &updatedTask, nil
}


func (r *taskRepository) DeleteTaskByID(ctx context.Context , taskID uuid.UUID) (bool , error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := r.Tasks.WithContext(opCtx).Where("id = ?" , taskID).Delete(&models.Task{})
	if res.Error != nil {
		return false , fmt.Errorf("failed to delete task : %w" , res.Error)
	}
	if res.RowsAffected == 0 {
		return false , fmt.Errorf("task not found")
	}
	return true , nil
}

func (r *taskRepository) FindTaskSubTasks(ctx context.Context , taskID uuid.UUID) ([]*models.Task , error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var tasks []*models.Task
	res := r.Tasks.WithContext(opCtx).Where("parent_task_id = ?" , taskID).Find(&tasks)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return []*models.Task{} , nil
		}
		return nil , fmt.Errorf("failed to find task sub tasks : %w" , res.Error)
	}
	return tasks , nil
}
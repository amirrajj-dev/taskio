package models

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	ProjectID    uuid.UUID `json:"project_id" gorm:"type:uuid;not null" validate:"required"`
	ParentTaskID *uuid.UUID `json:"parent_task_id" gorm:"type:uuid;" validate:"required"`
	Title        string    `json:"title" gorm:"not null" validate:"required"`
	Description  string    `json:"description" validate:"min=5,max=100"`
	Status       string    `json:"status" validate:"required,oneof=todo in_progress review done"`
	Priority     string    `json:"priority" validate:"required,oneof=low medium high urgent"`
	AssingedTo   uuid.UUID `json:"assigned_to" gorm:"type:uuid;not null" validate:"required"`
	DueDate      *time.Time `json:"due_date"`
	CreatedBy    uuid.UUID `json:"created_by" gorm:"type:uuid;not null" validate:"required"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

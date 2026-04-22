package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	OrgID       uuid.UUID `json:"org_id" gorm:"type:uuid;not null" validate:"required"`
	TeamID      uuid.UUID `json:"team_id" gorm:"type:uuid;not null" validate:"required"`
	Name        string    `json:"name" gorm:"not null;unique" validate:"required"`
	Description string    `json:"description" validate:"min=5,max=250"`
	CreatedBy   uuid.UUID `json:"created_by" gorm:"type:uuid;not null" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectResult struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}
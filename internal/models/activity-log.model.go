package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	TaskID    *uuid.UUID `json:"task_id" gorm:"type:uuid;"`
	UserID    *uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	OrgID     *uuid.UUID `json:"org_id" gorm:"type:uuid;"`
	ProjectID *uuid.UUID `json:"project_id" gorm:"type:uuid;"`
	Action    string    `json:"action" validate:"required"`
	Metadata  json.RawMessage `json:"meta_data" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at"`
}

type ActivityEvent struct {
    EventName  string            `json:"event_name"`
    UserID     *uuid.UUID         `json:"user_id"`
    OrgID      *uuid.UUID         `json:"org_id,omitempty"`
    ProjectID  *uuid.UUID         `json:"project_id,omitempty"`
    TaskID     *uuid.UUID         `json:"task_id,omitempty"`
    TeamID     *uuid.UUID         `json:"team_id,omitempty"`
    Metadata   map[string]any    `json:"metadata,omitempty"`
}
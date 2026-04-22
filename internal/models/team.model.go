package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	OrgID     uuid.UUID `json:"org_id"  gorm:"type:uuid;not null" validate:"required"`
	Name      string    `json:"name" gorm:"not null" validate:"required,min=5,max=100"`
	CreatorID uuid.UUID `json:"creator_id"  gorm:"type:uuid"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamMember struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	TeamID    uuid.UUID `json:"team_id" gorm:"type:uuid;not null" validate:"required"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null" validate:"required"`
	Role      string    `json:"role" gorm:"not null" validate:"required,oneof=owner admin member viewer"`
	JoinedAt time.Time `json:"joined_at"`
}

type TeamWithMemberCount struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	MembersCount int       `json:"membersCount" gorm:"column:members_count"`
}

// used for GET /teams/:teamId/members 
type TeamMemberResult struct {
	Name         string    `json:"name"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	ImageUrl  string    `json:"image_url"`
}

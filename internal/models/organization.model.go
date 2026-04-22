package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string    `json:"name" validate:"required,min=4,max=100"`
	OwnerID   uuid.UUID `json:"owner_id" gorm:"type:uuid;not null" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
}

type OrganizationUser struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	OrgID    uuid.UUID `json:"org_id" gorm:"type:uuid;not null" validate:"required"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null" validate:"required"`
	Role     string    `json:"role" validate:"required,oneof=owner admin member"`
	JoinedAt time.Time `json:"joined_at" validate:"required"`
}

type OrganizationUserResult struct {
	Name      string    `json:"name"`
	Role     string    `json:"role"`
	UserID   uuid.UUID `json:"user_id"`
	ImageUrl  string    `json:"image_url"`
	JoinedAt time.Time `json:"joined_at"`
}

type OrganizationInvite struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	OrgID      uuid.UUID `json:"org_id" gorm:"type:uuid;not null;index"`
	Email      string    `json:"email" gorm:"size=255;not null;index" validate:"required,email"`
	InvitedBy  uuid.UUID `json:"invited_by" gorm:"type:uuid;not null" validate:"required"`
	Token      string    `json:"token" gorm:"size=512;unique;not null" validate:"required"`
	ExpiresAt  time.Time `json:"expires_at" gorm:"not null;index" validate:"required"`
	Status     string    `json:"status" gorm:"size=50;not null;index" validate:"required,oneof=invited accepted revoked"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (invite *OrganizationInvite) IsExpired() bool {
    return time.Now().After(invite.ExpiresAt)
}



type CurrentUserOrganizationsResponse struct {
	ID   uuid.UUID `json:"id"` // holds org id and not orgUserID
	Name string    `json:"name"`
	Role string    `json:"role"`
}

type OrganizationWithMembersCount struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	OwnerID      uuid.UUID `json:"owner_id"`
	MembersCount int       `json:"members_count"`
	CreatedAt    time.Time `json:"created_at"`
}

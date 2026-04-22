package websocket

import (
	"time"

	"github.com/google/uuid"
)

// Event types
const (
	EventTaskAssigned                = "task_assigned"
	EventTaskUpdated                 = "task_updated"
	EventTaskDeleted                 = "task_deleted"
	EventCommentAdded                = "comment_added"
	EventProjectCreated              = "project_created"
	EventProjectUpdated              = "project_updated"
	EventProjectDeleted              = "project_deleted"
	EventTeamCreated                 = "team_created"
	EventTeamUpdated                 = "team_updated"
	EventTeamDeleted                 = "team_deleted"
	EventTeamMemberAdded             = "team_member_added"
	EventTeamMemberDeleted           = "team_member_deleted"
	EventTeamMemberRoleChanged       = "team_member_role_changed"
	EventTransferOwnerShip           = "transfer_ownership"
	EventStatusChanged               = "status_changed"
	EventOrganizationUpdated         = "org_updated"
	EventOrganizationDeleted         = "org_deleted"
	EventOrganizationUserRoleChanged = "org_user_role_changed"
)

// Event payloads
type TaskAssignedPayload struct {
	TaskID      uuid.UUID  `json:"task_id"`
	Title       string     `json:"title"`
	AssignedBy  uuid.UUID  `json:"assigned_by"`
	AssignedTo  uuid.UUID  `json:"assigned_to"`
	ProjectID   uuid.UUID  `json:"project_id"`
	ProjectName string     `json:"project_name"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type CommentAddedPayload struct {
	CommentID uuid.UUID `json:"comment_id"`
	TaskID    uuid.UUID `json:"task_id"`
	TaskTitle string    `json:"task_title"`
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	Content   string    `json:"content"`
	CreatedAt string    `json:"created_at"`
}

type TaskDeletedPayload struct {
	TaskID uuid.UUID `json:"task_id"`
	Title  string    `json:"title"`
}

type TaskUpdatedPayload struct {
	TaskID    uuid.UUID   `json:"task_id"`
	Title     string      `json:"title"`
	UpdatedBy uuid.UUID   `json:"updated_by"`
	Changes   interface{} `json:"changes"`
}

type TeamMemberAddedPayload struct {
	TeamID   uuid.UUID `json:"team_id"`
	TeamName string    `json:"team_name"`
	UserID   uuid.UUID `json:"user_id"`
	UserName string    `json:"user_name"`
	Role     string    `json:"role"`
	AddedBy  uuid.UUID `json:"added_by"`
}

type TeamMemberDeletedPayload struct {
	TeamID    uuid.UUID `json:"team_id"`
	TeamName  string    `json:"team_name"`
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	Role      string    `json:"role"`
	DeletedBy uuid.UUID `json:"deleted_by"`
}

type OrganizationUpdatedPayload struct {
	Name string `json:"name"`
}

type OrganizationDeletedPayload struct {
	Name string `json:"name"`
}

type OrganizationUserRoleChanged struct {
	NewRole string `json:"new_role"`
	Name    string `json:"name"`
}

type ProjectCreatedPayload struct {
	ProjectName string    `json:"project_name"`
	ProjectID   uuid.UUID `json:"project_id"`
	CreatedBy   uuid.UUID `json:"created_by"`
}

type ProjectUpdatedPayload struct {
	Updates   interface{} `json:"updates"`
}

type ProjectDeletedPayload struct {
	ProjectName string `json:"project_name"`
	DeletedBy   string `json:"deleted_by"`
}

type TeamCreatedPayload struct {
	Name      string `json:"name"`
	CreatedBy string `json:"created_by"`
}

type TeamUpdatedPayload struct {
	Name string `json:"name"`
}

type TeamDeletedPayload struct {
	Name string `json:"name"`
}

type TeamMemberRoleChangedPayload struct {
	NewRole  string    `json:"new_role"`
	ActorID  uuid.UUID `json:"actor_id"`
	TargetID uuid.UUID `json:"target_id"`
}

type TransferOwnerShipPayload struct {
	ActorID  uuid.UUID `json:"actor_id"`
	NewOwner uuid.UUID `json:"new_owner"`
}

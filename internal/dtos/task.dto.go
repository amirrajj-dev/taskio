package dtos

import "time"

type CreateTaskRequest struct {
	Title       string `json:"title" validate:"min=4,max=150" example:"Implement login API"`
	Description string `json:"description" validate:"max=250" example:"Create JWT authentication endpoint"`
	Priority    string `json:"priority" validate:"required,oneof=low medium high urgent" example:"high"`
	Status      string `json:"status" validate:"required,oneof=todo in_progress review done" example:"todo"`
	AssignedTo  *string `json:"assigned_to" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	DueDate     *time.Time `json:"due_date" example:"2024-12-31T23:59:59Z"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title" validate:"max=150" example:"Updated task title"`
	Description string `json:"description" validate:"max=250" example:"Updated task description"`
	Priority    string `json:"priority" example:"urgent"`
	Status      string `json:"status" example:"done"`
	AssignedTo  *string `json:"assigned_to" example:"550e8400-e29b-41d4-a716-446655440000"`
	DueDate     *time.Time `json:"due_date" example:"2024-12-31T23:59:59Z"`
}

package dtos

type CreateTeamRequest struct {
	Name string `json:"name" validate:"required,min=3,max=80" example:"Frontend Team"`
}

type UpdateTeamRequest struct {
	Name string `json:"name" validate:"required,min=3,max=80" example:"Front Team"`
}

type AddMemberToTeamRequest struct {
	UserID string `json:"user_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type ChangeTeamMemberRoleRequest struct {
	Role  string `json:"role" validate:"required,oneof=owner admin member viewer" example:"admin"`
}
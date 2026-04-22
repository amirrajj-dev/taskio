package dtos

type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=3,max=80" example:"Acme Corp"`
}

type UpdateOrganizationNameRequest struct {
	Name string `json:"name" validate:"required,min=3,max=80" example:"New Acme Corp"`
}

type ChangeMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member" example:"admin"`
}

type InviteToOrganizationRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

type AcceptInviteRequest struct {
    Token string `json:"token" validate:"required" example:"abc123def456..."`
}
package dtos

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=80" example:"Website Redesign"`
	Description string `json:"description" validate:"max=250" example:"Redesign company website with new UI"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" validate:"max=80" example:"Website Redesign v2"`
	Description string `json:"description" validate:"max=250" example:"Updated project scope"`
}

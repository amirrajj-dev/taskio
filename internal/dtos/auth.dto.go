package dtos

// RegisterRequest represents user registration payload
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6,max=12" example:"password123"`
	FullName string `json:"fullName" validate:"required,min=6,max=100" example:"John Doe"`
	Gender   string `json:"gender" validate:"required,oneof=male female" example:"male"`
}

// LoginRequest represents user login payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6,max=12" example:"password123"`
}
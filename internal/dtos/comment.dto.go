package dtos


// CreateCommentRequest represents create comment payload
type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,min=3,max=150" example:"This is a great task!"`
}
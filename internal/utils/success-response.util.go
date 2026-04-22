package utils

import "time"

type SuccessResponse struct {
	Message   string      `json:"message,omitempty"`
	TimeStamp time.Time   `json:"timestamp"`
	Success   bool        `json:"success"`
	Path      string      `json:"path"`
	Data      interface{} `json:"data,omitempty"`
}

type SuccessResponseWithTotal struct {
	Message   string      `json:"message,omitempty"`
	TimeStamp time.Time   `json:"timestamp"`
	Success   bool        `json:"success"`
	Total     int64       `json:"total"`
	Path      string      `json:"path"`
	Data      interface{} `json:"data,omitempty"`
}

func NewSuccessResponse(msg string, data interface{}, path string) SuccessResponse {
	return SuccessResponse{
		Success:   true,
		Message:   msg,
		Data:      data,
		Path:      path,
		TimeStamp: time.Now().UTC(),
	}
}

func NewSuccessResponseWithTotal(msg string, data interface{}, total int64, path string) SuccessResponseWithTotal {
	return SuccessResponseWithTotal{
		Success:   true,
		Message:   msg,
		Data:      data,
		Total:     total,
		Path:      path,
		TimeStamp: time.Now().UTC(),
	}
}
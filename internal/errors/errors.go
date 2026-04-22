package errors

import (
	"errors"
	"strings"
	"time"
	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Errors    []FieldError `json:"errors"`
	Success   bool         `json:"success"`
	Path      string       `json:"path"`
	TimeStamp time.Time    `json:"timestamp"`
}

func NewValidationError(filedErrors []FieldError, path string) *ValidationError {
	return &ValidationError{
		Errors:    filedErrors,
		Success:   false,
		Path:      path,
		TimeStamp: time.Now().UTC(),
	}
}

type BasicError struct {
	Error     string    `json:"error"`
	Success   bool      `json:"success"`
	Path      string    `json:"path"`
	TimeStamp time.Time `json:"timestamp"`
}

func NewBasicError(msg string , path string) *BasicError {
	return &BasicError{
		Error:     msg,
		Success:   false,
		Path:      path,
		TimeStamp: time.Now().UTC(),
	}
}

func ToFieldError(vErr validator.FieldError) FieldError {
	var msg string
	fieldName := strings.Title(vErr.Field())
	switch vErr.Tag() {
	case "required":
		msg = fieldName + " is required."
	case "email":
		msg = "Invalid email format."
	case "min":
		msg = fieldName + " must be at least " + vErr.Param() + " characters long."
	case "max":
		msg = fieldName + " must be at most " + vErr.Param() + " characters long."
	case "oneof":
		msg = fieldName + " must be one of " + vErr.Param() + "."
	case "uuid":
		msg = fieldName + " must be a valid UUID."
	case "url":
		msg = fieldName + " must be a valid URL."
	case "gte":
		msg = fieldName + " must be greater than or equal to " + vErr.Param() + "."
	case "gt":
		msg = fieldName + " must be greater than " + vErr.Param() + "."
	case "lte":
		msg = fieldName + " must be less than or equal to " + vErr.Param() + "."
	case "lt":
		msg = fieldName + " must be less than " + vErr.Param() + "."
	case "len":
		msg = fieldName + " must be exactly " + vErr.Param() + " characters long."
	default:
		msg = "Invalid value for " + fieldName + "."
	}
	return FieldError{
		Field:   vErr.Field(),
		Message: msg,
	}
}


// internal errors
var (
	// users
	ErrUserNotFound = errors.New("user not found")
	ErrRefreshNotFound = errors.New("refresh token not found")
)

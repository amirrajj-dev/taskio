package utils

import (
	"net/http"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func InitValidate(){
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func ValidateStruct(s interface{}) error {
	return Validate.Struct(s)
}

func ValidateRequest(c *gin.Context , req interface{}) error {
	if err := ValidateStruct(req); err != nil {
		var validationErrors []errors.FieldError
		if castedErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range castedErrors {
				validationErrors = append(validationErrors, errors.ToFieldError(fieldErr))
			}
			c.JSON(http.StatusBadRequest, errors.NewValidationError(validationErrors, c.Request.URL.Path))
			return err
		} else {
			c.JSON(http.StatusInternalServerError, errors.NewBasicError("An unexpected validation error occurred.", c.Request.URL.Path))
			return err
		}
	}
	return nil
}
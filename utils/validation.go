package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator"
	"github.com/ronak-89/simple-user-auth/internal/models"
)

func ValidateBinding(err error) []models.ErrorResponse {

	var validationErrors []models.ErrorResponse

	var unmarshalTypeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &unmarshalTypeErr):
		field := unmarshalTypeErr.Field
		expectedType := unmarshalTypeErr.Type.String()
		validationErrors = append(validationErrors, models.ErrorResponse{
			Field:   field,
			Message: fmt.Sprintf("%s must be of type %s", field, expectedType),
		})
	case strings.Contains(err.Error(), "required"):
		fieldName := strings.Split(err.Error(), "'")[1]
		fmt.Println("--------------------", fieldName)
		validationErrors = append(validationErrors, models.ErrorResponse{
			Field:   fieldName,
			Message: fmt.Sprintf("%s field is required", fieldName),
		})
	default:
		validationErrors = append(validationErrors, models.ErrorResponse{
			Field:   "request",
			Message: "Invalid JSON format",
		})
	}
	return validationErrors

}

func Validatefields(err error) []models.ErrorResponse {
	var validationFieldErrors []models.ErrorResponse

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, e := range ve {
			validationFieldErrors = append(validationFieldErrors, models.ErrorResponse{
				Field:   e.Field(),
				Message: GetValidationMessage(e.Field(), e.Tag(), e.Param()),
			})
		}
	}
	return validationFieldErrors
}

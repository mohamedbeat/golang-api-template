package validator

import (
	"golang-api-template/pkg/apperror"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Validate runs validation and returns a *apperror.AppError if it fails
func Validate(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var fieldErrors []apperror.FieldError

	for _, e := range err.(validator.ValidationErrors) {
		fieldErrors = append(fieldErrors, apperror.FieldError{
			Field:   toSnakeCase(e.Field()),
			Message: humanize(e),
		})
	}

	return apperror.ValidationError(fieldErrors)
}

// humanize turns validator tags into readable messages
func humanize(e validator.FieldError) string {

	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + e.Param() + " characters"
	case "max":
		return "must be at most " + e.Param() + " characters"
	case "gte":
		return "must be greater than or equal to " + e.Param()
	case "lte":
		return "must be less than or equal to " + e.Param()
	case "url":
		return "must be a valid URL"
	case "uuid":
		return "must be a valid UUID"
	case "oneof":
		return "must be one of: " + e.Param()
	default:
		return e.Field() + "is invalid"
	}
}

// toSnakeCase converts "FirstName" → "first_name" to match JSON field names
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

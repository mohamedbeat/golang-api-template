package apperror

import "net/http"

// AppError is your custom error that carries HTTP status + error code
type AppError struct {
	StatusCode int          `json:"-"`                 // used to write the HTTP status, not exposed
	Code       string       `json:"code"`              // machine-readable e.g. "USER_NOT_FOUND"
	Message    string       `json:"message"`           // human-readable
	Details    []FieldError `json:"details,omitempty"` // for validation errors
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// --- constructors for common cases ---

func NotFound(resource string) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    resource + " was not found",
	}
}

func BadRequest(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    message,
	}
}

func ValidationError(details []FieldError) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "VALIDATION_ERROR",
		Message:    "validation failed",
		Details:    details,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		Code:       "CONFLICT",
		Message:    message,
	}
}

func Unauthorized(msg string) *AppError {
	if msg == "" {
		msg = "authentication required"
	}
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    msg,
	}
}

func Forbidden() *AppError {
	return &AppError{
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    "you do not have permission to perform this action",
	}
}

func Internal() *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "an unexpected error occurred",
	}
}

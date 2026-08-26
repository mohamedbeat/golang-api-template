package response

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"golang-api-template/pkg/apperror"
)

type envelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
	Error   any  `json:"error,omitempty"`
	Meta    any  `json:"meta,omitempty"`
}

type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, envelope{Success: true, Data: data})
}

func WithMeta(w http.ResponseWriter, status int, data any, meta Meta) {
	write(w, status, envelope{Success: true, Data: data, Meta: meta})
}

func Error(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError

	// if it's our custom error, use it directly
	if errors.As(err, &appErr) {
		write(w, appErr.StatusCode, envelope{Success: false, Error: appErr})
		return
	}

	// anything else is an unexpected internal error — don't leak details
	log.Printf("unexpected error: %v", err)
	internal := apperror.Internal()
	write(w, internal.StatusCode, envelope{Success: false, Error: internal})
}

func write(w http.ResponseWriter, status int, body envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
func OK(w http.ResponseWriter) {
	JSON(w, 200, map[string]string{
		"message": "ok",
	})

}

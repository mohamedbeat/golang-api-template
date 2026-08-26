package users

import (
	"golang-api-template/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}
func (h *UserHandler) Routes(tokenValidator middleware.TokenValidator) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware(tokenValidator))
	r.Use(middleware.Only("user"))

	// r.Get("/{id}", h.GetUser)
	// r.Patch("/{id}", h.UpdateUser)
	// ...
	return r
}

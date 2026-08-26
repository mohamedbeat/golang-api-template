package auth

import (
	"encoding/json"
	"fmt"
	"golang-api-template/internal/auth/dtos"
	"golang-api-template/internal/middleware"
	"golang-api-template/internal/users"
	"golang-api-template/pkg/apperror"
	"golang-api-template/pkg/response"
	"golang-api-template/pkg/validator"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	AuthService AuthService
	UserService users.UserService
}

func NewAuthHandler(authService AuthService, userService users.UserService) *AuthHandler {
	return &AuthHandler{AuthService: authService, UserService: userService}
}

func (h *AuthHandler) Routes(tokenValidator middleware.TokenValidator) chi.Router {
	r := chi.NewRouter()
	// public
	r.Post("/register", h.RegisterUser)
	r.Post("/login", h.LoginUser)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	r.Group(func(r chi.Router) {
	})

	// protected: user
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(tokenValidator))
		r.Use(middleware.Only("user"))
		r.Get("/me", h.Me)
	})

	return r
}

//////////////////////////
// user related
//////////////////////////

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.BadRequest("Invalid request body"))
		return
	}

	if err := validator.Validate(req); err != nil {
		response.Error(w, err)
		return
	}
	created, err := h.AuthService.RegisterUser(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, created)
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req dtos.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.BadRequest("Invalid request body"))
		return
	}

	if err := validator.Validate(req); err != nil {
		response.Error(w, err)
		return
	}

	access, refresh, err := h.AuthService.LoginUser(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"access":  access,
		"refresh": refresh,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {

	refreshToken, err := h.getTokenFromHeader(r)
	if err != nil {
		response.Error(w, err)
		return
	}

	if refreshToken == "" {
		log.Println("empty refresh token")
		response.Error(w, apperror.Unauthorized("No refresh token provided"))
		return
	}

	access, refresh, err := h.AuthService.RefreshSession(r.Context(), refreshToken)
	if err != nil {
		log.Println(err)
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"access":  access,
		"refresh": refresh,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())

	user, err := h.UserService.GetUser(r.Context(), claims.UserID)
	if err != nil {
		fmt.Println(err)
		response.Error(w, apperror.Internal())
		return
	}
	if user == nil {
		response.Error(w, apperror.Unauthorized("user not found"))
		return
	}

	response.JSON(w, http.StatusOK, user)
	return
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := h.getTokenFromHeader(r)
	if err != nil {
		response.Error(w, err)
		return
	}

	if refreshToken != "" {
		_ = h.AuthService.Logout(r.Context(), refreshToken) // best-effort, ignore not-found
	}

	response.OK(w)
}

// helper func to get cookie from Authorization header
func (h *AuthHandler) getTokenFromHeader(r *http.Request) (string, error) {
	BearerTkn := strings.Split(r.Header.Get("Authorization"), " ")
	if len(BearerTkn) != 2 {
		return "", apperror.Unauthorized("Invalid token")

	}
	refreshToken := BearerTkn[1]
	return refreshToken, nil

}

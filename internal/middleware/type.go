package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id,omitempty"`
	Email  string    `json:"email"`
	Type   string    `json:"typ"` // user type
	jwt.RegisteredClaims
}

type TokenValidator interface {
	ValidateAccessToken(tokenString string) (*Claims, error)
}


package middleware

import (
	"context"
	"errors"
	"golang-api-template/pkg/apperror"
	"golang-api-template/pkg/response"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func AuthMiddleware(tokenValidator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r)
			if err != nil {
				response.Error(w, apperror.Unauthorized(err.Error()))
				return
			}

			claims, err := tokenValidator.ValidateAccessToken(token)
			if err != nil {
				response.Error(w, apperror.Unauthorized("invalid or expired token"))
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Only(allowedTypes ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool)
	for _, t := range allowedTypes {
		allowed[t] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := getClaimsFromContext(r.Context())
			if !ok {
				response.Error(w, apperror.Forbidden())
				return
			}

			if claims.UserID == uuid.Nil {
				response.Error(w, apperror.Forbidden())
				return
			}

			if len(allowed) > 0 && !allowed[claims.Type] {
				response.Error(w, apperror.Forbidden())
				return
			}

			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing auth header")
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid token format")
	}

	return parts[1], nil
}

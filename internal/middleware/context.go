package middleware

import "context"

type contextKey string

const ClaimsContextKey contextKey = "claims"

func getClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}

// this is for usage inside handlers
func GetClaimsFromContext(ctx context.Context) *Claims {
	return ctx.Value(ClaimsContextKey).(*Claims)
}


package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

type contextKey string

const (
	ContextKeyUserID contextKey = "userID"
	ContextKeyRole   contextKey = "role"
)

// Authenticate validates the Bearer access token and stores user id and role in context.
func Authenticate(tokens *security.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				response.Error(w, http.StatusUnauthorized, "Authentication required", errors.CodeUnauthorized)
				return
			}

			claims, err := tokens.ParseAccessToken(token)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Invalid or expired token", errors.CodeInvalidToken)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole allows the request only when the user has one of the given roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(ContextKeyRole).(string)
			if !ok || role == "" {
				response.Error(w, http.StatusUnauthorized, "Authentication required", errors.CodeUnauthorized)
				return
			}

			if _, ok := allowed[role]; !ok {
				response.Error(w, http.StatusForbidden, "Insufficient permissions", errors.CodeForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserIDFromContext returns the authenticated user id.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return id, ok
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

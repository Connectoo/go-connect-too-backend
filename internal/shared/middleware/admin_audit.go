package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AuditStore persists admin audit log entries.
type AuditStore interface {
	InsertAuditLog(ctx context.Context, entry AdminAuditEntry) error
}

// AdminAuditEntry is a single admin action log row.
type AdminAuditEntry struct {
	ID           uuid.UUID
	AdminUserID  uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Details      json.RawMessage
	CreatedAt    time.Time
}

// AdminAudit logs mutating admin requests.
func AdminAudit(store AuditStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if store == nil || !isAuditableAdminRequest(r) {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)

			resourceID := parseResourceID(chi.URLParam(r, "id"))
			details, _ := json.Marshal(map[string]string{
				"method": r.Method,
				"path":   r.URL.Path,
			})
			_ = store.InsertAuditLog(r.Context(), AdminAuditEntry{
				ID:           uuid.New(),
				AdminUserID:  userID,
				Action:       r.Method + " " + r.URL.Path,
				ResourceType: auditResourceType(r.URL.Path),
				ResourceID:   resourceID,
				Details:      details,
				CreatedAt:    time.Now().UTC(),
			})
		})
	}
}

func isAuditableAdminRequest(r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/api/v1/admin/") {
		return false
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func auditResourceType(path string) string {
	trimmed := strings.TrimPrefix(path, "/api/v1/admin/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "admin"
	}
	return parts[0]
}

func parseResourceID(value string) *uuid.UUID {
	if value == "" {
		return nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil
	}
	return &id
}

package settings

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts admin settings endpoints.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager, auditStore middleware.AuditStore) {
	r.Route("/admin/settings", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))
		r.Use(middleware.AdminAudit(auditStore))

		r.Get("/", h.getSettings)
		r.Put("/", h.updateSettings)
	})
}

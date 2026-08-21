package admin

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts admin dashboard and user management endpoints.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager, auditStore middleware.AuditStore) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))
		r.Use(middleware.AdminAudit(auditStore))

		r.Get("/dashboard/summary", h.dashboardSummary)

		r.Route("/users", func(r chi.Router) {
			r.Get("/", h.listUsers)
			r.Get("/{id}", h.getUser)
			r.Patch("/{id}/suspend", h.suspendUser)
			r.Patch("/{id}/activate", h.activateUser)
		})
	})
}

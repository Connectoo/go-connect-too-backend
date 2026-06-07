package payments

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager, auditStore middleware.AuditStore) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleEmployee))
		r.Get("/employee/payments", h.listEmployee)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))
		r.Use(middleware.AdminAudit(auditStore))
		r.Get("/admin/payments", h.listAdmin)
		r.Post("/admin/payments/{id}/refund", h.refund)
	})
}

package support

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager, auditStore middleware.AuditStore) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleCustomer))

		r.Post("/support/tickets", h.create)
		r.Get("/support/tickets", h.listCustomer)
	})

	r.Route("/admin/support/tickets", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))
		r.Use(middleware.AdminAudit(auditStore))

		r.Get("/", h.listAdmin)
		r.Get("/{id}", h.getAdmin)
		r.Patch("/{id}", h.updateAdmin)
		r.Post("/{id}/messages", h.addMessageAdmin)
	})
}

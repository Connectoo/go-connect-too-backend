package bookings

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts booking endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager, auditStore middleware.AuditStore) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleCustomer))

		r.Post("/bookings", h.create)
		r.Post("/bookings/rebook", h.rebook)
		r.Get("/bookings", h.listCustomer)
		r.Get("/bookings/{id}/rebook-preview", h.rebookPreview)
		r.Get("/bookings/{id}", h.getCustomer)
		r.Patch("/bookings/{id}/cancel", h.cancel)
		r.Patch("/bookings/{id}/reschedule", h.rescheduleCustomer)
	})

	r.Route("/employee/bookings", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleEmployee))

		r.Get("/", h.listEmployee)
		r.Patch("/{id}/accept", h.accept)
		r.Patch("/{id}/reject", h.reject)
		r.Patch("/{id}/start", h.start)
		r.Patch("/{id}/complete", h.complete)
		r.Patch("/{id}/reschedule", h.rescheduleEmployee)
		r.Patch("/{id}/cancel", h.cancelEmployee)
		r.Patch("/{id}/no-show", h.noShowEmployee)
	})

	r.Route("/admin/bookings", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))
		r.Use(middleware.AdminAudit(auditStore))

		r.Get("/", h.listAdmin)
		r.Get("/{id}", h.getAdmin)
		r.Patch("/{id}/status", h.updateStatusAdmin)
	})
}

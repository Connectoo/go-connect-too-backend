package services

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts employee service endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Get("/services", h.listPublic)
	r.Get("/services/{id}", h.getPublic)
	r.Get("/employees/{id}/services", h.listPublicByEmployee)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleEmployee))

		r.Post("/employee/services", h.create)
		r.Get("/employee/services", h.list)
		r.Put("/employee/services/{id}", h.update)
		r.Delete("/employee/services/{id}", h.delete)
		r.Patch("/employee/services/{id}/status", h.updateStatus)
	})

	r.Route("/admin/services", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))

		r.Get("/", h.listAdmin)
		r.Patch("/{id}/activate", h.activateAdmin)
		r.Patch("/{id}/deactivate", h.deactivateAdmin)
	})
}

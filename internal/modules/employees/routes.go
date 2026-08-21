package employees

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts employee profile endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Get("/employees/{id}/public-profile", h.getPublicProfile)

	r.Route("/employee", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleEmployee))

		r.Get("/profile", h.getProfile)
		r.Put("/profile", h.updateProfile)
	})

	r.Route("/admin/employees", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))

		r.Get("/", h.listAdmin)
		r.Get("/{id}", h.getAdmin)
		r.Patch("/{id}/approve", h.approveProfile)
		r.Patch("/{id}/reject", h.rejectProfile)
		r.Patch("/{id}/suspend", h.suspendProfile)
	})
}

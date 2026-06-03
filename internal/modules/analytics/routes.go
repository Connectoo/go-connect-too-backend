package analytics

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts employee and admin analytics endpoints.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Route("/employee/analytics", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleEmployee))

		r.Get("/summary", h.employeeSummary)
		r.Get("/bookings", h.employeeBookings)
		r.Get("/reviews", h.employeeReviews)
	})

	r.Route("/admin/analytics", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))

		r.Get("/summary", h.adminSummary)
		r.Get("/revenue", h.adminRevenue)
		r.Get("/bookings", h.adminBookings)
		r.Get("/categories", h.adminCategories)
	})
}

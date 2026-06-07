package moderation

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts admin moderation endpoints.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Route("/admin/reviews", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))

		r.Get("/", h.listReviews)
		r.Patch("/{id}/approve", h.approveReview)
		r.Patch("/{id}/hide", h.hideReview)
	})

	r.Route("/admin/reports", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))

		r.Get("/", h.listReports)
		r.Patch("/{id}/resolve", h.resolveReport)
	})
}

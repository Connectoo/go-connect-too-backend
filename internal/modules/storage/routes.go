package storage

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts storage endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Route("/uploads", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))

		r.Post("/presign", h.presign)
		r.Delete("/{id}", h.delete)
	})
}

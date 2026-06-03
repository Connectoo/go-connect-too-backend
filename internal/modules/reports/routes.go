package reports

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts authenticated report submission endpoints.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Post("/reports", h.create)
	})
}

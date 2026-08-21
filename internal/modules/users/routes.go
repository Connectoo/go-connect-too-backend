package users

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts user profile endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Route("/users", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))

		r.Get("/me", h.getMe)
		r.Put("/me", h.updateMe)
		r.Patch("/me/deactivate", h.deactivateMe)

		r.Route("/addresses", func(r chi.Router) {
			r.Get("/", h.listAddresses)
			r.Post("/", h.createAddress)
			r.Put("/{id}", h.updateAddress)
			r.Delete("/{id}", h.deleteAddress)
		})
	})
}

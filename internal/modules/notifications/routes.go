package notifications

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts notification endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))

		r.Get("/notifications", h.list)
		r.Patch("/notifications/{id}/read", h.markRead)
		r.Patch("/notifications/read-all", h.markAllRead)
		r.Post("/device-tokens", h.registerDeviceToken)
	})
}

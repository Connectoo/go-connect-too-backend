package chat

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts chat endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Route("/chat", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))

		r.Get("/conversations", h.listConversations)
		r.Get("/conversations/{id}/messages", h.listMessages)
		r.Post("/conversations/{id}/messages", h.sendMessage)
		r.Patch("/conversations/{id}/messages/{messageId}/read", h.markMessageRead)
	})
}

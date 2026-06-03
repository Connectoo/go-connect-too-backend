package websocket

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts the WebSocket endpoint.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/ws", h.Serve)
}

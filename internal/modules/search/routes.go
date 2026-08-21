package search

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts search endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/search", func(r chi.Router) {
		r.Get("/services", h.searchServices)
		r.Get("/employees", h.searchEmployees)
	})
}

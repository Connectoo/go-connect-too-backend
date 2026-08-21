package public

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts public website endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/public", func(r chi.Router) {
		r.Get("/home", h.home)
		r.Get("/categories", h.listCategories)
		r.Get("/providers", h.listProviders)
		r.Get("/providers/{id}", h.getProvider)
		r.Get("/services", h.listServices)
		r.Get("/services/{id}", h.getService)
	})
}

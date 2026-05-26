package webhooks

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) { r.Post("/webhooks/razorpay", h.razorpay) }

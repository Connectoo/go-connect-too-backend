package subscriptions

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Get("/subscription-plans", h.listPlans)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleEmployee))
		r.Post("/employee/subscriptions/create-order", h.createOrder)
		r.Post("/employee/subscriptions/verify-payment", h.verifyPayment)
		r.Post("/employee/subscriptions/cancel", h.cancel)
		r.Post("/employee/subscriptions/change-plan", h.changePlan)
		r.Patch("/employee/subscriptions/auto-renew", h.setAutoRenew)
		r.Get("/employee/subscriptions/current", h.current)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))
		r.Post("/admin/subscription-plans", h.createPlan)
		r.Put("/admin/subscription-plans/{id}", h.updatePlan)
		r.Get("/admin/subscriptions", h.listSubscriptions)
	})
}

package auth

import (
	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// RegisterRoutes mounts auth endpoints on the router.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register/customer", h.registerCustomer)
		r.Post("/register/employee", h.registerEmployee)
		r.Post("/login/customer", h.loginCustomer)
		r.Post("/login/employee", h.loginEmployee)
		r.Post("/login/admin", h.loginAdmin)
		r.Post("/refresh", h.refresh)
		r.Post("/logout", h.logout)
		r.Post("/forgot-password", h.forgotPassword)
		r.Post("/reset-password", h.resetPassword)
		r.Post("/verify-email", h.verifyEmail)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(tokens))
			r.Get("/me", h.me)
			r.Post("/resend-verification", h.resendVerification)
			r.Post("/change-password", h.changePassword)
		})
	})
}

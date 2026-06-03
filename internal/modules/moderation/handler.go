package moderation

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// Handler serves moderation placeholder endpoints.
type Handler struct {
	log *slog.Logger
}

// NewHandler creates a moderation handler.
func NewHandler(log *slog.Logger) *Handler {
	return &Handler{log: log}
}

func (h *Handler) status(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, "Moderation module placeholder", map[string]string{
		"status":  "not_implemented",
		"message": "Moderation workflows will be added in a later phase",
	})
}

// RegisterRoutes mounts moderation placeholder endpoints.
func RegisterRoutes(r chi.Router, h *Handler, tokens *security.TokenManager) {
	r.Route("/admin/moderation", func(r chi.Router) {
		r.Use(middleware.Authenticate(tokens))
		r.Use(middleware.RequireRole(users.RoleAdmin))
		r.Get("/", h.status)
	})
}

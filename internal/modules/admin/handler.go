package admin

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves admin HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates an admin handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) dashboardSummary(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.DashboardSummary(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Dashboard summary loaded", res)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	page := pagination.Parse(r.URL.Query())
	filter := UserListFilter{
		Role:   r.URL.Query().Get("role"),
		Status: r.URL.Query().Get("status"),
		Query:  r.URL.Query().Get("q"),
		Page:   page,
	}

	res, err := h.svc.ListUsers(r.Context(), filter)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Users loaded", res)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "User loaded", res)
}

func (h *Handler) suspendUser(w http.ResponseWriter, r *http.Request) {
	h.updateUserStatus(w, r, h.svc.SuspendUser, "User suspended")
}

func (h *Handler) activateUser(w http.ResponseWriter, r *http.Request) {
	h.updateUserStatus(w, r, h.svc.ActivateUser, "User activated")
}

type userStatusFunc func(context.Context, uuid.UUID) (*UserResponse, error)

func (h *Handler) updateUserStatus(w http.ResponseWriter, r *http.Request, action userStatusFunc, message string) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user id", sharederrors.CodeValidationError)
		return
	}

	res, err := action(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, message, res)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		response.Error(w, http.StatusBadRequest, "Validation failed", sharederrors.CodeValidationError)
	case errors.Is(err, ErrNotFound), errors.Is(err, users.ErrNotFound):
		response.Error(w, http.StatusNotFound, "User not found", sharederrors.CodeNotFound)
	default:
		if h.log != nil {
			h.log.Error("admin request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

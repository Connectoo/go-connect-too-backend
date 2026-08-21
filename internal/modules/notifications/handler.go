package notifications

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves notification HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a notification handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	page := pagination.Parse(r.URL.Query())
	items, err := h.svc.List(r.Context(), userID, page)
	if err != nil {
		h.log.Error("list notifications failed", slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "Failed to load notifications", sharederrors.CodeInternalError)
		return
	}

	response.JSON(w, http.StatusOK, "Notifications loaded", items)
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	notificationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid notification id", sharederrors.CodeValidationError)
		return
	}

	item, err := h.svc.MarkRead(r.Context(), userID, notificationID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Notification marked as read", item)
}

func (h *Handler) markAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	count, err := h.svc.MarkAllRead(r.Context(), userID)
	if err != nil {
		h.log.Error("mark all notifications read failed", slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "Failed to mark notifications as read", sharederrors.CodeInternalError)
		return
	}

	response.JSON(w, http.StatusOK, "Notifications marked as read", map[string]int64{"updated": count})
}

func (h *Handler) registerDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	var req RegisterDeviceTokenRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	item, err := h.svc.RegisterDeviceToken(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "Device token registered", item)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "Notification not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrForbidden):
		response.Error(w, http.StatusForbidden, "Insufficient permissions", sharederrors.CodeForbidden)
	case errors.Is(err, ErrValidation):
		response.Error(w, http.StatusBadRequest, err.Error(), sharederrors.CodeValidationError)
	default:
		h.log.Error("notification request failed", slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "Request failed", sharederrors.CodeInternalError)
	}
}

func decodeJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

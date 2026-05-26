package users

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves user profile HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a users handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := h.authContext(w, r)
	if !ok {
		return
	}

	res, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Profile loaded", res)
}

func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := h.authContext(w, r)
	if !ok {
		return
	}

	var req UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.UpdateProfile(r.Context(), userID, role, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Profile updated", res)
}

func (h *Handler) listAddresses(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := h.authContext(w, r)
	if !ok {
		return
	}

	items, err := h.svc.ListAddresses(r.Context(), userID, role)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Addresses loaded", items)
}

func (h *Handler) createAddress(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := h.authContext(w, r)
	if !ok {
		return
	}

	var req CreateAddressRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.CreateAddress(r.Context(), userID, role, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "Address created", res)
}

func (h *Handler) updateAddress(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := h.authContext(w, r)
	if !ok {
		return
	}

	addressID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid address id", sharederrors.CodeValidationError)
		return
	}

	var req UpdateAddressRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.UpdateAddress(r.Context(), userID, role, addressID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Address updated", res)
}

func (h *Handler) deleteAddress(w http.ResponseWriter, r *http.Request) {
	userID, role, ok := h.authContext(w, r)
	if !ok {
		return
	}

	addressID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid address id", sharederrors.CodeValidationError)
		return
	}

	if err := h.svc.DeleteAddress(r.Context(), userID, role, addressID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Address deleted", map[string]string{"id": addressID.String()})
}

func (h *Handler) authContext(w http.ResponseWriter, r *http.Request) (uuid.UUID, string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return uuid.Nil, "", false
	}
	role, _ := r.Context().Value(middleware.ContextKeyRole).(string)
	return userID, role, true
}

func decodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		response.Error(w, http.StatusBadRequest, "Validation failed", sharederrors.CodeValidationError)
	case errors.Is(err, ErrForbiddenProfile):
		response.Error(w, http.StatusForbidden, "Insufficient permissions", sharederrors.CodeForbidden)
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "User not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrAddressNotFound):
		response.Error(w, http.StatusNotFound, "Address not found", sharederrors.CodeNotFound)
	default:
		if h.log != nil {
			h.log.Error("users request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

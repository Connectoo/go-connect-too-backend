package support

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

type Handler struct {
	svc *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	var req CreateTicketRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, "Support ticket created", res)
}

func (h *Handler) listCustomer(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	items, err := h.svc.ListForCustomer(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Support tickets loaded", items)
}

func (h *Handler) listAdmin(w http.ResponseWriter, r *http.Request) {
	page := pagination.Parse(r.URL.Query())
	res, err := h.svc.ListForAdmin(r.Context(), r.URL.Query().Get("status"), page)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Support tickets loaded", res)
}

func (h *Handler) getAdmin(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ticket id", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.GetForAdmin(r.Context(), ticketID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Support ticket loaded", res)
}

func (h *Handler) updateAdmin(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ticket id", sharederrors.CodeValidationError)
		return
	}
	var req UpdateTicketRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.UpdateForAdmin(r.Context(), ticketID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Support ticket updated", res)
}

func (h *Handler) addMessageAdmin(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ticket id", sharederrors.CodeValidationError)
		return
	}
	var req AddMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.AddMessageForAdmin(r.Context(), adminUserID, ticketID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, "Support message added", res)
}

func decodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		response.Error(w, http.StatusBadRequest, "Validation failed", sharederrors.CodeValidationError)
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrCustomerNotFound):
		response.Error(w, http.StatusNotFound, "Resource not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrForbidden):
		response.Error(w, http.StatusForbidden, "Insufficient permissions", sharederrors.CodeForbidden)
	default:
		if h.log != nil {
			h.log.Error("support request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

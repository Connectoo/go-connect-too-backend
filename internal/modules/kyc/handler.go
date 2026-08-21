package kyc

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves employee KYC HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a KYC handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	var req SubmitRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Submit(r.Context(), userID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "KYC submitted", res)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	res, err := h.svc.Get(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "KYC loaded", res)
}

func (h *Handler) listAdmin(w http.ResponseWriter, r *http.Request) {
	page := pagination.Parse(r.URL.Query())
	res, err := h.svc.ListForAdmin(r.Context(), r.URL.Query().Get("status"), page)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "KYC records loaded", res)
}

func (h *Handler) getAdmin(w http.ResponseWriter, r *http.Request) {
	kycID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid KYC id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.GetForAdmin(r.Context(), kycID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "KYC record loaded", res)
}

func (h *Handler) approveAdmin(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	kycID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid KYC id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Approve(r.Context(), reviewerID, kycID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "KYC approved", res)
}

func (h *Handler) rejectAdmin(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	kycID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid KYC id", sharederrors.CodeValidationError)
		return
	}

	var req RejectRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Reject(r.Context(), reviewerID, kycID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "KYC rejected", res)
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
	case errors.Is(err, employees.ErrNotFound):
		response.Error(w, http.StatusNotFound, "Employee profile not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "KYC record not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrAlreadyExists):
		response.Error(w, http.StatusConflict, "KYC already submitted", sharederrors.CodeConflict)
	case errors.Is(err, ErrInvalidStatus):
		response.Error(w, http.StatusConflict, "KYC is not pending review", sharederrors.CodeConflict)
	case errors.Is(err, ErrFileNotFound), errors.Is(err, ErrFileNotOwned):
		response.Error(w, http.StatusBadRequest, "Invalid uploaded file", sharederrors.CodeValidationError)
	default:
		if h.log != nil {
			h.log.Error("kyc request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

package payments

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

type Handler struct {
	svc *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler { return &Handler{svc: svc, log: log} }

func (h *Handler) listEmployee(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	items, err := h.svc.ListForEmployeeUser(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Payments loaded", items)
}

func (h *Handler) refund(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	paymentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid payment id", sharederrors.CodeValidationError)
		return
	}

	var req RefundRequest
	if r.ContentLength > 0 {
		if err := decodeJSON(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
			return
		}
	}

	res, err := h.svc.Refund(r.Context(), adminUserID, paymentID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, "Refund created", res)
}

func (h *Handler) listAdmin(w http.ResponseWriter, r *http.Request) {
	page := pagination.Parse(r.URL.Query())
	res, err := h.svc.ListAll(r.Context(), r.URL.Query().Get("status"), page)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Payments loaded", res)
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
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "Payment not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrPaymentNotRefundable), errors.Is(err, ErrPaymentNotPending):
		response.Error(w, http.StatusConflict, "Payment cannot be refunded", sharederrors.CodeConflict)
	case errors.Is(err, employees.ErrNotFound):
		response.Error(w, http.StatusNotFound, "Employee profile not found", sharederrors.CodeNotFound)
	default:
		if h.log != nil {
			h.log.Error("payment request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

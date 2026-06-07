package subscriptions

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/payments"
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

func (h *Handler) listPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.svc.ListPlans(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Subscription plans loaded", plans)
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	var req CreateOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.CreateOrder(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, "Subscription order created", res)
}

func (h *Handler) current(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	res, err := h.svc.Current(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Current subscription loaded", res)
}

func (h *Handler) createPlan(w http.ResponseWriter, r *http.Request) {
	var req CreatePlanRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.CreatePlan(r.Context(), req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, "Subscription plan created", res)
}

func (h *Handler) updatePlan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid plan id", sharederrors.CodeValidationError)
		return
	}
	var req UpdatePlanRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.UpdatePlan(r.Context(), id, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Subscription plan updated", res)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	var req CancelSubscriptionRequest
	if r.ContentLength > 0 {
		if err := decodeJSON(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
			return
		}
	}
	res, err := h.svc.Cancel(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Subscription cancelled", res)
}

func (h *Handler) changePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	var req ChangePlanRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.ChangePlan(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, "Plan change initiated", res)
}

func (h *Handler) setAutoRenew(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	var req AutoRenewRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.SetAutoRenew(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Auto-renew updated", res)
}

func (h *Handler) verifyPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}
	var req VerifyPaymentRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}
	res, err := h.svc.VerifyPayment(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Payment verified", res)
}

func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	page := pagination.Parse(r.URL.Query())
	res, err := h.svc.ListSubscriptions(r.Context(), r.URL.Query().Get("status"), page)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Subscriptions loaded", res)
}

func decodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation), errors.Is(err, payments.ErrValidation):
		response.Error(w, http.StatusBadRequest, "Validation failed", sharederrors.CodeValidationError)
	case errors.Is(err, ErrNotFound), errors.Is(err, employees.ErrNotFound):
		response.Error(w, http.StatusNotFound, "Subscription resource not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrPlanInactive):
		response.Error(w, http.StatusConflict, "Subscription plan is inactive", sharederrors.CodeConflict)
	case errors.Is(err, ErrNoActive), errors.Is(err, ErrAlreadyCancelled):
		response.Error(w, http.StatusConflict, "Subscription cannot be changed", sharederrors.CodeConflict)
	case errors.Is(err, ErrInvalidSignature):
		response.Error(w, http.StatusBadRequest, "Invalid payment signature", sharederrors.CodeValidationError)
	case errors.Is(err, ErrPaymentNotPending), errors.Is(err, payments.ErrPaymentNotPending):
		response.Error(w, http.StatusConflict, "Payment is not pending", sharederrors.CodeConflict)
	case errors.Is(err, payments.ErrGatewayNotConfigured):
		response.Error(w, http.StatusConflict, "Payment gateway is not configured", sharederrors.CodeConflict)
	default:
		if h.log != nil {
			h.log.Error("subscription request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

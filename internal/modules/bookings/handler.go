package bookings

import (
	"context"
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

// Handler serves booking HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a booking handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) rebookPreview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid booking id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.RebookPreview(r.Context(), userID, bookingID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Rebook preview loaded", res)
}

func (h *Handler) rebook(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	var req RebookBookingRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Rebook(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "Booking rebooked", res)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	var req CreateBookingRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "Booking created", res)
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

	response.JSON(w, http.StatusOK, "Bookings loaded", items)
}

func (h *Handler) getCustomer(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid booking id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.GetForCustomer(r.Context(), userID, bookingID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Booking loaded", res)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid booking id", sharederrors.CodeValidationError)
		return
	}

	var req CancelBookingRequest
	if r.ContentLength > 0 {
		if err := decodeJSON(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
			return
		}
	}

	res, err := h.svc.Cancel(r.Context(), userID, bookingID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Booking cancelled", res)
}

func (h *Handler) listEmployee(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	items, err := h.svc.ListForEmployee(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Bookings loaded", items)
}

func (h *Handler) accept(w http.ResponseWriter, r *http.Request) {
	h.employeeAction(w, r, func(ctx context.Context, userID, bookingID uuid.UUID, req EmployeeActionRequest) (*BookingResponse, error) {
		return h.svc.Accept(ctx, userID, bookingID, req)
	}, "Booking accepted")
}

func (h *Handler) reject(w http.ResponseWriter, r *http.Request) {
	h.employeeAction(w, r, func(ctx context.Context, userID, bookingID uuid.UUID, req EmployeeActionRequest) (*BookingResponse, error) {
		return h.svc.Reject(ctx, userID, bookingID, req)
	}, "Booking rejected")
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	h.employeeAction(w, r, func(ctx context.Context, userID, bookingID uuid.UUID, req EmployeeActionRequest) (*BookingResponse, error) {
		return h.svc.Start(ctx, userID, bookingID, req)
	}, "Booking started")
}

func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
	h.employeeAction(w, r, func(ctx context.Context, userID, bookingID uuid.UUID, req EmployeeActionRequest) (*BookingResponse, error) {
		return h.svc.Complete(ctx, userID, bookingID, req)
	}, "Booking completed")
}

func (h *Handler) listAdmin(w http.ResponseWriter, r *http.Request) {
	page := pagination.Parse(r.URL.Query())
	res, err := h.svc.ListForAdmin(r.Context(), r.URL.Query().Get("status"), page)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Bookings loaded", res)
}

func (h *Handler) getAdmin(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid booking id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.GetForAdmin(r.Context(), bookingID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Booking loaded", res)
}

type employeeActionFunc func(context.Context, uuid.UUID, uuid.UUID, EmployeeActionRequest) (*BookingResponse, error)

func (h *Handler) employeeAction(w http.ResponseWriter, r *http.Request, action employeeActionFunc, message string) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid booking id", sharederrors.CodeValidationError)
		return
	}

	var req EmployeeActionRequest
	if r.ContentLength > 0 {
		if err := decodeJSON(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
			return
		}
	}

	res, err := action(r.Context(), userID, bookingID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, message, res)
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
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrServiceNotFound), errors.Is(err, ErrCustomerProfileNotFound):
		response.Error(w, http.StatusNotFound, "Resource not found", sharederrors.CodeNotFound)
	case errors.Is(err, employees.ErrNotFound):
		response.Error(w, http.StatusNotFound, "Employee profile not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrForbidden):
		response.Error(w, http.StatusForbidden, "Insufficient permissions", sharederrors.CodeForbidden)
	case errors.Is(err, ErrDoubleBooking), errors.Is(err, ErrEmployeeUnavailable), errors.Is(err, ErrEmployeeNotApproved), errors.Is(err, ErrRebookNotAllowed):
		response.Error(w, http.StatusConflict, "Booking cannot be created", sharederrors.CodeConflict)
	case errors.Is(err, ErrInvalidStatusTransition):
		response.Error(w, http.StatusConflict, "Invalid booking status transition", sharederrors.CodeConflict)
	default:
		if h.log != nil {
			h.log.Error("booking request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

package analytics

import (
	"errors"
	"log/slog"
	"net/http"

	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves analytics HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates an analytics handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) employeeSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	res, err := h.svc.EmployeeSummary(r.Context(), userID, r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Employee analytics summary loaded", res)
}

func (h *Handler) employeeBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	res, err := h.svc.EmployeeBookings(r.Context(), userID, r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Employee booking analytics loaded", res)
}

func (h *Handler) employeeReviews(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	res, err := h.svc.EmployeeReviews(r.Context(), userID, r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Employee review analytics loaded", res)
}

func (h *Handler) adminSummary(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.AdminSummary(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Admin analytics summary loaded", res)
}

func (h *Handler) adminRevenue(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.AdminRevenue(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Admin revenue analytics loaded", res)
}

func (h *Handler) adminBookings(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.AdminBookings(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Admin booking analytics loaded", res)
}

func (h *Handler) adminCategories(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.AdminCategories(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Admin category analytics loaded", res)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidDateRange):
		response.Error(w, http.StatusBadRequest, "Invalid date range", sharederrors.CodeValidationError)
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "Employee profile not found", sharederrors.CodeNotFound)
	default:
		if h.log != nil {
			h.log.Error("analytics request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

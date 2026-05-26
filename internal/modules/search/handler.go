package search

import (
	"errors"
	"log/slog"
	"net/http"

	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves search HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a search handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) searchServices(w http.ResponseWriter, r *http.Request) {
	params, err := ParseServiceSearchParams(r.URL.Query())
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid search parameters", sharederrors.CodeValidationError)
		return
	}

	items, err := h.svc.SearchServices(r.Context(), params)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Service search results", items)
}

func (h *Handler) searchEmployees(w http.ResponseWriter, r *http.Request) {
	params, err := ParseEmployeeSearchParams(r.URL.Query())
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid search parameters", sharederrors.CodeValidationError)
		return
	}

	items, err := h.svc.SearchEmployees(r.Context(), params)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Employee search results", items)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		response.Error(w, http.StatusBadRequest, "Validation failed", sharederrors.CodeValidationError)
	default:
		if h.log != nil {
			h.log.Error("search request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

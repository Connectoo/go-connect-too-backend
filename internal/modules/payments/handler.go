package payments

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
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

func (h *Handler) listAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAll(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Payments loaded", items)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, employees.ErrNotFound):
		response.Error(w, http.StatusNotFound, "Employee profile not found", sharederrors.CodeNotFound)
	default:
		if h.log != nil {
			h.log.Error("payment request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

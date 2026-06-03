package settings

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves settings HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a settings handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.GetSettings(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Settings loaded", res)
}

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var req UpdateSettingsRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.UpdateSettings(r.Context(), req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Settings updated", res)
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
	default:
		if h.log != nil {
			h.log.Error("settings request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

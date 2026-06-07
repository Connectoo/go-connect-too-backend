package storage

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

// Handler serves storage HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a storage handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) presign(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	var req PresignRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.PresignUpload(r.Context(), userID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Upload URL created", res)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	fileID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid file id", sharederrors.CodeValidationError)
		return
	}

	if err := h.svc.Delete(r.Context(), userID, fileID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "File deleted", map[string]string{"id": fileID.String()})
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
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "File not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrForbidden):
		response.Error(w, http.StatusForbidden, "Insufficient permissions", sharederrors.CodeForbidden)
	case errors.Is(err, ErrUnavailable):
		response.Error(w, http.StatusServiceUnavailable, "Storage is not configured", sharederrors.CodeInternalError)
	default:
		if h.log != nil {
			h.log.Error("storage request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

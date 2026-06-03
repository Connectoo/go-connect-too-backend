package categories

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves category HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a category handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListActive(r.Context())
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Categories loaded", items)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.CreateCategory(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "Category created", res)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid category id", sharederrors.CodeValidationError)
		return
	}

	var req UpdateCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.UpdateCategory(r.Context(), id, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Category updated", res)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid category id", sharederrors.CodeValidationError)
		return
	}

	if err := h.svc.DeleteCategory(r.Context(), id); err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Category deleted", map[string]string{"id": id.String()})
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
	case errors.Is(err, ErrDuplicateName):
		response.Error(w, http.StatusConflict, "Category name already exists", sharederrors.CodeConflict)
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "Category not found", sharederrors.CodeNotFound)
	default:
		if h.log != nil {
			h.log.Error("category request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

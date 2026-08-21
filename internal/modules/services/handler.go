package services

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves employee service HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates an employee service handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) listPublic(w http.ResponseWriter, r *http.Request) {
	var categoryID *uuid.UUID
	if raw := r.URL.Query().Get("category_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid category_id", sharederrors.CodeValidationError)
			return
		}
		categoryID = &id
	}

	items, err := h.svc.ListPublic(r.Context(), categoryID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Services loaded", items)
}

func (h *Handler) listPublicByEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid employee id", sharederrors.CodeValidationError)
		return
	}

	items, err := h.svc.ListPublicByEmployee(r.Context(), employeeID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Employee services loaded", items)
}

func (h *Handler) getPublic(w http.ResponseWriter, r *http.Request) {
	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid service id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.GetPublic(r.Context(), serviceID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Service loaded", res)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	var req CreateServiceRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "Service created", res)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	res, err := h.svc.List(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Services loaded", res)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid service id", sharederrors.CodeValidationError)
		return
	}

	var req UpdateServiceRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Update(r.Context(), userID, serviceID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Service updated", res)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid service id", sharederrors.CodeValidationError)
		return
	}

	if err := h.svc.Delete(r.Context(), userID, serviceID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Service deleted", map[string]string{"id": serviceID.String()})
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid service id", sharederrors.CodeValidationError)
		return
	}

	var req UpdateServiceStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.UpdateStatus(r.Context(), userID, serviceID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Service status updated", res)
}

func (h *Handler) listAdmin(w http.ResponseWriter, r *http.Request) {
	page := pagination.Parse(r.URL.Query())
	query := r.URL.Query()
	res, err := h.svc.ListForAdmin(
		r.Context(),
		ParseOptionalUUID(query.Get("category_id")),
		ParseOptionalBool(query.Get("is_active")),
		query.Get("q"),
		page,
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Services loaded", res)
}

func (h *Handler) activateAdmin(w http.ResponseWriter, r *http.Request) {
	h.adminSetStatus(w, r, true, "Service activated")
}

func (h *Handler) deactivateAdmin(w http.ResponseWriter, r *http.Request) {
	h.adminSetStatus(w, r, false, "Service deactivated")
}

func (h *Handler) adminSetStatus(w http.ResponseWriter, r *http.Request, active bool, message string) {
	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid service id", sharederrors.CodeValidationError)
		return
	}

	var res *ServiceResponse
	if active {
		res, err = h.svc.AdminActivate(r.Context(), serviceID)
	} else {
		res, err = h.svc.AdminDeactivate(r.Context(), serviceID)
	}
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, message, res)
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
	case errors.Is(err, ErrCategoryNotFound):
		response.Error(w, http.StatusNotFound, "Category not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "Service not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrProfileIncomplete):
		response.Error(w, http.StatusConflict, "Complete employee profile before activating services", sharederrors.CodeConflict)
	case errors.Is(err, ErrServiceLimit):
		response.Error(w, http.StatusConflict, "Subscription service limit reached", sharederrors.CodeConflict)
	default:
		if h.log != nil {
			h.log.Error("employee service request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

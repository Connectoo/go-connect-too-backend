package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

// Handler serves auth HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates an auth handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) registerCustomer(w http.ResponseWriter, r *http.Request) {
	var req RegisterCustomerRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.RegisterCustomer(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "Customer registered", res)
}

func (h *Handler) registerEmployee(w http.ResponseWriter, r *http.Request) {
	var req RegisterEmployeeRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.RegisterEmployee(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "Employee registered", res)
}

func (h *Handler) loginCustomer(w http.ResponseWriter, r *http.Request) {
	h.login(w, r, "Customer", h.svc.LoginCustomer)
}

func (h *Handler) loginEmployee(w http.ResponseWriter, r *http.Request) {
	h.login(w, r, "Employee", h.svc.LoginEmployee)
}

func (h *Handler) loginAdmin(w http.ResponseWriter, r *http.Request) {
	h.login(w, r, "Admin", h.svc.LoginAdmin)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request, label string, fn func(context.Context, LoginRequest) (*AuthResponse, error)) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := fn(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, label+" login successful", res)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Token refreshed", res)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Logged out", nil)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	res, err := h.svc.Me(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Profile loaded", res)
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
	case errors.Is(err, users.ErrDuplicateEmail):
		response.Error(w, http.StatusConflict, "Email already registered", sharederrors.CodeConflict)
	case errors.Is(err, users.ErrDuplicatePhone):
		response.Error(w, http.StatusConflict, "Phone already registered", sharederrors.CodeConflict)
	case errors.Is(err, ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "Invalid email or password", sharederrors.CodeInvalidCredentials)
	case errors.Is(err, ErrInvalidToken):
		response.Error(w, http.StatusUnauthorized, "Invalid or expired token", sharederrors.CodeInvalidToken)
	case errors.Is(err, ErrUserInactive):
		response.Error(w, http.StatusForbidden, "Account is not active", sharederrors.CodeForbidden)
	case errors.Is(err, users.ErrNotFound):
		response.Error(w, http.StatusNotFound, "User not found", sharederrors.CodeUnauthorized)
	default:
		if h.log != nil {
			h.log.Error("auth request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

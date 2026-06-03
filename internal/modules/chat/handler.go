package chat

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

// Handler serves chat HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// NewHandler creates a chat handler.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) listConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	items, err := h.svc.ListConversations(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Conversations loaded", items)
}

func (h *Handler) listMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	conversationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid conversation id", sharederrors.CodeValidationError)
		return
	}

	page := pagination.Parse(r.URL.Query())
	items, err := h.svc.ListMessages(r.Context(), userID, conversationID, page)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, "Messages loaded", items)
}

func (h *Handler) sendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	conversationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid conversation id", sharederrors.CodeValidationError)
		return
	}

	var req SendMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", sharederrors.CodeValidationError)
		return
	}

	item, err := h.svc.SendMessage(r.Context(), userID, conversationID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, "Message sent", item)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "Conversation not found", sharederrors.CodeNotFound)
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrChatNotAllowed):
		response.Error(w, http.StatusForbidden, "Insufficient permissions", sharederrors.CodeForbidden)
	case errors.Is(err, ErrValidation):
		response.Error(w, http.StatusBadRequest, err.Error(), sharederrors.CodeValidationError)
	default:
		h.log.Error("chat request failed", slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "Request failed", sharederrors.CodeInternalError)
	}
}

func decodeJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

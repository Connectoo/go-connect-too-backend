package webhooks

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/payments"
	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

type PaymentWebhookProcessor interface {
	ProcessRazorpayWebhook(ctx context.Context, payload []byte, signature, eventID string) error
}

type Handler struct {
	payments PaymentWebhookProcessor
	log      *slog.Logger
}

func NewHandler(payments PaymentWebhookProcessor, log *slog.Logger) *Handler {
	return &Handler{payments: payments, log: log}
}

func (h *Handler) razorpay(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid webhook payload", sharederrors.CodeValidationError)
		return
	}
	signature := r.Header.Get("X-Razorpay-Signature")
	eventID := r.Header.Get("X-Razorpay-Event-Id")
	if err := h.payments.ProcessRazorpayWebhook(r.Context(), payload, signature, eventID); err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Webhook processed", map[string]string{"provider": payments.ProviderRazorpay})
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, payments.ErrDuplicateWebhook):
		response.JSON(w, http.StatusOK, "Webhook already processed", map[string]string{"provider": payments.ProviderRazorpay})
	case errors.Is(err, payments.ErrInvalidSignature):
		response.Error(w, http.StatusUnauthorized, "Invalid webhook signature", sharederrors.CodeUnauthorized)
	case errors.Is(err, payments.ErrValidation):
		response.Error(w, http.StatusBadRequest, "Invalid webhook payload", sharederrors.CodeValidationError)
	case errors.Is(err, payments.ErrNotFound):
		response.Error(w, http.StatusNotFound, "Payment not found", sharederrors.CodeNotFound)
	default:
		if h.log != nil {
			h.log.Error("webhook processing failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

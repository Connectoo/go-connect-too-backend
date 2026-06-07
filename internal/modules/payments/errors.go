package payments

import "errors"

var (
	ErrValidation           = errors.New("validation error")
	ErrNotFound             = errors.New("payment not found")
	ErrDuplicateWebhook     = errors.New("webhook already processed")
	ErrInvalidSignature     = errors.New("invalid payment signature")
	ErrPaymentNotPending    = errors.New("payment is not pending")
	ErrPaymentNotRefundable = errors.New("payment is not refundable")
	ErrRefundExists         = errors.New("refund already exists")
)

package subscriptions

import "errors"

var (
	ErrValidation        = errors.New("validation error")
	ErrNotFound          = errors.New("subscription not found")
	ErrPlanInactive      = errors.New("subscription plan inactive")
	ErrNoActive          = errors.New("no active subscription")
	ErrAlreadyCancelled  = errors.New("subscription already cancelled")
	ErrInvalidSignature  = errors.New("invalid payment signature")
	ErrPaymentNotPending = errors.New("payment is not pending")
)

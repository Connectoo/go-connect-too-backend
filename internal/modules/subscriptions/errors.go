package subscriptions

import "errors"

var (
	ErrValidation   = errors.New("validation error")
	ErrNotFound     = errors.New("subscription not found")
	ErrPlanInactive = errors.New("subscription plan inactive")
)

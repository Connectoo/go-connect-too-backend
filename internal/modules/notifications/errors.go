package notifications

import "errors"

var (
	ErrNotFound   = errors.New("notification not found")
	ErrForbidden  = errors.New("notification access denied")
	ErrValidation = errors.New("notification validation failed")
)

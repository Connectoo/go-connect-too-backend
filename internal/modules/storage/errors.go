package storage

import "errors"

var (
	ErrNotFound    = errors.New("uploaded file not found")
	ErrValidation  = errors.New("validation error")
	ErrForbidden   = errors.New("forbidden")
	ErrUnavailable = errors.New("storage unavailable")
)

package settings

import "errors"

var (
	ErrValidation = errors.New("validation failed")
	ErrNotFound   = errors.New("setting not found")
)

package admin

import "errors"

var (
	ErrValidation = errors.New("validation failed")
	ErrNotFound   = errors.New("user not found")
)

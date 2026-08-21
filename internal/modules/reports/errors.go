package reports

import "errors"

var (
	ErrNotFound        = errors.New("report not found")
	ErrValidation      = errors.New("validation error")
	ErrAlreadyResolved = errors.New("report already resolved")
)

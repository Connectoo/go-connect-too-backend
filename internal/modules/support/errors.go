package support

import "errors"

var (
	ErrNotFound         = errors.New("support ticket not found")
	ErrValidation       = errors.New("validation error")
	ErrForbidden        = errors.New("forbidden")
	ErrCustomerNotFound = errors.New("customer profile not found")
)

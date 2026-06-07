package users

import "errors"

var (
	ErrNotFound         = errors.New("user not found")
	ErrDuplicateEmail   = errors.New("email already registered")
	ErrDuplicatePhone   = errors.New("phone already registered")
	ErrAddressNotFound  = errors.New("address not found")
	ErrValidation       = errors.New("validation failed")
	ErrForbiddenProfile = errors.New("profile update not allowed")
	ErrDeactivated      = errors.New("account deactivated")
)

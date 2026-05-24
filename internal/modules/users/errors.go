package users

import "errors"

var (
	ErrNotFound       = errors.New("user not found")
	ErrDuplicateEmail = errors.New("email already registered")
	ErrDuplicatePhone = errors.New("phone already registered")
)

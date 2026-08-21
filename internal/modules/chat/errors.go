package chat

import "errors"

var (
	ErrNotFound         = errors.New("conversation not found")
	ErrForbidden        = errors.New("conversation access denied")
	ErrValidation       = errors.New("chat validation failed")
	ErrChatNotAllowed   = errors.New("chat not allowed before booking")
	ErrCustomerNotFound = errors.New("customer profile not found")
	ErrEmployeeNotFound = errors.New("employee profile not found")
)

package kyc

import "errors"

var (
	ErrNotFound      = errors.New("kyc record not found")
	ErrValidation    = errors.New("validation error")
	ErrAlreadyExists = errors.New("kyc already submitted")
)

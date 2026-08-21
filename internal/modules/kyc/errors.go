package kyc

import "errors"

var (
	ErrNotFound       = errors.New("kyc record not found")
	ErrValidation     = errors.New("validation error")
	ErrAlreadyExists  = errors.New("kyc already submitted")
	ErrInvalidStatus  = errors.New("invalid kyc status for operation")
	ErrKYCNotApproved = errors.New("kyc not approved")
	ErrFileNotFound   = errors.New("uploaded file not found")
	ErrFileNotOwned   = errors.New("uploaded file not owned by user")
)

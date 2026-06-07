package employees

import "errors"

var (
	ErrNotFound       = errors.New("employee profile not found")
	ErrValidation     = errors.New("validation error")
	ErrKYCNotApproved = errors.New("kyc not approved")
)

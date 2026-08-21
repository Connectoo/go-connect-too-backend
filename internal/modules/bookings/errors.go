package bookings

import "errors"

var (
	ErrNotFound                = errors.New("booking not found")
	ErrValidation              = errors.New("validation error")
	ErrForbidden               = errors.New("forbidden")
	ErrServiceNotFound         = errors.New("service not found")
	ErrEmployeeNotApproved     = errors.New("employee not approved")
	ErrEmployeeUnavailable     = errors.New("employee unavailable")
	ErrDoubleBooking           = errors.New("double booking")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrCustomerProfileNotFound = errors.New("customer profile not found")
	ErrRebookNotAllowed        = errors.New("booking cannot be rebooked")
)

package reviews

import "errors"

var (
	ErrNotFound                = errors.New("review not found")
	ErrValidation              = errors.New("validation error")
	ErrForbidden               = errors.New("forbidden")
	ErrBookingNotCompleted     = errors.New("booking not completed")
	ErrReviewAlreadyExists     = errors.New("review already exists")
	ErrReplyAlreadyExists      = errors.New("reply already exists")
	ErrInvalidStatus           = errors.New("invalid review status")
	ErrCustomerProfileNotFound = errors.New("customer profile not found")
)

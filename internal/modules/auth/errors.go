package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserInactive       = errors.New("user inactive")
	ErrValidation         = errors.New("validation error")

	ErrOTPInvalid         = errors.New("invalid otp code")
	ErrOTPExpired         = errors.New("otp code expired")
	ErrOTPTooManyAttempts = errors.New("too many otp attempts")
	ErrOTPRateLimited     = errors.New("otp request rate limited")
	ErrOTPNotConfigured   = errors.New("otp store not configured")
)

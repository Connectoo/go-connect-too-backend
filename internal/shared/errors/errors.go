package errors

// API error codes returned to clients.
const (
	CodeInternalError     = "INTERNAL_ERROR"
	CodeHealthCheckFailed = "HEALTH_CHECK_FAILED"

	CodeValidationError    = "VALIDATION_ERROR"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeConflict           = "CONFLICT"
	CodeNotFound           = "NOT_FOUND"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
)

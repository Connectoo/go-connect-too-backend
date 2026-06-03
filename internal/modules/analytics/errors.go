package analytics

import "errors"

var (
	ErrInvalidDateRange = errors.New("invalid date range")
	ErrNotFound         = errors.New("employee profile not found")
)

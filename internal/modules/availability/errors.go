package availability

import "errors"

var (
	ErrNotFound   = errors.New("availability slot not found")
	ErrValidation = errors.New("validation error")
	ErrOverlap    = errors.New("availability slot overlaps existing slot")
)

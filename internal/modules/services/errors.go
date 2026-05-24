package services

import "errors"

var (
	ErrNotFound          = errors.New("service not found")
	ErrValidation        = errors.New("validation error")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrProfileIncomplete = errors.New("employee profile incomplete")
)

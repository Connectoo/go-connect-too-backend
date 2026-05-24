package categories

import "errors"

var (
	ErrValidation    = errors.New("validation error")
	ErrDuplicateName = errors.New("category name already exists")
)

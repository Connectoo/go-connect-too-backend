package pagination

import (
	"net/url"
	"strconv"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// Params holds parsed pagination query values.
type Params struct {
	Page  int
	Limit int
}

// Result wraps list items with pagination metadata.
type Result[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// Parse reads page and limit from query parameters.
func Parse(values url.Values) Params {
	page := intValue(values.Get("page"), DefaultPage)
	limit := intValue(values.Get("limit"), DefaultLimit)

	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Params{Page: page, Limit: limit}
}

// Offset returns the SQL offset for the current page.
func (p Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

// NewResult builds a paginated response envelope.
func NewResult[T any](items []T, params Params, total int) Result[T] {
	if items == nil {
		items = []T{}
	}
	return Result[T]{
		Items: items,
		Page:  params.Page,
		Limit: params.Limit,
		Total: total,
	}
}

func intValue(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

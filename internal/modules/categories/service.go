package categories

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxNameLength = 100

// CategoryStore loads and creates categories.
type CategoryStore interface {
	ListActive(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, category *Category) (*Category, error)
}

// Service handles category business logic.
type Service struct {
	categories CategoryStore
	now        func() time.Time
}

// NewService creates a category service.
func NewService(categories CategoryStore) *Service {
	return &Service{
		categories: categories,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

// ListActive returns active categories for authenticated users.
func (s *Service) ListActive(ctx context.Context) ([]CategoryResponse, error) {
	items, err := s.categories.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]CategoryResponse, 0, len(items))
	for i := range items {
		out = append(out, *toCategoryResponse(&items[i]))
	}
	return out, nil
}

// CreateCategory creates a category (admin).
func (s *Service) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*CategoryResponse, error) {
	name, description, err := validateCreateCategory(req)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	at := s.now()
	created, err := s.categories.Create(ctx, &Category{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		IsActive:    isActive,
		CreatedAt:   at,
		UpdatedAt:   at,
	})
	if err != nil {
		return nil, err
	}

	return toCategoryResponse(created), nil
}

func validateCreateCategory(req CreateCategoryRequest) (string, *string, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return "", nil, fmt.Errorf("%w: name is required", ErrValidation)
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		return "", nil, fmt.Errorf("%w: name must be at most %d characters", ErrValidation, maxNameLength)
	}

	var description *string
	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed != "" {
			description = &trimmed
		}
	}

	return name, description, nil
}

func toCategoryResponse(category *Category) *CategoryResponse {
	return &CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
		CreatedAt:   category.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   category.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

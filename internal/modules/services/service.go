package services

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

const (
	maxTitleLength       = 150
	maxDescriptionLength = 1000
)

// EmployeeProfileStore resolves authenticated users to employee profiles.
type EmployeeProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*employees.Profile, error)
}

// Store persists employee services and category checks.
type Store interface {
	CategoryExists(ctx context.Context, categoryID uuid.UUID) (bool, error)
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]EmployeeService, error)
	Create(ctx context.Context, service *EmployeeService) (*EmployeeService, error)
	Update(ctx context.Context, employeeID, serviceID uuid.UUID, service *EmployeeService, at time.Time) (*EmployeeService, error)
	Delete(ctx context.Context, employeeID, serviceID uuid.UUID) error
	UpdateStatus(ctx context.Context, employeeID, serviceID uuid.UUID, isActive bool, at time.Time) (*EmployeeService, error)
}

// Service handles employee service business logic.
type Service struct {
	profiles EmployeeProfileStore
	store    Store
	now      func() time.Time
}

// NewService creates an employee service manager.
func NewService(profiles EmployeeProfileStore, store Store) *Service {
	return &Service{
		profiles: profiles,
		store:    store,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

// Create creates a service listing for the authenticated employee.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateServiceRequest) (*ServiceResponse, error) {
	profile, err := s.profileForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	title, description, err := validateServiceFields(req.CategoryID, req.Title, req.Description, req.Price, req.DurationMinutes)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCategoryExists(ctx, req.CategoryID); err != nil {
		return nil, err
	}
	if req.IsActive && !profileIsComplete(profile) {
		return nil, ErrProfileIncomplete
	}

	at := s.now()
	created, err := s.store.Create(ctx, &EmployeeService{
		ID:              uuid.New(),
		EmployeeID:      profile.ID,
		CategoryID:      req.CategoryID,
		Title:           title,
		Description:     description,
		Price:           req.Price,
		DurationMinutes: req.DurationMinutes,
		IsActive:        req.IsActive,
		CreatedAt:       at,
		UpdatedAt:       at,
	})
	if err != nil {
		return nil, err
	}

	return toResponse(created), nil
}

// List returns services owned by the authenticated employee.
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]ServiceResponse, error) {
	profile, err := s.profileForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, err := s.store.ListByEmployeeID(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	out := make([]ServiceResponse, 0, len(items))
	for i := range items {
		out = append(out, *toResponse(&items[i]))
	}
	return out, nil
}

// Update replaces a service listing owned by the authenticated employee.
func (s *Service) Update(ctx context.Context, userID, serviceID uuid.UUID, req UpdateServiceRequest) (*ServiceResponse, error) {
	profile, err := s.profileForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	title, description, err := validateServiceFields(req.CategoryID, req.Title, req.Description, req.Price, req.DurationMinutes)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCategoryExists(ctx, req.CategoryID); err != nil {
		return nil, err
	}
	if req.IsActive && !profileIsComplete(profile) {
		return nil, ErrProfileIncomplete
	}

	updated, err := s.store.Update(ctx, profile.ID, serviceID, &EmployeeService{
		CategoryID:      req.CategoryID,
		Title:           title,
		Description:     description,
		Price:           req.Price,
		DurationMinutes: req.DurationMinutes,
		IsActive:        req.IsActive,
	}, s.now())
	if err != nil {
		return nil, err
	}

	return toResponse(updated), nil
}

// Delete removes a service listing owned by the authenticated employee.
func (s *Service) Delete(ctx context.Context, userID, serviceID uuid.UUID) error {
	profile, err := s.profileForUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.store.Delete(ctx, profile.ID, serviceID)
}

// UpdateStatus activates or deactivates a service listing owned by the authenticated employee.
func (s *Service) UpdateStatus(ctx context.Context, userID, serviceID uuid.UUID, req UpdateServiceStatusRequest) (*ServiceResponse, error) {
	profile, err := s.profileForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if req.IsActive && !profileIsComplete(profile) {
		return nil, ErrProfileIncomplete
	}

	updated, err := s.store.UpdateStatus(ctx, profile.ID, serviceID, req.IsActive, s.now())
	if err != nil {
		return nil, err
	}
	return toResponse(updated), nil
}

func (s *Service) profileForUser(ctx context.Context, userID uuid.UUID) (*employees.Profile, error) {
	return s.profiles.GetByUserID(ctx, userID)
}

func (s *Service) ensureCategoryExists(ctx context.Context, categoryID uuid.UUID) error {
	exists, err := s.store.CategoryExists(ctx, categoryID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCategoryNotFound
	}
	return nil
}

func validateServiceFields(categoryID uuid.UUID, title string, description *string, price float64, durationMinutes int) (string, *string, error) {
	if categoryID == uuid.Nil {
		return "", nil, fmt.Errorf("%w: category_id is required", ErrValidation)
	}

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return "", nil, fmt.Errorf("%w: title is required", ErrValidation)
	}
	if utf8.RuneCountInString(trimmedTitle) > maxTitleLength {
		return "", nil, fmt.Errorf("%w: title must be at most %d characters", ErrValidation, maxTitleLength)
	}
	if price <= 0 {
		return "", nil, fmt.Errorf("%w: price must be greater than zero", ErrValidation)
	}
	if durationMinutes <= 0 {
		return "", nil, fmt.Errorf("%w: duration_minutes must be greater than zero", ErrValidation)
	}

	var trimmedDescription *string
	if description != nil {
		value := strings.TrimSpace(*description)
		if utf8.RuneCountInString(value) > maxDescriptionLength {
			return "", nil, fmt.Errorf("%w: description must be at most %d characters", ErrValidation, maxDescriptionLength)
		}
		if value != "" {
			trimmedDescription = &value
		}
	}

	return trimmedTitle, trimmedDescription, nil
}

func profileIsComplete(profile *employees.Profile) bool {
	if profile == nil || profile.DisplayName == nil || profile.Phone == nil {
		return false
	}
	return strings.TrimSpace(*profile.DisplayName) != "" && strings.TrimSpace(*profile.Phone) != ""
}

func toResponse(service *EmployeeService) *ServiceResponse {
	return &ServiceResponse{
		ID:              service.ID,
		EmployeeID:      service.EmployeeID,
		CategoryID:      service.CategoryID,
		Title:           service.Title,
		Description:     service.Description,
		Price:           service.Price,
		DurationMinutes: service.DurationMinutes,
		IsActive:        service.IsActive,
		CreatedAt:       service.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       service.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

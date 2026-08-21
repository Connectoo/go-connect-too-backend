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
	CountActiveByEmployeeID(ctx context.Context, employeeID uuid.UUID) (int, error)
	CountActiveExcludingID(ctx context.Context, employeeID, serviceID uuid.UUID) (int, error)
	ListPublicActive(ctx context.Context, categoryID *uuid.UUID, limit int) ([]EmployeeService, error)
	GetPublicActiveByID(ctx context.Context, serviceID uuid.UUID) (*EmployeeService, error)
	ListActiveByEmployeeProfileID(ctx context.Context, employeeID uuid.UUID) ([]EmployeeService, error)
	Create(ctx context.Context, service *EmployeeService) (*EmployeeService, error)
	Update(ctx context.Context, employeeID, serviceID uuid.UUID, service *EmployeeService, at time.Time) (*EmployeeService, error)
	Delete(ctx context.Context, employeeID, serviceID uuid.UUID) error
	UpdateStatus(ctx context.Context, employeeID, serviceID uuid.UUID, isActive bool, at time.Time) (*EmployeeService, error)
}

// SubscriptionLimitStore returns the current active plan limit for an employee.
type SubscriptionLimitStore interface {
	CurrentServiceLimit(ctx context.Context, employeeID uuid.UUID, at time.Time) (int, error)
}

// Service handles employee service business logic.
type Service struct {
	profiles EmployeeProfileStore
	store    Store
	limits   SubscriptionLimitStore
	now      func() time.Time
}

// NewService creates an employee service manager.
func NewService(profiles EmployeeProfileStore, store Store, limits ...SubscriptionLimitStore) *Service {
	var limitStore SubscriptionLimitStore
	if len(limits) > 0 {
		limitStore = limits[0]
	}
	return &Service{
		profiles: profiles,
		store:    store,
		limits:   limitStore,
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
	if req.IsActive {
		if err := s.ensureServiceLimit(ctx, profile.ID); err != nil {
			return nil, err
		}
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
	if req.IsActive {
		if err := s.ensureServiceLimit(ctx, profile.ID, serviceID); err != nil {
			return nil, err
		}
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
	if req.IsActive {
		if err := s.ensureServiceLimit(ctx, profile.ID); err != nil {
			return nil, err
		}
	}

	updated, err := s.store.UpdateStatus(ctx, profile.ID, serviceID, req.IsActive, s.now())
	if err != nil {
		return nil, err
	}
	return toResponse(updated), nil
}

// ListPublic returns active services from approved employees.
func (s *Service) ListPublic(ctx context.Context, categoryID *uuid.UUID) ([]ServiceResponse, error) {
	items, err := s.store.ListPublicActive(ctx, categoryID, 50)
	if err != nil {
		return nil, err
	}
	return toResponseList(items), nil
}

// GetPublic returns an active service from an approved employee.
func (s *Service) GetPublic(ctx context.Context, serviceID uuid.UUID) (*ServiceResponse, error) {
	service, err := s.store.GetPublicActiveByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	return toResponse(service), nil
}

// ListPublicByEmployee returns active services for an approved employee profile.
func (s *Service) ListPublicByEmployee(ctx context.Context, employeeID uuid.UUID) ([]ServiceResponse, error) {
	items, err := s.store.ListActiveByEmployeeProfileID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	return toResponseList(items), nil
}

func (s *Service) profileForUser(ctx context.Context, userID uuid.UUID) (*employees.Profile, error) {
	return s.profiles.GetByUserID(ctx, userID)
}

func toResponseList(items []EmployeeService) []ServiceResponse {
	out := make([]ServiceResponse, 0, len(items))
	for i := range items {
		out = append(out, *toResponse(&items[i]))
	}
	return out
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

func (s *Service) ensureServiceLimit(ctx context.Context, employeeID uuid.UUID, excludeServiceID ...uuid.UUID) error {
	if s.limits == nil {
		return nil
	}
	limit, err := s.limits.CurrentServiceLimit(ctx, employeeID, s.now())
	if err != nil {
		return err
	}
	if limit < 0 {
		return nil
	}
	var activeCount int
	if len(excludeServiceID) > 0 {
		activeCount, err = s.store.CountActiveExcludingID(ctx, employeeID, excludeServiceID[0])
	} else {
		activeCount, err = s.store.CountActiveByEmployeeID(ctx, employeeID)
	}
	if err != nil {
		return err
	}
	if activeCount >= limit {
		return ErrServiceLimit
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

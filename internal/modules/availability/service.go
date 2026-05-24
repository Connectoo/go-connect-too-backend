package availability

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

// EmployeeProfileStore resolves authenticated users to employee profiles.
type EmployeeProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*employees.Profile, error)
}

// Store persists employee availability slots.
type Store interface {
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]Availability, error)
	Create(ctx context.Context, slot *Availability) (*Availability, error)
	Update(ctx context.Context, employeeID, slotID uuid.UUID, slot *Availability, at time.Time) (*Availability, error)
	Delete(ctx context.Context, employeeID, slotID uuid.UUID) error
}

// Service handles employee availability business logic.
type Service struct {
	profiles EmployeeProfileStore
	store    Store
	now      func() time.Time
}

// NewService creates an availability service.
func NewService(profiles EmployeeProfileStore, store Store) *Service {
	return &Service{
		profiles: profiles,
		store:    store,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

// List returns availability slots for the authenticated employee.
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]AvailabilityResponse, error) {
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, err := s.store.ListByEmployeeID(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	out := make([]AvailabilityResponse, 0, len(items))
	for i := range items {
		out = append(out, *toResponse(&items[i]))
	}
	return out, nil
}

// Create creates an availability slot for the authenticated employee.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateAvailabilityRequest) (*AvailabilityResponse, error) {
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := validateSlotFields(req.DayOfWeek, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	at := s.now()
	created, err := s.store.Create(ctx, &Availability{
		ID:          uuid.New(),
		EmployeeID:  profile.ID,
		DayOfWeek:   req.DayOfWeek,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IsAvailable: isAvailable,
		CreatedAt:   at,
		UpdatedAt:   at,
	})
	if err != nil {
		return nil, err
	}

	return toResponse(created), nil
}

// Update replaces an availability slot owned by the authenticated employee.
func (s *Service) Update(ctx context.Context, userID, slotID uuid.UUID, req UpdateAvailabilityRequest) (*AvailabilityResponse, error) {
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := validateSlotFields(req.DayOfWeek, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	updated, err := s.store.Update(ctx, profile.ID, slotID, &Availability{
		DayOfWeek:   req.DayOfWeek,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IsAvailable: req.IsAvailable,
	}, s.now())
	if err != nil {
		return nil, err
	}

	return toResponse(updated), nil
}

// Delete removes an availability slot owned by the authenticated employee.
func (s *Service) Delete(ctx context.Context, userID, slotID uuid.UUID) error {
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.store.Delete(ctx, profile.ID, slotID)
}

func validateSlotFields(day int, start, end TimeOfDay) error {
	if day < MinDayOfWeek || day > MaxDayOfWeek {
		return fmt.Errorf("%w: day_of_week must be between 0 (Sunday) and 6 (Saturday)", ErrValidation)
	}
	if !start.Before(end) {
		return fmt.Errorf("%w: start_time must be before end_time", ErrValidation)
	}
	return nil
}

func toResponse(slot *Availability) *AvailabilityResponse {
	return &AvailabilityResponse{
		ID:          slot.ID,
		EmployeeID:  slot.EmployeeID,
		DayOfWeek:   slot.DayOfWeek,
		StartTime:   slot.StartTime,
		EndTime:     slot.EndTime,
		IsAvailable: slot.IsAvailable,
		CreatedAt:   slot.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   slot.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

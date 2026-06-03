package services

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

// AdminStore supports admin service management.
type AdminStore interface {
	ListAdmin(ctx context.Context, filter AdminListFilter) ([]EmployeeService, int, error)
	UpdateStatusByID(ctx context.Context, serviceID uuid.UUID, isActive bool, at time.Time) (*EmployeeService, error)
}

// ListForAdmin returns paginated services for admin views.
func (s *Service) ListForAdmin(
	ctx context.Context,
	categoryID *uuid.UUID,
	isActive *bool,
	query string,
	page pagination.Params,
) (pagination.Result[ServiceResponse], error) {
	adminStore, ok := s.store.(AdminStore)
	if !ok {
		return pagination.Result[ServiceResponse]{}, ErrValidation
	}

	items, total, err := adminStore.ListAdmin(ctx, AdminListFilter{
		CategoryID: categoryID,
		IsActive:   isActive,
		Query:      query,
		Offset:     page.Offset(),
		Limit:      page.Limit,
	})
	if err != nil {
		return pagination.Result[ServiceResponse]{}, err
	}

	return pagination.NewResult(toResponseList(items), page, total), nil
}

// AdminActivate marks a service as active.
func (s *Service) AdminActivate(ctx context.Context, serviceID uuid.UUID) (*ServiceResponse, error) {
	return s.adminSetStatus(ctx, serviceID, true)
}

// AdminDeactivate marks a service as inactive.
func (s *Service) AdminDeactivate(ctx context.Context, serviceID uuid.UUID) (*ServiceResponse, error) {
	return s.adminSetStatus(ctx, serviceID, false)
}

func (s *Service) adminSetStatus(ctx context.Context, serviceID uuid.UUID, isActive bool) (*ServiceResponse, error) {
	adminStore, ok := s.store.(AdminStore)
	if !ok {
		return nil, ErrValidation
	}

	updated, err := adminStore.UpdateStatusByID(ctx, serviceID, isActive, s.now())
	if err != nil {
		return nil, err
	}
	return toResponse(updated), nil
}

// ParseOptionalUUID parses an optional UUID query parameter.
func ParseOptionalUUID(raw string) *uuid.UUID {
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

// ParseOptionalBool parses an optional bool query parameter.
func ParseOptionalBool(raw string) *bool {
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &value
}

package badges

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Store persists badges.
type Store interface {
	AwardIfNotExists(ctx context.Context, employeeID uuid.UUID, badgeType string, at time.Time) (*Badge, error)
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]Badge, error)
}

// Service handles badge business logic.
type Service struct {
	store Store
	now   func() time.Time
}

// NewService creates a badges service.
func NewService(store Store) *Service {
	return &Service{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

// AwardVerifiedBookingReview awards the verified booking review badge.
func (s *Service) AwardVerifiedBookingReview(ctx context.Context, employeeID uuid.UUID) error {
	_, err := s.store.AwardIfNotExists(ctx, employeeID, TypeVerifiedBookingReview, s.now())
	return err
}

// ListForEmployee returns badges for an employee profile.
func (s *Service) ListForEmployee(ctx context.Context, employeeID uuid.UUID) ([]string, error) {
	items, err := s.store.ListByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for i := range items {
		out = append(out, items[i].BadgeType)
	}
	return out, nil
}

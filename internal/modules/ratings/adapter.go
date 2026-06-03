package ratings

import (
	"context"

	"github.com/google/uuid"
)

type ratingRefresher struct {
	svc *Service
}

// NewRefresher adapts the ratings service to review moderation hooks.
func NewRefresher(svc *Service) interface {
	RefreshEmployeeRating(ctx context.Context, employeeID uuid.UUID) error
} {
	return ratingRefresher{svc: svc}
}

func (r ratingRefresher) RefreshEmployeeRating(ctx context.Context, employeeID uuid.UUID) error {
	_, err := r.svc.RefreshEmployeeRating(ctx, employeeID)
	return err
}

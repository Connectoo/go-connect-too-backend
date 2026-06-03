package users

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// StatusUpdater adapts Repository for modules that only need status changes.
type StatusUpdater struct {
	repo *Repository
}

// NewStatusUpdater creates a status updater adapter.
func NewStatusUpdater(repo *Repository) *StatusUpdater {
	return &StatusUpdater{repo: repo}
}

// UpdateStatus sets the account status for a user.
func (u *StatusUpdater) UpdateStatus(ctx context.Context, id uuid.UUID, status string, at time.Time) error {
	_, err := u.repo.UpdateStatus(ctx, id, status, at)
	return err
}

package kyc

import (
	"context"

	"github.com/google/uuid"
)

// StatusAdapter exposes KYC status lookup for other modules.
type StatusAdapter struct {
	repo *Repository
}

// NewStatusAdapter creates a KYC status adapter.
func NewStatusAdapter(repo *Repository) *StatusAdapter {
	return &StatusAdapter{repo: repo}
}

// GetStatusByEmployeeID returns the KYC status for an employee profile.
func (a *StatusAdapter) GetStatusByEmployeeID(ctx context.Context, employeeID uuid.UUID) (string, error) {
	record, err := a.repo.GetByEmployeeID(ctx, employeeID)
	if err != nil {
		return "", err
	}
	return record.Status, nil
}

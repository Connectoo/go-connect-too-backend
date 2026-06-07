package kyc

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

// EmployeeVerificationAdapter syncs employee profile verification status.
type EmployeeVerificationAdapter struct {
	repo *employees.Repository
}

// NewEmployeeVerificationAdapter creates an adapter backed by the employees repository.
func NewEmployeeVerificationAdapter(repo *employees.Repository) *EmployeeVerificationAdapter {
	return &EmployeeVerificationAdapter{repo: repo}
}

// UpdateVerificationStatus sets verification_status on an employee profile.
func (a *EmployeeVerificationAdapter) UpdateVerificationStatus(ctx context.Context, employeeID uuid.UUID, status string, at time.Time) error {
	_, err := a.repo.UpdateVerificationStatus(ctx, employeeID, status, at)
	return err
}

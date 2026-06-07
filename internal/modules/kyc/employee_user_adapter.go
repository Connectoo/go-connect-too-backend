package kyc

import (
	"context"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

// EmployeeUserAdapter resolves user ids from employee profile ids.
type EmployeeUserAdapter struct {
	repo *employees.Repository
}

// NewEmployeeUserAdapter creates an employee user lookup adapter.
func NewEmployeeUserAdapter(repo *employees.Repository) *EmployeeUserAdapter {
	return &EmployeeUserAdapter{repo: repo}
}

// GetUserIDByEmployeeID returns the user id linked to an employee profile.
func (a *EmployeeUserAdapter) GetUserIDByEmployeeID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error) {
	profile, err := a.repo.GetByID(ctx, employeeID)
	if err != nil {
		return uuid.Nil, err
	}
	return profile.UserID, nil
}

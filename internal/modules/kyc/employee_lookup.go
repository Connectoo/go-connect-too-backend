package kyc

import (
	"context"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

// EmployeeRepositoryAdapter resolves employee profile ids from user ids.
type EmployeeRepositoryAdapter struct {
	repo *employees.Repository
}

// NewEmployeeRepositoryAdapter creates an employee lookup backed by the employees repository.
func NewEmployeeRepositoryAdapter(repo *employees.Repository) *EmployeeRepositoryAdapter {
	return &EmployeeRepositoryAdapter{repo: repo}
}

// GetByUserID returns the employee profile id for the given user.
func (a *EmployeeRepositoryAdapter) GetByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	profile, err := a.repo.GetByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
	return profile.ID, nil
}

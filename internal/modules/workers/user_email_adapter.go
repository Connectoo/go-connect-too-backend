package workers

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
)

// UserEmailAdapter resolves user emails from the users repository.
type UserEmailAdapter struct {
	users *users.Repository
}

// NewUserEmailAdapter creates a user email lookup adapter.
func NewUserEmailAdapter(users *users.Repository) *UserEmailAdapter {
	return &UserEmailAdapter{users: users}
}

// GetByID returns the email for a user id.
func (a *UserEmailAdapter) GetByID(ctx context.Context, id uuid.UUID) (string, error) {
	user, err := a.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return "", err
		}
		return "", err
	}
	return user.Email, nil
}

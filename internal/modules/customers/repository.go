package customers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists customer profiles.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a customer repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetByUserID loads a customer profile for the given user.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	query := `SELECT id, user_id, created_at, updated_at FROM customer_profiles WHERE user_id = $1`

	var profile Profile
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get customer profile by user id: %w", err)
	}
	return &profile, nil
}

// CreateForUserInTx inserts a profile row for a newly registered customer user.
func (r *Repository) CreateForUserInTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, at time.Time) error {
	query := `
		INSERT INTO customer_profiles (id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`

	_, err := tx.ExecContext(ctx, query, uuid.New(), userID, at, at)
	if err != nil {
		return fmt.Errorf("insert customer profile: %w", err)
	}
	return nil
}

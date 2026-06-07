package users

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const userColumns = `id, name, email, phone, password_hash, role, status, email_verified_at, deactivated_at, created_at, updated_at`

// Deactivate marks the authenticated user's account as deactivated.
func (r *Repository) Deactivate(ctx context.Context, id uuid.UUID, at time.Time) (*User, error) {
	query := `
		UPDATE users
		SET status = $2,
		    deactivated_at = $3,
		    updated_at = $3
		WHERE id = $1 AND deactivated_at IS NULL
		RETURNING ` + userColumns

	row := r.db.QueryRowContext(ctx, query, id, StatusInactive, at)
	return scanUserRow(row)
}

// UpdatePassword replaces the password hash for a user.
func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string, at time.Time) error {
	query := `
		UPDATE users
		SET password_hash = $2,
		    updated_at = $3
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, passwordHash, at)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// MarkEmailVerified sets email_verified_at for a user.
func (r *Repository) MarkEmailVerified(ctx context.Context, id uuid.UUID, at time.Time) (*User, error) {
	query := `
		UPDATE users
		SET email_verified_at = $2,
		    updated_at = $2
		WHERE id = $1
		RETURNING ` + userColumns

	row := r.db.QueryRowContext(ctx, query, id, at)
	return scanUserRow(row)
}

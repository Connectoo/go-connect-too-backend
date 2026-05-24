package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Repository handles refresh token persistence.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an auth repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateRefreshToken stores a hashed refresh token.
func (r *Repository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

// GetActiveByHash returns a non-revoked refresh token by hash.
func (r *Repository) GetActiveByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL`

	row := r.db.QueryRowContext(ctx, query, tokenHash)

	var token RefreshToken
	var revokedAt sql.NullTime
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&revokedAt,
		&token.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("scan refresh token: %w", err)
	}
	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}

	return &token, nil
}

// RevokeByHash marks a refresh token as revoked.
func (r *Repository) RevokeByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE token_hash = $1 AND revoked_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, tokenHash, revokedAt)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke rows affected: %w", err)
	}
	if rows == 0 {
		return ErrInvalidToken
	}

	return nil
}

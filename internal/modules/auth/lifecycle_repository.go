package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreatePasswordResetToken stores a hashed password reset token.
func (r *Repository) CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert password reset token: %w", err)
	}
	return nil
}

// GetActivePasswordResetByHash returns an unused password reset token.
func (r *Repository) GetActivePasswordResetByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL`

	row := r.db.QueryRowContext(ctx, query, tokenHash)
	return scanPasswordResetToken(row)
}

// MarkPasswordResetUsed marks a password reset token as consumed.
func (r *Repository) MarkPasswordResetUsed(ctx context.Context, tokenHash string, usedAt time.Time) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = $2
		WHERE token_hash = $1 AND used_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, tokenHash, usedAt)
	if err != nil {
		return fmt.Errorf("mark password reset used: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvalidToken
	}
	return nil
}

// CreateEmailVerificationToken stores a hashed email verification token.
func (r *Repository) CreateEmailVerificationToken(ctx context.Context, token *EmailVerificationToken) error {
	query := `
		INSERT INTO email_verification_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert email verification token: %w", err)
	}
	return nil
}

// GetActiveEmailVerificationByHash returns an unused email verification token.
func (r *Repository) GetActiveEmailVerificationByHash(ctx context.Context, tokenHash string) (*EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM email_verification_tokens
		WHERE token_hash = $1 AND used_at IS NULL`

	row := r.db.QueryRowContext(ctx, query, tokenHash)
	return scanEmailVerificationToken(row)
}

// MarkEmailVerificationUsed marks an email verification token as consumed.
func (r *Repository) MarkEmailVerificationUsed(ctx context.Context, tokenHash string, usedAt time.Time) error {
	query := `
		UPDATE email_verification_tokens
		SET used_at = $2
		WHERE token_hash = $1 AND used_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, tokenHash, usedAt)
	if err != nil {
		return fmt.Errorf("mark email verification used: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvalidToken
	}
	return nil
}

type lifecycleScanner interface {
	Scan(dest ...any) error
}

func scanPasswordResetToken(row lifecycleScanner) (*PasswordResetToken, error) {
	var token PasswordResetToken
	var usedAt sql.NullTime
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &usedAt, &token.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("scan password reset token: %w", err)
	}
	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}
	return &token, nil
}

func scanEmailVerificationToken(row lifecycleScanner) (*EmailVerificationToken, error) {
	var token EmailVerificationToken
	var usedAt sql.NullTime
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &usedAt, &token.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("scan email verification token: %w", err)
	}
	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}
	return &token, nil
}

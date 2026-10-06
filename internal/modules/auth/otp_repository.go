package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateOTPCode stores a hashed one-time login code.
func (r *Repository) CreateOTPCode(ctx context.Context, code *OTPLoginCode) error {
	query := `
		INSERT INTO otp_login_codes (id, phone, role, code_hash, expires_at, attempt_count, consumed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		code.ID,
		code.Phone,
		code.Role,
		code.CodeHash,
		code.ExpiresAt,
		code.AttemptCount,
		code.ConsumedAt,
		code.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert otp login code: %w", err)
	}
	return nil
}

// GetActiveOTPByPhoneRole returns the latest unconsumed OTP for a phone+role.
func (r *Repository) GetActiveOTPByPhoneRole(ctx context.Context, phone, role string) (*OTPLoginCode, error) {
	query := `
		SELECT id, phone, role, code_hash, expires_at, attempt_count, consumed_at, created_at
		FROM otp_login_codes
		WHERE phone = $1 AND role = $2 AND consumed_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, phone, role)
	return scanOTPLoginCode(row)
}

// IncrementOTPAttempt increments the attempt counter and returns the new value.
func (r *Repository) IncrementOTPAttempt(ctx context.Context, id uuid.UUID) (int, error) {
	query := `
		UPDATE otp_login_codes
		SET attempt_count = attempt_count + 1
		WHERE id = $1
		RETURNING attempt_count`

	var count int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrOTPInvalid
	}
	if err != nil {
		return 0, fmt.Errorf("increment otp attempt: %w", err)
	}
	return count, nil
}

// MarkOTPConsumed marks an OTP code as consumed.
func (r *Repository) MarkOTPConsumed(ctx context.Context, id uuid.UUID, at time.Time) error {
	query := `
		UPDATE otp_login_codes
		SET consumed_at = $2
		WHERE id = $1 AND consumed_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, at)
	if err != nil {
		return fmt.Errorf("mark otp consumed: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOTPInvalid
	}
	return nil
}

// CountOTPRequestsSince counts OTP codes created for a phone+role since a time.
func (r *Repository) CountOTPRequestsSince(ctx context.Context, phone, role string, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM otp_login_codes
		WHERE phone = $1 AND role = $2 AND created_at >= $3`

	var count int
	if err := r.db.QueryRowContext(ctx, query, phone, role, since).Scan(&count); err != nil {
		return 0, fmt.Errorf("count otp requests: %w", err)
	}
	return count, nil
}

// LatestOTPCreatedAt returns the most recent OTP creation time for a phone+role.
func (r *Repository) LatestOTPCreatedAt(ctx context.Context, phone, role string) (*time.Time, error) {
	query := `
		SELECT created_at
		FROM otp_login_codes
		WHERE phone = $1 AND role = $2
		ORDER BY created_at DESC
		LIMIT 1`

	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, phone, role).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("latest otp created at: %w", err)
	}
	return &createdAt, nil
}

func scanOTPLoginCode(row lifecycleScanner) (*OTPLoginCode, error) {
	var code OTPLoginCode
	var consumedAt sql.NullTime
	err := row.Scan(
		&code.ID,
		&code.Phone,
		&code.Role,
		&code.CodeHash,
		&code.ExpiresAt,
		&code.AttemptCount,
		&consumedAt,
		&code.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOTPInvalid
	}
	if err != nil {
		return nil, fmt.Errorf("scan otp login code: %w", err)
	}
	if consumedAt.Valid {
		code.ConsumedAt = &consumedAt.Time
	}
	return &code, nil
}

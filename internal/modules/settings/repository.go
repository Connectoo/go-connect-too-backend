package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Repository persists platform settings.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a settings repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetByKey returns a settings row by key.
func (r *Repository) GetByKey(ctx context.Context, key string) (*Setting, error) {
	query := `SELECT key, value, updated_at FROM platform_settings WHERE key = $1`

	var setting Setting
	if err := r.db.QueryRowContext(ctx, query, key).Scan(&setting.Key, &setting.Value, &setting.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get setting: %w", err)
	}
	return &setting, nil
}

// Upsert stores a settings row by key.
func (r *Repository) Upsert(ctx context.Context, key string, value []byte, at time.Time) (*Setting, error) {
	query := `
		INSERT INTO platform_settings (key, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value,
		    updated_at = EXCLUDED.updated_at
		RETURNING key, value, updated_at`

	var setting Setting
	if err := r.db.QueryRowContext(ctx, query, key, value, at).Scan(&setting.Key, &setting.Value, &setting.UpdatedAt); err != nil {
		return nil, fmt.Errorf("upsert setting: %w", err)
	}
	return &setting, nil
}

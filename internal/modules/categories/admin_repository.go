package categories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// GetByID returns a category by primary key.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	query := `SELECT ` + categoryColumns + ` FROM categories WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	category, err := scanCategory(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get category by id: %w", err)
	}
	return category, nil
}

// Update replaces editable category fields.
func (r *Repository) Update(ctx context.Context, id uuid.UUID, category *Category, at time.Time) (*Category, error) {
	query := `
		UPDATE categories
		SET name = $2,
		    description = $3,
		    is_active = $4,
		    updated_at = $5
		WHERE id = $1
		RETURNING ` + categoryColumns

	row := r.db.QueryRowContext(ctx, query,
		id,
		category.Name,
		category.Description,
		category.IsActive,
		at,
	)

	updated, err := scanCategory(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "categories_name_lower_unique" {
			return nil, ErrDuplicateName
		}
		return nil, fmt.Errorf("update category: %w", err)
	}
	return updated, nil
}

// Delete soft-deactivates a category.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID, at time.Time) error {
	query := `
		UPDATE categories
		SET is_active = false,
		    updated_at = $2
		WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id, at)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted category rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// CountActive returns the number of active categories.
func (r *Repository) CountActive(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories WHERE is_active = true`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active categories: %w", err)
	}
	return count, nil
}

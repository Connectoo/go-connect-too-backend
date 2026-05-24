package categories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// Repository persists categories.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a category repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const categoryColumns = `id, name, description, is_active, created_at, updated_at`

// ListActive returns active categories ordered by name.
func (r *Repository) ListActive(ctx context.Context) ([]Category, error) {
	query := `
		SELECT ` + categoryColumns + `
		FROM categories
		WHERE is_active = true
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list active categories: %w", err)
	}
	defer rows.Close()

	var out []Category
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		out = append(out, *category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	if out == nil {
		out = []Category{}
	}
	return out, nil
}

// Create inserts a new category.
func (r *Repository) Create(ctx context.Context, category *Category) (*Category, error) {
	query := `
		INSERT INTO categories (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + categoryColumns

	row := r.db.QueryRowContext(ctx, query,
		category.ID,
		category.Name,
		category.Description,
		category.IsActive,
		category.CreatedAt,
		category.UpdatedAt,
	)

	created, err := scanCategory(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "categories_name_lower_unique" {
			return nil, ErrDuplicateName
		}
		return nil, fmt.Errorf("insert category: %w", err)
	}
	return created, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCategory(row rowScanner) (*Category, error) {
	var category Category
	var description sql.NullString

	err := row.Scan(
		&category.ID,
		&category.Name,
		&description,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description.Valid {
		category.Description = &description.String
	}

	return &category, nil
}

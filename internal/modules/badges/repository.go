package badges

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists employee badges.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a badges repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const badgeColumns = `id, employee_id, badge_type, created_at`

// AwardIfNotExists inserts a badge when the employee does not already have it.
func (r *Repository) AwardIfNotExists(ctx context.Context, employeeID uuid.UUID, badgeType string, at time.Time) (*Badge, error) {
	query := `
		INSERT INTO badges (id, employee_id, badge_type, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (employee_id, badge_type) DO NOTHING
		RETURNING` + badgeColumns

	row := r.db.QueryRowContext(ctx, query, uuid.New(), employeeID, badgeType, at)
	badge, err := scanBadge(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return r.GetByEmployeeAndType(ctx, employeeID, badgeType)
		}
		return nil, fmt.Errorf("award badge: %w", err)
	}
	return badge, nil
}

// GetByEmployeeAndType loads a badge for an employee and type.
func (r *Repository) GetByEmployeeAndType(ctx context.Context, employeeID uuid.UUID, badgeType string) (*Badge, error) {
	query := `SELECT` + badgeColumns + ` FROM badges WHERE employee_id = $1 AND badge_type = $2`
	row := r.db.QueryRowContext(ctx, query, employeeID, badgeType)
	badge, err := scanBadge(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get badge: %w", err)
	}
	return badge, nil
}

// ListByEmployeeID returns all badges for an employee.
func (r *Repository) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]Badge, error) {
	query := `SELECT` + badgeColumns + ` FROM badges WHERE employee_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, employeeID)
	if err != nil {
		return nil, fmt.Errorf("list badges: %w", err)
	}
	defer rows.Close()

	items := make([]Badge, 0)
	for rows.Next() {
		badge, err := scanBadge(rows)
		if err != nil {
			return nil, fmt.Errorf("scan badge: %w", err)
		}
		items = append(items, *badge)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate badges: %w", err)
	}
	return items, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBadge(row rowScanner) (*Badge, error) {
	var badge Badge
	err := row.Scan(&badge.ID, &badge.EmployeeID, &badge.BadgeType, &badge.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &badge, nil
}

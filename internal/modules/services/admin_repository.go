package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AdminListFilter filters admin service listings.
type AdminListFilter struct {
	CategoryID *uuid.UUID
	IsActive   *bool
	Query      string
	Offset     int
	Limit      int
}

// ListAdmin returns all employee services for admin views.
func (r *Repository) ListAdmin(ctx context.Context, filter AdminListFilter) ([]EmployeeService, int, error) {
	where, args := buildServiceAdminWhere(filter)

	countQuery := `SELECT COUNT(*) FROM employee_services es` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin services: %w", err)
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	query := `
		SELECT es.id, es.employee_id, es.category_id, es.title, es.description, es.price,
		       es.duration_minutes, es.is_active, es.created_at, es.updated_at
		FROM employee_services es` + where + fmt.Sprintf(`
		ORDER BY es.created_at DESC
		LIMIT $%d OFFSET $%d`, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin services: %w", err)
	}
	defer rows.Close()

	var out []EmployeeService
	for rows.Next() {
		service, err := scanEmployeeService(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan admin service: %w", err)
		}
		out = append(out, *service)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate admin services: %w", err)
	}
	if out == nil {
		out = []EmployeeService{}
	}
	return out, total, nil
}

// GetByID loads a service by primary key.
func (r *Repository) GetByID(ctx context.Context, serviceID uuid.UUID) (*EmployeeService, error) {
	query := `SELECT` + employeeServiceColumns + ` FROM employee_services WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, serviceID)
	service, err := scanEmployeeService(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get service by id: %w", err)
	}
	return service, nil
}

// UpdateStatusByID activates or deactivates a service without employee ownership checks.
func (r *Repository) UpdateStatusByID(ctx context.Context, serviceID uuid.UUID, isActive bool, at time.Time) (*EmployeeService, error) {
	query := `
		UPDATE employee_services
		SET is_active = $2,
		    updated_at = $3
		WHERE id = $1
		RETURNING` + employeeServiceColumns

	row := r.db.QueryRowContext(ctx, query, serviceID, isActive, at)
	updated, err := scanEmployeeService(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update service status by id: %w", err)
	}
	return updated, nil
}

func buildServiceAdminWhere(filter AdminListFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := []any{}
	pos := 1

	if filter.CategoryID != nil {
		clauses = append(clauses, fmt.Sprintf("es.category_id = $%d", pos))
		args = append(args, *filter.CategoryID)
		pos++
	}
	if filter.IsActive != nil {
		clauses = append(clauses, fmt.Sprintf("es.is_active = $%d", pos))
		args = append(args, *filter.IsActive)
		pos++
	}
	if filter.Query != "" {
		clauses = append(clauses, fmt.Sprintf("(es.title ILIKE $%d OR COALESCE(es.description, '') ILIKE $%d)", pos, pos))
		args = append(args, "%"+strings.TrimSpace(filter.Query)+"%")
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

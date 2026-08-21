package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

const publicServiceSelect = `
	es.id, es.employee_id, es.category_id, es.title, es.description, es.price,
	es.duration_minutes, es.is_active, es.created_at, es.updated_at`

const publicServiceJoin = `
	FROM employee_services es
	INNER JOIN employee_profiles ep ON ep.id = es.employee_id
	WHERE es.is_active = true
	  AND ep.verification_status = $1`

// ListPublicActive returns active services from approved employees.
func (r *Repository) ListPublicActive(ctx context.Context, categoryID *uuid.UUID, limit int) ([]EmployeeService, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := `SELECT` + publicServiceSelect + publicServiceJoin
	args := []any{employees.VerificationApproved, limit}

	if categoryID != nil {
		query += ` AND es.category_id = $2 ORDER BY es.created_at DESC LIMIT $3`
		args = []any{employees.VerificationApproved, *categoryID, limit}
	} else {
		query += ` ORDER BY es.created_at DESC LIMIT $2`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list public services: %w", err)
	}
	defer rows.Close()

	return scanEmployeeServiceRows(rows)
}

// GetPublicActiveByID returns an active service from an approved employee.
func (r *Repository) GetPublicActiveByID(ctx context.Context, serviceID uuid.UUID) (*EmployeeService, error) {
	query := `SELECT` + publicServiceSelect + publicServiceJoin + ` AND es.id = $2`

	row := r.db.QueryRowContext(ctx, query, employees.VerificationApproved, serviceID)
	service, err := scanEmployeeService(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get public service: %w", err)
	}
	return service, nil
}

// CountActivePublic returns the number of active public services.
func (r *Repository) CountActivePublic(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM employee_services es
		INNER JOIN employee_profiles ep ON ep.id = es.employee_id
		WHERE es.is_active = true AND ep.verification_status = $1`

	var count int
	if err := r.db.QueryRowContext(ctx, query, employees.VerificationApproved).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active public services: %w", err)
	}
	return count, nil
}

// ListActiveByEmployeeProfileID returns active services for an approved employee profile.
func (r *Repository) ListActiveByEmployeeProfileID(ctx context.Context, employeeID uuid.UUID) ([]EmployeeService, error) {
	query := `
		SELECT` + employeeServiceColumns + `
		FROM employee_services es
		INNER JOIN employee_profiles ep ON ep.id = es.employee_id
		WHERE es.employee_id = $1
		  AND es.is_active = true
		  AND ep.verification_status = $2
		ORDER BY es.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, employeeID, employees.VerificationApproved)
	if err != nil {
		return nil, fmt.Errorf("list employee public services: %w", err)
	}
	defer rows.Close()

	return scanEmployeeServiceRows(rows)
}

func scanEmployeeServiceRows(rows *sql.Rows) ([]EmployeeService, error) {
	var out []EmployeeService
	for rows.Next() {
		service, err := scanEmployeeService(rows)
		if err != nil {
			return nil, fmt.Errorf("scan employee service: %w", err)
		}
		out = append(out, *service)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate employee services: %w", err)
	}
	if out == nil {
		out = []EmployeeService{}
	}
	return out, nil
}

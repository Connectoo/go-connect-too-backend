package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists employee services.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an employee services repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const employeeServiceColumns = `
	id, employee_id, category_id, title, description, price, duration_minutes, is_active, created_at, updated_at`

// CategoryExists checks whether a category exists.
func (r *Repository) CategoryExists(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check category exists: %w", err)
	}
	return exists, nil
}

// ListByEmployeeID returns services owned by an employee.
func (r *Repository) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]EmployeeService, error) {
	query := `SELECT` + employeeServiceColumns + ` FROM employee_services WHERE employee_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, employeeID)
	if err != nil {
		return nil, fmt.Errorf("list employee services: %w", err)
	}
	defer rows.Close()

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

// Create inserts a service listing.
func (r *Repository) Create(ctx context.Context, service *EmployeeService) (*EmployeeService, error) {
	query := `
		INSERT INTO employee_services (
			id, employee_id, category_id, title, description, price, duration_minutes, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING` + employeeServiceColumns

	row := r.db.QueryRowContext(ctx, query,
		service.ID,
		service.EmployeeID,
		service.CategoryID,
		service.Title,
		service.Description,
		service.Price,
		service.DurationMinutes,
		service.IsActive,
		service.CreatedAt,
		service.UpdatedAt,
	)

	created, err := scanEmployeeService(row)
	if err != nil {
		return nil, fmt.Errorf("insert employee service: %w", err)
	}
	return created, nil
}

// Update replaces service fields when the service belongs to the employee.
func (r *Repository) Update(ctx context.Context, employeeID, serviceID uuid.UUID, service *EmployeeService, at time.Time) (*EmployeeService, error) {
	query := `
		UPDATE employee_services
		SET category_id = $3,
		    title = $4,
		    description = $5,
		    price = $6,
		    duration_minutes = $7,
		    is_active = $8,
		    updated_at = $9
		WHERE id = $1 AND employee_id = $2
		RETURNING` + employeeServiceColumns

	row := r.db.QueryRowContext(ctx, query,
		serviceID,
		employeeID,
		service.CategoryID,
		service.Title,
		service.Description,
		service.Price,
		service.DurationMinutes,
		service.IsActive,
		at,
	)

	updated, err := scanEmployeeService(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update employee service: %w", err)
	}
	return updated, nil
}

// Delete removes a service when it belongs to the employee.
func (r *Repository) Delete(ctx context.Context, employeeID, serviceID uuid.UUID) error {
	query := `DELETE FROM employee_services WHERE id = $1 AND employee_id = $2`

	res, err := r.db.ExecContext(ctx, query, serviceID, employeeID)
	if err != nil {
		return fmt.Errorf("delete employee service: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted employee service rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateStatus activates or deactivates a service when it belongs to the employee.
func (r *Repository) UpdateStatus(ctx context.Context, employeeID, serviceID uuid.UUID, isActive bool, at time.Time) (*EmployeeService, error) {
	query := `
		UPDATE employee_services
		SET is_active = $3,
		    updated_at = $4
		WHERE id = $1 AND employee_id = $2
		RETURNING` + employeeServiceColumns

	row := r.db.QueryRowContext(ctx, query, serviceID, employeeID, isActive, at)
	updated, err := scanEmployeeService(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update employee service status: %w", err)
	}
	return updated, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEmployeeService(row rowScanner) (*EmployeeService, error) {
	var service EmployeeService
	var description sql.NullString

	err := row.Scan(
		&service.ID,
		&service.EmployeeID,
		&service.CategoryID,
		&service.Title,
		&description,
		&service.Price,
		&service.DurationMinutes,
		&service.IsActive,
		&service.CreatedAt,
		&service.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if description.Valid {
		service.Description = &description.String
	}

	return &service, nil
}

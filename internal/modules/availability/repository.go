package availability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists employee availability slots.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an availability repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const availabilityColumns = `
	id, employee_id, day_of_week, start_time, end_time, is_available, created_at, updated_at`

// ListByEmployeeID returns availability slots for an employee ordered by day and start time.
func (r *Repository) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]Availability, error) {
	query := `SELECT` + availabilityColumns + `
		FROM employee_availability
		WHERE employee_id = $1
		ORDER BY day_of_week ASC, start_time ASC`

	rows, err := r.db.QueryContext(ctx, query, employeeID)
	if err != nil {
		return nil, fmt.Errorf("list employee availability: %w", err)
	}
	defer rows.Close()

	var out []Availability
	for rows.Next() {
		slot, err := scanAvailability(rows)
		if err != nil {
			return nil, fmt.Errorf("scan employee availability: %w", err)
		}
		out = append(out, *slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate employee availability: %w", err)
	}
	if out == nil {
		out = []Availability{}
	}
	return out, nil
}

// Create inserts a new availability slot, rejecting overlaps within the same employee + day.
func (r *Repository) Create(ctx context.Context, slot *Availability) (*Availability, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin availability create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	overlap, err := overlapsInTx(ctx, tx, slot.EmployeeID, slot.DayOfWeek, slot.StartTime, slot.EndTime, uuid.Nil)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrOverlap
	}

	created, err := insertInTx(ctx, tx, slot)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit availability create transaction: %w", err)
	}
	return created, nil
}

// Update replaces an availability slot owned by the employee, rejecting overlaps.
func (r *Repository) Update(ctx context.Context, employeeID, slotID uuid.UUID, slot *Availability, at time.Time) (*Availability, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin availability update transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	existing, err := getByIDInTx(ctx, tx, slotID)
	if err != nil {
		return nil, err
	}
	if existing.EmployeeID != employeeID {
		return nil, ErrNotFound
	}

	overlap, err := overlapsInTx(ctx, tx, employeeID, slot.DayOfWeek, slot.StartTime, slot.EndTime, slotID)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrOverlap
	}

	updated, err := updateInTx(ctx, tx, employeeID, slotID, slot, at)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit availability update transaction: %w", err)
	}
	return updated, nil
}

// Delete removes an availability slot owned by the employee.
func (r *Repository) Delete(ctx context.Context, employeeID, slotID uuid.UUID) error {
	query := `DELETE FROM employee_availability WHERE id = $1 AND employee_id = $2`

	res, err := r.db.ExecContext(ctx, query, slotID, employeeID)
	if err != nil {
		return fmt.Errorf("delete employee availability: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted availability rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func insertInTx(ctx context.Context, tx *sql.Tx, slot *Availability) (*Availability, error) {
	query := `
		INSERT INTO employee_availability (
			id, employee_id, day_of_week, start_time, end_time, is_available, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING` + availabilityColumns

	row := tx.QueryRowContext(ctx, query,
		slot.ID,
		slot.EmployeeID,
		slot.DayOfWeek,
		slot.StartTime,
		slot.EndTime,
		slot.IsAvailable,
		slot.CreatedAt,
		slot.UpdatedAt,
	)

	created, err := scanAvailability(row)
	if err != nil {
		return nil, fmt.Errorf("insert employee availability: %w", err)
	}
	return created, nil
}

func updateInTx(ctx context.Context, tx *sql.Tx, employeeID, slotID uuid.UUID, slot *Availability, at time.Time) (*Availability, error) {
	query := `
		UPDATE employee_availability
		SET day_of_week = $3,
		    start_time = $4,
		    end_time = $5,
		    is_available = $6,
		    updated_at = $7
		WHERE id = $1 AND employee_id = $2
		RETURNING` + availabilityColumns

	row := tx.QueryRowContext(ctx, query,
		slotID,
		employeeID,
		slot.DayOfWeek,
		slot.StartTime,
		slot.EndTime,
		slot.IsAvailable,
		at,
	)

	updated, err := scanAvailability(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update employee availability: %w", err)
	}
	return updated, nil
}

func getByIDInTx(ctx context.Context, tx *sql.Tx, slotID uuid.UUID) (*Availability, error) {
	query := `SELECT` + availabilityColumns + ` FROM employee_availability WHERE id = $1`

	row := tx.QueryRowContext(ctx, query, slotID)
	slot, err := scanAvailability(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get employee availability by id: %w", err)
	}
	return slot, nil
}

func overlapsInTx(ctx context.Context, tx *sql.Tx, employeeID uuid.UUID, day int, start, end TimeOfDay, excludeID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM employee_availability
			WHERE employee_id = $1
			  AND day_of_week = $2
			  AND id <> $3
			  AND start_time < $5
			  AND end_time > $4
		)`

	var exists bool
	if err := tx.QueryRowContext(ctx, query, employeeID, day, excludeID, start, end).Scan(&exists); err != nil {
		return false, fmt.Errorf("check employee availability overlap: %w", err)
	}
	return exists, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAvailability(row rowScanner) (*Availability, error) {
	var slot Availability
	err := row.Scan(
		&slot.ID,
		&slot.EmployeeID,
		&slot.DayOfWeek,
		&slot.StartTime,
		&slot.EndTime,
		&slot.IsAvailable,
		&slot.CreatedAt,
		&slot.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

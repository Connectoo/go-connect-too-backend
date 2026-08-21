package bookings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/availability"
	"github.com/MustafaKheda/go-connect-too-backend/internal/platform/database"
)

// Repository persists bookings and status history.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a bookings repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const bookingColumns = `
	id, customer_id, employee_id, service_id, booking_date, start_time, end_time,
	status, customer_notes, employee_notes, total_amount, source_booking_id, rescheduled_from_id, created_at, updated_at`

// Create inserts a booking and its initial status history inside a transaction.
func (r *Repository) Create(ctx context.Context, booking *Booking, changedByUserID uuid.UUID) (*Booking, error) {
	var created *Booking
	err := database.RunInTx(ctx, r.db, func(tx *sql.Tx) error {
		overlap, err := hasOverlappingBookingInTx(ctx, tx, booking.EmployeeID, booking.BookingDate, booking.StartTime, booking.EndTime, uuid.Nil)
		if err != nil {
			return err
		}
		if overlap {
			return ErrDoubleBooking
		}

		inserted, err := insertBookingInTx(ctx, tx, booking)
		if err != nil {
			return err
		}

		if err := insertStatusHistoryInTx(ctx, tx, inserted.ID, nil, inserted.Status, changedByUserID, nil); err != nil {
			return err
		}

		created = inserted
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateStatus updates booking status and appends history inside a transaction.
func (r *Repository) UpdateStatus(
	ctx context.Context,
	bookingID uuid.UUID,
	newStatus string,
	changedByUserID uuid.UUID,
	reason *string,
	employeeNotes *string,
	at time.Time,
) (*Booking, error) {
	var updated *Booking
	err := database.RunInTx(ctx, r.db, func(tx *sql.Tx) error {
		existing, err := getByIDInTx(ctx, tx, bookingID, true)
		if err != nil {
			return err
		}

		if employeeNotes != nil {
			existing.EmployeeNotes = employeeNotes
		}

		query := `
			UPDATE bookings
			SET status = $2,
			    employee_notes = $3,
			    updated_at = $4
			WHERE id = $1
			RETURNING` + bookingColumns

		row := tx.QueryRowContext(ctx, query, bookingID, newStatus, existing.EmployeeNotes, at)
		booking, err := scanBooking(row)
		if err != nil {
			return fmt.Errorf("update booking status: %w", err)
		}

		oldStatus := existing.Status
		if err := insertStatusHistoryInTx(ctx, tx, bookingID, &oldStatus, newStatus, changedByUserID, reason); err != nil {
			return err
		}

		updated = booking
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// GetByID loads a booking by primary key.
func (r *Repository) GetByID(ctx context.Context, bookingID uuid.UUID) (*Booking, error) {
	query := `SELECT` + bookingColumns + ` FROM bookings WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, bookingID)
	booking, err := scanBooking(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get booking by id: %w", err)
	}
	return booking, nil
}

// ListByCustomerID returns bookings for a customer profile.
func (r *Repository) ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]Booking, error) {
	query := `SELECT` + bookingColumns + ` FROM bookings WHERE customer_id = $1 ORDER BY booking_date DESC, start_time DESC`
	return r.list(ctx, query, customerID)
}

// ListByEmployeeID returns bookings for an employee profile.
func (r *Repository) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]Booking, error) {
	query := `SELECT` + bookingColumns + ` FROM bookings WHERE employee_id = $1 ORDER BY booking_date DESC, start_time DESC`
	return r.list(ctx, query, employeeID)
}

// ListAll returns all bookings for admin views.
func (r *Repository) ListAll(ctx context.Context) ([]Booking, error) {
	query := `SELECT` + bookingColumns + ` FROM bookings ORDER BY created_at DESC LIMIT 200`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all bookings: %w", err)
	}
	defer rows.Close()

	var out []Booking
	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		out = append(out, *booking)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}
	if out == nil {
		out = []Booking{}
	}
	return out, nil
}

// EmployeeIsAvailable reports whether the employee has a covering availability slot.
func (r *Repository) EmployeeIsAvailable(
	ctx context.Context,
	employeeID uuid.UUID,
	dayOfWeek int,
	start, end availability.TimeOfDay,
) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM employee_availability
			WHERE employee_id = $1
			  AND day_of_week = $2
			  AND is_available = true
			  AND start_time <= $3
			  AND end_time >= $4
		)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, employeeID, dayOfWeek, start, end).Scan(&exists); err != nil {
		return false, fmt.Errorf("check employee availability: %w", err)
	}
	return exists, nil
}

func (r *Repository) list(ctx context.Context, query string, id uuid.UUID) ([]Booking, error) {
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	var out []Booking
	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		out = append(out, *booking)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}
	if out == nil {
		out = []Booking{}
	}
	return out, nil
}

func insertBookingInTx(ctx context.Context, tx *sql.Tx, booking *Booking) (*Booking, error) {
	query := `
		INSERT INTO bookings (
			id, customer_id, employee_id, service_id, booking_date, start_time, end_time,
			status, customer_notes, employee_notes, total_amount, source_booking_id, rescheduled_from_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING` + bookingColumns

	row := tx.QueryRowContext(ctx, query,
		booking.ID,
		booking.CustomerID,
		booking.EmployeeID,
		booking.ServiceID,
		booking.BookingDate,
		booking.StartTime,
		booking.EndTime,
		booking.Status,
		booking.CustomerNotes,
		booking.EmployeeNotes,
		booking.TotalAmount,
		booking.SourceBookingID,
		booking.RescheduledFromID,
		booking.CreatedAt,
		booking.UpdatedAt,
	)
	return scanBooking(row)
}

// Reschedule updates booking schedule with overlap prevention.
func (r *Repository) Reschedule(
	ctx context.Context,
	bookingID uuid.UUID,
	bookingDate time.Time,
	start, end availability.TimeOfDay,
	changedByUserID uuid.UUID,
	reason *string,
	at time.Time,
) (*Booking, error) {
	var updated *Booking
	err := database.RunInTx(ctx, r.db, func(tx *sql.Tx) error {
		existing, err := getByIDInTx(ctx, tx, bookingID, true)
		if err != nil {
			return err
		}

		overlap, err := hasOverlappingBookingInTx(ctx, tx, existing.EmployeeID, bookingDate, start, end, bookingID)
		if err != nil {
			return err
		}
		if overlap {
			return ErrDoubleBooking
		}

		query := `
			UPDATE bookings
			SET booking_date = $2,
			    start_time = $3,
			    end_time = $4,
			    updated_at = $5
			WHERE id = $1
			RETURNING` + bookingColumns

		row := tx.QueryRowContext(ctx, query, bookingID, bookingDate, start, end, at)
		booking, err := scanBooking(row)
		if err != nil {
			return fmt.Errorf("reschedule booking: %w", err)
		}

		if err := insertStatusHistoryInTx(ctx, tx, bookingID, &existing.Status, existing.Status, changedByUserID, reason); err != nil {
			return err
		}

		updated = booking
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func insertStatusHistoryInTx(
	ctx context.Context,
	tx *sql.Tx,
	bookingID uuid.UUID,
	oldStatus *string,
	newStatus string,
	changedByUserID uuid.UUID,
	reason *string,
) error {
	query := `
		INSERT INTO booking_status_history (
			id, booking_id, old_status, new_status, changed_by_user_id, reason, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := tx.ExecContext(ctx, query,
		uuid.New(),
		bookingID,
		oldStatus,
		newStatus,
		changedByUserID,
		reason,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert booking status history: %w", err)
	}
	return nil
}

func getByIDInTx(ctx context.Context, tx *sql.Tx, bookingID uuid.UUID, forUpdate bool) (*Booking, error) {
	query := `SELECT` + bookingColumns + ` FROM bookings WHERE id = $1`
	if forUpdate {
		query += ` FOR UPDATE`
	}

	row := tx.QueryRowContext(ctx, query, bookingID)
	booking, err := scanBooking(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get booking in tx: %w", err)
	}
	return booking, nil
}

func hasOverlappingBookingInTx(
	ctx context.Context,
	tx *sql.Tx,
	employeeID uuid.UUID,
	bookingDate time.Time,
	start, end availability.TimeOfDay,
	excludeID uuid.UUID,
) (bool, error) {
	statuses := ActiveStatuses()
	placeholders := make([]string, len(statuses))
	args := []any{employeeID, bookingDate, start, end, excludeID}
	for i, status := range statuses {
		placeholders[i] = fmt.Sprintf("$%d", i+6)
		args = append(args, status)
	}

	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1
			FROM bookings
			WHERE employee_id = $1
			  AND booking_date = $2
			  AND id <> $5
			  AND start_time < $4
			  AND end_time > $3
			  AND status IN (%s)
			FOR UPDATE
		)`, strings.Join(placeholders, ", "))

	var exists bool
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("check booking overlap: %w", err)
	}
	return exists, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBooking(row rowScanner) (*Booking, error) {
	var booking Booking
	err := row.Scan(
		&booking.ID,
		&booking.CustomerID,
		&booking.EmployeeID,
		&booking.ServiceID,
		&booking.BookingDate,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.CustomerNotes,
		&booking.EmployeeNotes,
		&booking.TotalAmount,
		&booking.SourceBookingID,
		&booking.RescheduledFromID,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

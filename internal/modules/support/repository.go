package support

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists support tickets and messages.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a support repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const ticketColumns = `id, customer_id, subject, status, priority, created_at, updated_at`
const messageColumns = `id, ticket_id, sender_id, message, is_staff, created_at`

// CreateTicket inserts a ticket and initial message in a transaction.
func (r *Repository) CreateTicket(ctx context.Context, ticket *Ticket, message *Message) (*Ticket, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin support ticket tx: %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
		INSERT INTO support_tickets (id, customer_id, subject, status, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING `+ticketColumns,
		ticket.ID, ticket.CustomerID, ticket.Subject, ticket.Status, ticket.Priority, ticket.CreatedAt,
	)
	created, err := scanTicket(row)
	if err != nil {
		return nil, fmt.Errorf("insert support ticket: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO support_messages (id, ticket_id, sender_id, message, is_staff, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		message.ID, created.ID, message.SenderID, message.Message, message.IsStaff, message.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert support message: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit support ticket tx: %w", err)
	}
	return created, nil
}

// GetByID loads a ticket by id.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Ticket, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+ticketColumns+` FROM support_tickets WHERE id = $1`, id)
	ticket, err := scanTicket(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get support ticket: %w", err)
	}
	return ticket, nil
}

// ListByCustomerID returns tickets for a customer.
func (r *Repository) ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]Ticket, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+ticketColumns+` FROM support_tickets WHERE customer_id = $1 ORDER BY updated_at DESC`, customerID)
	if err != nil {
		return nil, fmt.Errorf("list customer support tickets: %w", err)
	}
	defer rows.Close()
	return scanTickets(rows)
}

// ListAdmin returns paginated tickets for admins.
func (r *Repository) ListAdmin(ctx context.Context, status string, offset, limit int) ([]Ticket, int, error) {
	countQuery := `SELECT COUNT(*) FROM support_tickets`
	listQuery := `SELECT ` + ticketColumns + ` FROM support_tickets`
	args := []any{}

	if status != "" {
		countQuery += ` WHERE status = $1`
		listQuery += ` WHERE status = $1`
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count support tickets: %w", err)
	}

	listQuery += ` ORDER BY updated_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin support tickets: %w", err)
	}
	defer rows.Close()

	items, err := scanTickets(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpdateTicket updates ticket status and/or priority.
func (r *Repository) UpdateTicket(ctx context.Context, id uuid.UUID, status, priority *string, at time.Time) (*Ticket, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if status != nil {
		current.Status = *status
	}
	if priority != nil {
		current.Priority = *priority
	}

	row := r.db.QueryRowContext(ctx, `
		UPDATE support_tickets SET status = $2, priority = $3, updated_at = $4 WHERE id = $1
		RETURNING `+ticketColumns,
		id, current.Status, current.Priority, at,
	)
	return scanTicket(row)
}

// AddMessage inserts a support message and bumps ticket updated_at.
func (r *Repository) AddMessage(ctx context.Context, message *Message, at time.Time) (*Message, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin support message tx: %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
		INSERT INTO support_messages (id, ticket_id, sender_id, message, is_staff, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+messageColumns,
		message.ID, message.TicketID, message.SenderID, message.Message, message.IsStaff, message.CreatedAt,
	)
	created, err := scanMessage(row)
	if err != nil {
		return nil, fmt.Errorf("insert support message: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE support_tickets SET updated_at = $2 WHERE id = $1`, message.TicketID, at); err != nil {
		return nil, fmt.Errorf("update support ticket timestamp: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit support message tx: %w", err)
	}
	return created, nil
}

// ListMessages returns messages for a ticket.
func (r *Repository) ListMessages(ctx context.Context, ticketID uuid.UUID) ([]Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+messageColumns+` FROM support_messages WHERE ticket_id = $1 ORDER BY created_at ASC`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list support messages: %w", err)
	}
	defer rows.Close()

	items := make([]Message, 0)
	for rows.Next() {
		item, err := scanMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan support message: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate support messages: %w", err)
	}
	return items, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTicket(row rowScanner) (*Ticket, error) {
	var ticket Ticket
	if err := row.Scan(&ticket.ID, &ticket.CustomerID, &ticket.Subject, &ticket.Status, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt); err != nil {
		return nil, err
	}
	return &ticket, nil
}

func scanTickets(rows *sql.Rows) ([]Ticket, error) {
	items := make([]Ticket, 0)
	for rows.Next() {
		item, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan support ticket: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate support tickets: %w", err)
	}
	return items, nil
}

func scanMessage(row rowScanner) (*Message, error) {
	var message Message
	if err := row.Scan(&message.ID, &message.TicketID, &message.SenderID, &message.Message, &message.IsStaff, &message.CreatedAt); err != nil {
		return nil, err
	}
	return &message, nil
}

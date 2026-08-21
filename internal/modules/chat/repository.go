package chat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists chat conversations and messages.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a chat repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const conversationColumns = `id, customer_id, employee_id, booking_id, created_at, updated_at`
const messageColumns = `id, conversation_id, sender_id, message, attachment_url, content_type, read_at, created_at`

// EnsureForBooking creates a conversation for a booking when missing.
func (r *Repository) EnsureForBooking(ctx context.Context, bookingID, customerID, employeeID uuid.UUID, at time.Time) (*Conversation, error) {
	existing, err := r.GetByBookingID(ctx, bookingID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	query := `
		INSERT INTO chat_conversations (id, customer_id, employee_id, booking_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + conversationColumns

	row := r.db.QueryRowContext(ctx, query, uuid.New(), customerID, employeeID, bookingID, at, at)
	created, err := scanConversation(row)
	if err != nil {
		return nil, fmt.Errorf("insert chat conversation: %w", err)
	}
	return created, nil
}

// GetByBookingID loads a conversation by booking id.
func (r *Repository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*Conversation, error) {
	query := `SELECT ` + conversationColumns + ` FROM chat_conversations WHERE booking_id = $1`
	row := r.db.QueryRowContext(ctx, query, bookingID)
	item, err := scanConversation(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get conversation by booking id: %w", err)
	}
	return item, nil
}

// GetByID loads a conversation by id.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	query := `SELECT ` + conversationColumns + ` FROM chat_conversations WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanConversation(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get conversation: %w", err)
	}
	return item, nil
}

// ListByParticipant returns conversations for a customer or employee profile.
func (r *Repository) ListByParticipant(ctx context.Context, customerID, employeeID *uuid.UUID) ([]Conversation, error) {
	var (
		query string
		args  []any
	)
	switch {
	case customerID != nil:
		query = `SELECT ` + conversationColumns + ` FROM chat_conversations WHERE customer_id = $1 ORDER BY updated_at DESC`
		args = []any{*customerID}
	case employeeID != nil:
		query = `SELECT ` + conversationColumns + ` FROM chat_conversations WHERE employee_id = $1 ORDER BY updated_at DESC`
		args = []any{*employeeID}
	default:
		return []Conversation{}, nil
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	items := make([]Conversation, 0)
	for rows.Next() {
		item, err := scanConversation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate conversations: %w", err)
	}
	return items, nil
}

// HasBookingBetween reports whether a booking exists between customer and employee.
func (r *Repository) HasBookingBetween(ctx context.Context, customerID, employeeID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM bookings
			WHERE customer_id = $1 AND employee_id = $2
		)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, customerID, employeeID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check booking between participants: %w", err)
	}
	return exists, nil
}

// CreateMessage inserts a chat message and bumps conversation updated_at.
func (r *Repository) CreateMessage(ctx context.Context, message *Message, at time.Time) (*Message, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin chat message tx: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO chat_messages (id, conversation_id, sender_id, message, attachment_url, content_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + messageColumns

	row := tx.QueryRowContext(
		ctx,
		insertQuery,
		message.ID,
		message.ConversationID,
		message.SenderID,
		message.Message,
		message.AttachmentURL,
		message.ContentType,
		message.CreatedAt,
	)
	created, err := scanMessage(row)
	if err != nil {
		return nil, fmt.Errorf("insert chat message: %w", err)
	}

	updateQuery := `UPDATE chat_conversations SET updated_at = $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateQuery, at, message.ConversationID); err != nil {
		return nil, fmt.Errorf("update conversation timestamp: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit chat message tx: %w", err)
	}
	return created, nil
}

// ListMessages returns paginated messages for a conversation.
func (r *Repository) ListMessages(ctx context.Context, conversationID uuid.UUID, offset, limit int) ([]Message, int, error) {
	countQuery := `SELECT COUNT(*) FROM chat_messages WHERE conversation_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, conversationID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count chat messages: %w", err)
	}

	query := `
		SELECT ` + messageColumns + `
		FROM chat_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		OFFSET $2 LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, conversationID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list chat messages: %w", err)
	}
	defer rows.Close()

	items := make([]Message, 0)
	for rows.Next() {
		item, err := scanMessage(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan chat message: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate chat messages: %w", err)
	}
	return items, total, nil
}

// MarkMessageRead sets read_at for a message not sent by the reader.
func (r *Repository) MarkMessageRead(ctx context.Context, conversationID, messageID, readerUserID uuid.UUID, at time.Time) (*Message, error) {
	query := `
		UPDATE chat_messages
		SET read_at = $4
		WHERE id = $1
		  AND conversation_id = $2
		  AND sender_id <> $3
		  AND read_at IS NULL
		RETURNING ` + messageColumns

	row := r.db.QueryRowContext(ctx, query, messageID, conversationID, readerUserID, at)
	item, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("mark message read: %w", err)
	}
	return item, nil
}

type conversationScanner interface {
	Scan(dest ...any) error
}

func scanConversation(row conversationScanner) (*Conversation, error) {
	var item Conversation
	if err := row.Scan(
		&item.ID,
		&item.CustomerID,
		&item.EmployeeID,
		&item.BookingID,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanMessage(row conversationScanner) (*Message, error) {
	var item Message
	if err := row.Scan(
		&item.ID,
		&item.ConversationID,
		&item.SenderID,
		&item.Message,
		&item.AttachmentURL,
		&item.ContentType,
		&item.ReadAt,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

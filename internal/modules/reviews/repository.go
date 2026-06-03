package reviews

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/platform/database"
)

// Repository persists reviews and replies.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a reviews repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const reviewColumns = `
	id, booking_id, customer_id, employee_id, rating, comment, status, created_at, updated_at`

const replyColumns = `
	id, review_id, employee_id, reply, created_at, updated_at`

// Create inserts a review.
func (r *Repository) Create(ctx context.Context, review *Review) (*Review, error) {
	query := `
		INSERT INTO reviews (
			id, booking_id, customer_id, employee_id, rating, comment, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING` + reviewColumns

	row := r.db.QueryRowContext(ctx, query,
		review.ID,
		review.BookingID,
		review.CustomerID,
		review.EmployeeID,
		review.Rating,
		review.Comment,
		review.Status,
		review.CreatedAt,
		review.UpdatedAt,
	)
	return scanReview(row)
}

// GetByID loads a review by primary key.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Review, error) {
	query := `SELECT` + reviewColumns + ` FROM reviews WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	review, err := scanReview(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get review by id: %w", err)
	}
	return review, nil
}

// GetByBookingID loads a review for a booking.
func (r *Repository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*Review, error) {
	query := `SELECT` + reviewColumns + ` FROM reviews WHERE booking_id = $1`
	row := r.db.QueryRowContext(ctx, query, bookingID)
	review, err := scanReview(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get review by booking id: %w", err)
	}
	return review, nil
}

// ListByEmployeeID returns reviews for an employee filtered by status when provided.
func (r *Repository) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID, status string, offset, limit int) ([]Review, int, error) {
	countQuery := `SELECT COUNT(*) FROM reviews WHERE employee_id = $1`
	listQuery := `SELECT` + reviewColumns + ` FROM reviews WHERE employee_id = $1`
	args := []any{employeeID}

	if status != "" {
		countQuery += ` AND status = $2`
		listQuery += ` AND status = $2`
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}

	listQuery += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	items, err := scanReviews(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListAdmin returns paginated reviews for admin moderation.
func (r *Repository) ListAdmin(ctx context.Context, status string, offset, limit int) ([]Review, int, error) {
	countQuery := `SELECT COUNT(*) FROM reviews`
	listQuery := `SELECT` + reviewColumns + ` FROM reviews`
	args := []any{}

	if status != "" {
		countQuery += ` WHERE status = $1`
		listQuery += ` WHERE status = $1`
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin reviews: %w", err)
	}

	listQuery += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin reviews: %w", err)
	}
	defer rows.Close()

	items, err := scanReviews(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpdateStatus sets review moderation status.
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, at time.Time) (*Review, error) {
	query := `
		UPDATE reviews
		SET status = $2, updated_at = $3
		WHERE id = $1
		RETURNING` + reviewColumns

	row := r.db.QueryRowContext(ctx, query, id, status, at)
	review, err := scanReview(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update review status: %w", err)
	}
	return review, nil
}

// CreateReply inserts an employee reply inside a transaction.
func (r *Repository) CreateReply(ctx context.Context, reply *Reply) (*Reply, error) {
	var created *Reply
	err := database.RunInTx(ctx, r.db, func(tx *sql.Tx) error {
		var existingID uuid.UUID
		err := tx.QueryRowContext(ctx, `SELECT id FROM review_replies WHERE review_id = $1`, reply.ReviewID).Scan(&existingID)
		if err == nil {
			return ErrReplyAlreadyExists
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check existing reply: %w", err)
		}

		query := `
			INSERT INTO review_replies (id, review_id, employee_id, reply, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING` + replyColumns

		row := tx.QueryRowContext(ctx, query,
			reply.ID,
			reply.ReviewID,
			reply.EmployeeID,
			reply.Reply,
			reply.CreatedAt,
			reply.UpdatedAt,
		)
		inserted, err := scanReply(row)
		if err != nil {
			return fmt.Errorf("insert review reply: %w", err)
		}
		created = inserted
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// GetReplyByReviewID loads a reply for a review.
func (r *Repository) GetReplyByReviewID(ctx context.Context, reviewID uuid.UUID) (*Reply, error) {
	query := `SELECT` + replyColumns + ` FROM review_replies WHERE review_id = $1`
	row := r.db.QueryRowContext(ctx, query, reviewID)
	reply, err := scanReply(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get review reply: %w", err)
	}
	return reply, nil
}

// ListRepliesByReviewIDs loads replies keyed by review id.
func (r *Repository) ListRepliesByReviewIDs(ctx context.Context, reviewIDs []uuid.UUID) (map[uuid.UUID]Reply, error) {
	if len(reviewIDs) == 0 {
		return map[uuid.UUID]Reply{}, nil
	}

	placeholders := make([]string, len(reviewIDs))
	args := make([]any, len(reviewIDs))
	for i, id := range reviewIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := `SELECT` + replyColumns + ` FROM review_replies WHERE review_id IN (` + joinStrings(placeholders, ",") + `)`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list review replies: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID]Reply)
	for rows.Next() {
		reply, err := scanReply(rows)
		if err != nil {
			return nil, err
		}
		out[reply.ReviewID] = *reply
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review replies: %w", err)
	}
	return out, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanReview(row rowScanner) (*Review, error) {
	var review Review
	err := row.Scan(
		&review.ID,
		&review.BookingID,
		&review.CustomerID,
		&review.EmployeeID,
		&review.Rating,
		&review.Comment,
		&review.Status,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func scanReviews(rows *sql.Rows) ([]Review, error) {
	items := make([]Review, 0)
	for rows.Next() {
		review, err := scanReview(rows)
		if err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		items = append(items, *review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviews: %w", err)
	}
	return items, nil
}

func scanReply(row rowScanner) (*Reply, error) {
	var reply Reply
	err := row.Scan(
		&reply.ID,
		&reply.ReviewID,
		&reply.EmployeeID,
		&reply.Reply,
		&reply.CreatedAt,
		&reply.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

func joinStrings(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for i := 1; i < len(items); i++ {
		out += sep + items[i]
	}
	return out
}

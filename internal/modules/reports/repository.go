package reports

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists user reports.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a reports repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const reportColumns = `
	id, reporter_id, reported_user_id, booking_id, reason, description, status, created_at, updated_at`

// Create inserts a report.
func (r *Repository) Create(ctx context.Context, report *Report) (*Report, error) {
	query := `
		INSERT INTO reports (
			id, reporter_id, reported_user_id, booking_id, reason, description, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING` + reportColumns

	row := r.db.QueryRowContext(ctx, query,
		report.ID,
		report.ReporterID,
		report.ReportedUserID,
		report.BookingID,
		report.Reason,
		report.Description,
		report.Status,
		report.CreatedAt,
		report.UpdatedAt,
	)
	return scanReport(row)
}

// GetByID loads a report by primary key.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Report, error) {
	query := `SELECT` + reportColumns + ` FROM reports WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	report, err := scanReport(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get report by id: %w", err)
	}
	return report, nil
}

// ListAdmin returns paginated reports for admin moderation.
func (r *Repository) ListAdmin(ctx context.Context, status string, offset, limit int) ([]Report, int, error) {
	countQuery := `SELECT COUNT(*) FROM reports`
	listQuery := `SELECT` + reportColumns + ` FROM reports`
	args := []any{}

	if status != "" {
		countQuery += ` WHERE status = $1`
		listQuery += ` WHERE status = $1`
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reports: %w", err)
	}

	listQuery += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	items, err := scanReports(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Resolve marks a report as resolved.
func (r *Repository) Resolve(ctx context.Context, id uuid.UUID, at time.Time) (*Report, error) {
	query := `
		UPDATE reports
		SET status = $2, updated_at = $3
		WHERE id = $1
		RETURNING` + reportColumns

	row := r.db.QueryRowContext(ctx, query, id, StatusResolved, at)
	report, err := scanReport(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("resolve report: %w", err)
	}
	return report, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanReport(row rowScanner) (*Report, error) {
	var report Report
	err := row.Scan(
		&report.ID,
		&report.ReporterID,
		&report.ReportedUserID,
		&report.BookingID,
		&report.Reason,
		&report.Description,
		&report.Status,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func scanReports(rows *sql.Rows) ([]Report, error) {
	items := make([]Report, 0)
	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		items = append(items, *report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports: %w", err)
	}
	return items, nil
}

package kyc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// Repository persists employee KYC records.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a KYC repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// AdminListFilter filters admin KYC listings.
type AdminListFilter struct {
	Status string
	Offset int
	Limit  int
}

const recordColumns = `
	id, employee_id, id_proof_url, address_proof_url, status, rejection_reason,
	reviewed_by, reviewed_at, created_at, updated_at`

// GetByID loads a KYC record by primary key.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Record, error) {
	query := `SELECT` + recordColumns + ` FROM employee_kyc WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	record, err := scanRecord(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get kyc by id: %w", err)
	}
	return record, nil
}

// GetByEmployeeID loads the KYC record for an employee profile.
func (r *Repository) GetByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*Record, error) {
	query := `SELECT` + recordColumns + ` FROM employee_kyc WHERE employee_id = $1`

	row := r.db.QueryRowContext(ctx, query, employeeID)
	record, err := scanRecord(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get kyc by employee id: %w", err)
	}
	return record, nil
}

// ListForAdmin returns paginated KYC records with employee and user metadata.
func (r *Repository) ListForAdmin(ctx context.Context, filter AdminListFilter) ([]AdminListItem, int, error) {
	where, args := buildAdminWhere(filter)

	countQuery := `
		SELECT COUNT(*)
		FROM employee_kyc k
		INNER JOIN employee_profiles ep ON ep.id = k.employee_id
		INNER JOIN users u ON u.id = ep.user_id` + where

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin kyc: %w", err)
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	query := `
		SELECT` + recordColumns + `, ep.display_name, u.name, u.email
		FROM employee_kyc k
		INNER JOIN employee_profiles ep ON ep.id = k.employee_id
		INNER JOIN users u ON u.id = ep.user_id` + where + fmt.Sprintf(`
		ORDER BY k.created_at DESC
		LIMIT $%d OFFSET $%d`, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin kyc: %w", err)
	}
	defer rows.Close()

	var out []AdminListItem
	for rows.Next() {
		item, err := scanAdminListItem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan admin kyc: %w", err)
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate admin kyc: %w", err)
	}
	if out == nil {
		out = []AdminListItem{}
	}
	return out, total, nil
}

// GetAdminByID returns a KYC record with employee and user metadata.
func (r *Repository) GetAdminByID(ctx context.Context, id uuid.UUID) (*AdminListItem, error) {
	query := `
		SELECT` + recordColumns + `, ep.display_name, u.name, u.email
		FROM employee_kyc k
		INNER JOIN employee_profiles ep ON ep.id = k.employee_id
		INNER JOIN users u ON u.id = ep.user_id
		WHERE k.id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanAdminListItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get admin kyc: %w", err)
	}
	return item, nil
}

// Approve marks a pending KYC record as approved.
func (r *Repository) Approve(ctx context.Context, id, reviewerID uuid.UUID, at time.Time) (*Record, error) {
	query := `
		UPDATE employee_kyc
		SET status = $2,
		    rejection_reason = NULL,
		    reviewed_by = $3,
		    reviewed_at = $4,
		    updated_at = $4
		WHERE id = $1 AND status = $5
		RETURNING` + recordColumns

	row := r.db.QueryRowContext(ctx, query, id, StatusApproved, reviewerID, at, StatusPending)
	record, err := scanRecord(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidStatus
		}
		return nil, fmt.Errorf("approve kyc: %w", err)
	}
	return record, nil
}

// Reject marks a pending KYC record as rejected with a reason.
func (r *Repository) Reject(ctx context.Context, id, reviewerID uuid.UUID, reason string, at time.Time) (*Record, error) {
	query := `
		UPDATE employee_kyc
		SET status = $2,
		    rejection_reason = $3,
		    reviewed_by = $4,
		    reviewed_at = $5,
		    updated_at = $5
		WHERE id = $1 AND status = $6
		RETURNING` + recordColumns

	row := r.db.QueryRowContext(ctx, query, id, StatusRejected, reason, reviewerID, at, StatusPending)
	record, err := scanRecord(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidStatus
		}
		return nil, fmt.Errorf("reject kyc: %w", err)
	}
	return record, nil
}

// Submit creates a new KYC record or resubmits after rejection inside a transaction.
func (r *Repository) Submit(ctx context.Context, employeeID uuid.UUID, idProofURL, addressProofURL string, at time.Time) (*Record, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin kyc submit transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	existing, err := getByEmployeeIDInTx(ctx, tx, employeeID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	var record *Record
	switch {
	case existing == nil:
		record, err = insertInTx(ctx, tx, employeeID, idProofURL, addressProofURL, at)
	case existing.Status == StatusRejected:
		record, err = resubmitInTx(ctx, tx, existing.ID, idProofURL, addressProofURL, at)
	default:
		return nil, ErrAlreadyExists
	}
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit kyc submit transaction: %w", err)
	}
	return record, nil
}

func buildAdminWhere(filter AdminListFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := []any{}
	pos := 1

	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("k.status = $%d", pos))
		args = append(args, filter.Status)
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func getByEmployeeIDInTx(ctx context.Context, tx *sql.Tx, employeeID uuid.UUID) (*Record, error) {
	query := `SELECT` + recordColumns + ` FROM employee_kyc WHERE employee_id = $1`

	row := tx.QueryRowContext(ctx, query, employeeID)
	record, err := scanRecord(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get kyc by employee id in tx: %w", err)
	}
	return record, nil
}

func insertInTx(ctx context.Context, tx *sql.Tx, employeeID uuid.UUID, idProofURL, addressProofURL string, at time.Time) (*Record, error) {
	query := `
		INSERT INTO employee_kyc (
			id, employee_id, id_proof_url, address_proof_url, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING` + recordColumns

	row := tx.QueryRowContext(ctx, query,
		uuid.New(),
		employeeID,
		idProofURL,
		addressProofURL,
		StatusPending,
		at,
		at,
	)

	record, err := scanRecord(row)
	if err != nil {
		if isEmployeeKYCUniqueViolation(err) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("insert kyc record: %w", err)
	}
	return record, nil
}

func isEmployeeKYCUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "employee_kyc_employee_id_key"
}

func resubmitInTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, idProofURL, addressProofURL string, at time.Time) (*Record, error) {
	query := `
		UPDATE employee_kyc
		SET id_proof_url = $2,
		    address_proof_url = $3,
		    status = $4,
		    rejection_reason = NULL,
		    reviewed_by = NULL,
		    reviewed_at = NULL,
		    updated_at = $5
		WHERE id = $1
		RETURNING` + recordColumns

	row := tx.QueryRowContext(ctx, query, id, idProofURL, addressProofURL, StatusPending, at)
	record, err := scanRecord(row)
	if err != nil {
		return nil, fmt.Errorf("resubmit kyc record: %w", err)
	}
	return record, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRecord(row rowScanner) (*Record, error) {
	var record Record
	err := row.Scan(
		&record.ID,
		&record.EmployeeID,
		&record.IDProofURL,
		&record.AddressProofURL,
		&record.Status,
		&record.RejectionReason,
		&record.ReviewedBy,
		&record.ReviewedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func scanAdminListItem(row rowScanner) (*AdminListItem, error) {
	var item AdminListItem
	err := row.Scan(
		&item.ID,
		&item.EmployeeID,
		&item.IDProofURL,
		&item.AddressProofURL,
		&item.Status,
		&item.RejectionReason,
		&item.ReviewedBy,
		&item.ReviewedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.EmployeeDisplayName,
		&item.UserName,
		&item.UserEmail,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

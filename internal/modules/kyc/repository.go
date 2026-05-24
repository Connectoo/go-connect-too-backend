package kyc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

const recordColumns = `
	id, employee_id, id_proof_url, address_proof_url, status, rejection_reason, created_at, updated_at`

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
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

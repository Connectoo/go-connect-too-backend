package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists uploaded file metadata.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a storage repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const fileColumns = `id, user_id, object_key, content_type, size_bytes, purpose, created_at, deleted_at`

// Create inserts uploaded file metadata.
func (r *Repository) Create(ctx context.Context, file *UploadedFile) (*UploadedFile, error) {
	query := `
		INSERT INTO uploaded_files (id, user_id, object_key, content_type, size_bytes, purpose, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + fileColumns

	row := r.db.QueryRowContext(ctx, query,
		file.ID,
		file.UserID,
		file.ObjectKey,
		file.ContentType,
		file.SizeBytes,
		file.Purpose,
		file.CreatedAt,
	)

	created, err := scanFile(row)
	if err != nil {
		return nil, fmt.Errorf("insert uploaded file: %w", err)
	}
	return created, nil
}

// GetByID loads an active uploaded file by id.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*UploadedFile, error) {
	query := `SELECT ` + fileColumns + ` FROM uploaded_files WHERE id = $1 AND deleted_at IS NULL`

	row := r.db.QueryRowContext(ctx, query, id)
	file, err := scanFile(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get uploaded file: %w", err)
	}
	return file, nil
}

// SoftDelete marks an uploaded file as deleted for the owning user.
func (r *Repository) SoftDelete(ctx context.Context, userID, fileID uuid.UUID, at time.Time) error {
	query := `
		UPDATE uploaded_files
		SET deleted_at = $3
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, fileID, userID, at)
	if err != nil {
		return fmt.Errorf("delete uploaded file: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete uploaded file rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanFile(row rowScanner) (*UploadedFile, error) {
	var file UploadedFile
	err := row.Scan(
		&file.ID,
		&file.UserID,
		&file.ObjectKey,
		&file.ContentType,
		&file.SizeBytes,
		&file.Purpose,
		&file.CreatedAt,
		&file.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

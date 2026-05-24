package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// Repository handles user persistence.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a user repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type execContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Create inserts a new user.
func (r *Repository) Create(ctx context.Context, user *User) error {
	return r.CreateInTx(ctx, r.db, user)
}

// CreateInTx inserts a new user using the given executor (connection or transaction).
func (r *Repository) CreateInTx(ctx context.Context, exec execContext, user *User) error {
	query := `
		INSERT INTO users (id, name, email, phone, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := exec.ExecContext(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_email_role_unique" {
				return ErrDuplicateEmail
			}
			if pgErr.ConstraintName == "users_phone_role_unique" {
				return ErrDuplicatePhone
			}
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// GetByEmailAndRole returns a user by email and role.
func (r *Repository) GetByEmailAndRole(ctx context.Context, email, role string) (*User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE email = $1 AND role = $2`

	row := r.db.QueryRowContext(ctx, query, email, role)
	return scanUser(row)
}

// GetByID returns a user by id.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*User, error) {
	var user User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &user, nil
}

// NowUTC is used for deterministic timestamps in tests.
var NowUTC = func() time.Time {
	return time.Now().UTC()
}

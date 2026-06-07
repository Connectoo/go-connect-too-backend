package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ListFilter filters admin user listings.
type ListFilter struct {
	Role   string
	Status string
	Query  string
	Offset int
	Limit  int
}

// List returns users matching the filter with total count.
func (r *Repository) List(ctx context.Context, filter ListFilter) ([]User, int, error) {
	where, args := buildUserListWhere(filter)

	countQuery := `SELECT COUNT(*) FROM users` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	query := `
		SELECT ` + userColumns + `
		FROM users` + where + fmt.Sprintf(`
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		user, err := scanUserFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}
	if out == nil {
		out = []User{}
	}
	return out, total, nil
}

// UpdateStatus sets the account status for a user.
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, at time.Time) (*User, error) {
	query := `
		UPDATE users
		SET status = $2,
		    updated_at = $3
		WHERE id = $1
		RETURNING ` + userColumns

	row := r.db.QueryRowContext(ctx, query, id, status, at)
	updated, err := scanUserRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update user status: %w", err)
	}
	return updated, nil
}

func scanUserFromRows(row *sql.Rows) (*User, error) {
	return scanUserRow(row)
}

func buildUserListWhere(filter ListFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := []any{}
	pos := 1

	if filter.Role != "" {
		clauses = append(clauses, fmt.Sprintf("role = $%d", pos))
		args = append(args, filter.Role)
		pos++
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", pos))
		args = append(args, filter.Status)
		pos++
	}
	if filter.Query != "" {
		clauses = append(clauses, fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d)", pos, pos))
		args = append(args, "%"+strings.TrimSpace(filter.Query)+"%")
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

package employees

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// AdminListFilter filters admin employee profile listings.
type AdminListFilter struct {
	VerificationStatus string
	Query              string
	Offset             int
	Limit              int
}

// AdminListItem is an employee profile row with linked user account fields.
type AdminListItem struct {
	Profile
	UserName   string
	UserEmail  string
	UserStatus string
}

// ListAdmin returns employee profiles with user account metadata.
func (r *Repository) ListAdmin(ctx context.Context, filter AdminListFilter) ([]AdminListItem, int, error) {
	where, args := buildEmployeeAdminWhere(filter)

	countQuery := `
		SELECT COUNT(*)
		FROM employee_profiles ep
		INNER JOIN users u ON u.id = ep.user_id` + where

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin employees: %w", err)
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	query := `
		SELECT` + profileColumns + `, u.name, u.email, u.status
		FROM employee_profiles ep
		INNER JOIN users u ON u.id = ep.user_id` + where + fmt.Sprintf(`
		ORDER BY ep.created_at DESC
		LIMIT $%d OFFSET $%d`, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin employees: %w", err)
	}
	defer rows.Close()

	var out []AdminListItem
	for rows.Next() {
		item, err := scanAdminListItem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan admin employee: %w", err)
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate admin employees: %w", err)
	}
	if out == nil {
		out = []AdminListItem{}
	}
	return out, total, nil
}

// GetAdminByID returns an employee profile with linked user account fields.
func (r *Repository) GetAdminByID(ctx context.Context, id uuid.UUID) (*AdminListItem, error) {
	query := `
		SELECT` + profileColumns + `, u.name, u.email, u.status
		FROM employee_profiles ep
		INNER JOIN users u ON u.id = ep.user_id
		WHERE ep.id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanAdminListItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get admin employee: %w", err)
	}
	return item, nil
}

func buildEmployeeAdminWhere(filter AdminListFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := []any{}
	pos := 1

	if filter.VerificationStatus != "" {
		clauses = append(clauses, fmt.Sprintf("ep.verification_status = $%d", pos))
		args = append(args, filter.VerificationStatus)
		pos++
	}
	if filter.Query != "" {
		pattern := "%" + strings.TrimSpace(filter.Query) + "%"
		clauses = append(clauses, fmt.Sprintf(
			"(COALESCE(ep.display_name, '') ILIKE $%d OR u.email ILIKE $%d OR u.name ILIKE $%d)",
			pos, pos, pos,
		))
		args = append(args, pattern)
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func scanAdminListItem(row rowScanner) (*AdminListItem, error) {
	var item AdminListItem
	var languages pgtype.FlatArray[string]
	var skills pgtype.FlatArray[string]

	err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.DisplayName,
		&item.Phone,
		&item.Bio,
		&item.ExperienceYears,
		&item.ProfilePhotoURL,
		&item.LocationText,
		&item.Latitude,
		&item.Longitude,
		&item.ServiceAreaRadiusKm,
		&languages,
		&skills,
		&item.VerificationStatus,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.UserName,
		&item.UserEmail,
		&item.UserStatus,
	)
	if err != nil {
		return nil, err
	}

	item.Languages = []string(languages)
	item.Skills = []string(skills)
	if item.Languages == nil {
		item.Languages = []string{}
	}
	if item.Skills == nil {
		item.Skills = []string{}
	}
	return &item, nil
}

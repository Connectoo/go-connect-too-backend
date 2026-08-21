package bookings

import (
	"context"
	"fmt"
	"strings"
)

// AdminListFilter filters admin booking listings.
type AdminListFilter struct {
	Status string
	Offset int
	Limit  int
}

// ListAdmin returns paginated bookings for admin views.
func (r *Repository) ListAdmin(ctx context.Context, filter AdminListFilter) ([]Booking, int, error) {
	where, args := buildBookingAdminWhere(filter)

	countQuery := `SELECT COUNT(*) FROM bookings` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin bookings: %w", err)
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	query := `SELECT` + bookingColumns + ` FROM bookings` + where + fmt.Sprintf(`
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin bookings: %w", err)
	}
	defer rows.Close()

	var out []Booking
	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan admin booking: %w", err)
		}
		out = append(out, *booking)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate admin bookings: %w", err)
	}
	if out == nil {
		out = []Booking{}
	}
	return out, total, nil
}

func buildBookingAdminWhere(filter AdminListFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := []any{}
	pos := 1

	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", pos))
		args = append(args, filter.Status)
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

package payments

import (
	"context"
	"fmt"
	"strings"
)

// AdminListFilter filters admin payment listings.
type AdminListFilter struct {
	Status string
	Offset int
	Limit  int
}

// ListAdmin returns paginated payments for admin views.
func (r *Repository) ListAdmin(ctx context.Context, filter AdminListFilter) ([]Payment, int, error) {
	where, args := buildPaymentAdminWhere(filter)

	countQuery := `SELECT COUNT(*) FROM payments` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin payments: %w", err)
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	query := `SELECT ` + paymentColumns + ` FROM payments` + where + fmt.Sprintf(`
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, limitPos, offsetPos)

	return r.listWithTotal(ctx, query, listArgs, total)
}

func buildPaymentAdminWhere(filter AdminListFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := []any{}
	pos := 1

	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", pos))
		args = append(args, filter.Status)
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *Repository) listWithTotal(ctx context.Context, query string, args []any, total int) ([]Payment, int, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	out := []Payment{}
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan payment: %w", err)
		}
		out = append(out, *payment)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate payments: %w", err)
	}
	return out, total, nil
}

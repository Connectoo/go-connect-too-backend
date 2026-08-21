package subscriptions

import (
	"context"
	"fmt"
	"strings"
)

// AdminListFilter filters admin subscription listings.
type AdminListFilter struct {
	Status string
	Offset int
	Limit  int
}

// ListAdmin returns paginated employee subscriptions for admin views.
func (r *Repository) ListAdmin(ctx context.Context, filter AdminListFilter) ([]EmployeeSubscription, int, error) {
	where, args := buildSubscriptionAdminWhere(filter)

	countQuery := `
		SELECT COUNT(*)
		FROM employee_subscriptions es` + where

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin subscriptions: %w", err)
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	query := `
		SELECT es.id, es.employee_id, es.plan_id, sp.name, es.status, es.starts_at, es.expires_at,
		       es.auto_renew, es.cancelled_at, es.cancellation_reason, es.created_at, es.updated_at,
		       sp.id, sp.name, sp.price, sp.currency, sp.duration_days, sp.service_limit,
		       sp.is_featured_allowed, sp.is_priority_allowed, sp.is_active, sp.created_at, sp.updated_at
		FROM employee_subscriptions es
		JOIN subscription_plans sp ON sp.id = es.plan_id` + where + fmt.Sprintf(`
		ORDER BY es.created_at DESC
		LIMIT $%d OFFSET $%d`, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin subscriptions: %w", err)
	}
	defer rows.Close()

	out := []EmployeeSubscription{}
	for rows.Next() {
		sub, err := scanSubscriptionWithPlan(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan admin subscription: %w", err)
		}
		out = append(out, *sub)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate admin subscriptions: %w", err)
	}
	return out, total, nil
}

func buildSubscriptionAdminWhere(filter AdminListFilter) (string, []any) {
	clauses := []string{"1 = 1"}
	args := []any{}
	pos := 1

	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("es.status = $%d", pos))
		args = append(args, filter.Status)
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

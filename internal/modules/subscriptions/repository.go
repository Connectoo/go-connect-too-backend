package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const planColumns = `id, name, price, currency, duration_days, service_limit, is_featured_allowed, is_priority_allowed, is_active, created_at, updated_at`

func (r *Repository) ListActivePlans(ctx context.Context) ([]Plan, error) {
	return r.listPlans(ctx, `SELECT `+planColumns+` FROM subscription_plans WHERE is_active = true ORDER BY price ASC`)
}
func (r *Repository) ListAllPlans(ctx context.Context) ([]Plan, error) {
	return r.listPlans(ctx, `SELECT `+planColumns+` FROM subscription_plans ORDER BY price ASC`)
}

func (r *Repository) listPlans(ctx context.Context, query string, args ...any) ([]Plan, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list subscription plans: %w", err)
	}
	defer rows.Close()
	out := []Plan{}
	for rows.Next() {
		plan, err := scanPlan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan subscription plan: %w", err)
		}
		out = append(out, *plan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscription plans: %w", err)
	}
	return out, nil
}

func (r *Repository) GetPlanByID(ctx context.Context, id uuid.UUID) (*Plan, error) {
	plan, err := scanPlan(r.db.QueryRowContext(ctx, `SELECT `+planColumns+` FROM subscription_plans WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get subscription plan: %w", err)
	}
	return plan, nil
}

func (r *Repository) CreatePlan(ctx context.Context, plan *Plan) (*Plan, error) {
	created, err := scanPlan(r.db.QueryRowContext(ctx, `
		INSERT INTO subscription_plans (id, name, price, currency, duration_days, service_limit, is_featured_allowed, is_priority_allowed, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING `+planColumns, plan.ID, plan.Name, plan.Price, plan.Currency, plan.DurationDays, plan.ServiceLimit, plan.IsFeaturedAllowed, plan.IsPriorityAllowed, plan.IsActive, plan.CreatedAt))
	if err != nil {
		return nil, fmt.Errorf("insert subscription plan: %w", err)
	}
	return created, nil
}

func (r *Repository) UpdatePlan(ctx context.Context, id uuid.UUID, plan *Plan, at time.Time) (*Plan, error) {
	updated, err := scanPlan(r.db.QueryRowContext(ctx, `
		UPDATE subscription_plans
		SET name = $2, price = $3, currency = $4, duration_days = $5, service_limit = $6,
		    is_featured_allowed = $7, is_priority_allowed = $8, is_active = $9, updated_at = $10
		WHERE id = $1 RETURNING `+planColumns, id, plan.Name, plan.Price, plan.Currency, plan.DurationDays, plan.ServiceLimit, plan.IsFeaturedAllowed, plan.IsPriorityAllowed, plan.IsActive, at))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update subscription plan: %w", err)
	}
	return updated, nil
}

func (r *Repository) CurrentByEmployeeID(ctx context.Context, employeeID uuid.UUID, at time.Time) (*EmployeeSubscription, error) {
	sub, err := scanSubscriptionWithPlan(r.db.QueryRowContext(ctx, currentSubscriptionQuery()+` LIMIT 1`, employeeID, at))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get current subscription: %w", err)
	}
	return sub, nil
}

func (r *Repository) ListAllSubscriptions(ctx context.Context) ([]EmployeeSubscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT es.id, es.employee_id, es.plan_id, sp.name, es.status, es.starts_at, es.expires_at, es.created_at, es.updated_at,
		       sp.id, sp.name, sp.price, sp.currency, sp.duration_days, sp.service_limit,
		       sp.is_featured_allowed, sp.is_priority_allowed, sp.is_active, sp.created_at, sp.updated_at
		FROM employee_subscriptions es JOIN subscription_plans sp ON sp.id = es.plan_id
		ORDER BY es.created_at DESC LIMIT 200`)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()
	out := []EmployeeSubscription{}
	for rows.Next() {
		sub, err := scanSubscriptionWithPlan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		out = append(out, *sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions: %w", err)
	}
	return out, nil
}

func (r *Repository) CurrentServiceLimit(ctx context.Context, employeeID uuid.UUID, at time.Time) (int, error) {
	var limit int
	err := r.db.QueryRowContext(ctx, `
		SELECT sp.service_limit FROM employee_subscriptions es JOIN subscription_plans sp ON sp.id = es.plan_id
		WHERE es.employee_id = $1 AND es.status = 'active' AND es.starts_at <= $2 AND es.expires_at > $2
		ORDER BY es.expires_at DESC LIMIT 1`, employeeID, at).Scan(&limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 1, nil
		}
		return 0, fmt.Errorf("get current service limit: %w", err)
	}
	return limit, nil
}

func currentSubscriptionQuery() string {
	return `
	SELECT es.id, es.employee_id, es.plan_id, sp.name, es.status, es.starts_at, es.expires_at, es.created_at, es.updated_at,
	       sp.id, sp.name, sp.price, sp.currency, sp.duration_days, sp.service_limit,
	       sp.is_featured_allowed, sp.is_priority_allowed, sp.is_active, sp.created_at, sp.updated_at
	FROM employee_subscriptions es JOIN subscription_plans sp ON sp.id = es.plan_id
	WHERE es.employee_id = $1 AND es.status = 'active' AND es.starts_at <= $2 AND es.expires_at > $2
	ORDER BY es.expires_at DESC`
}

type rowScanner interface{ Scan(dest ...any) error }

func scanPlan(row rowScanner) (*Plan, error) {
	var plan Plan
	err := row.Scan(&plan.ID, &plan.Name, &plan.Price, &plan.Currency, &plan.DurationDays, &plan.ServiceLimit, &plan.IsFeaturedAllowed, &plan.IsPriorityAllowed, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		return nil, err
	}
	plan.Currency = strings.TrimSpace(plan.Currency)
	return &plan, nil
}

func scanSubscriptionWithPlan(row rowScanner) (*EmployeeSubscription, error) {
	var sub EmployeeSubscription
	var startsAt, expiresAt sql.NullTime
	var plan Plan
	err := row.Scan(&sub.ID, &sub.EmployeeID, &sub.PlanID, &sub.PlanName, &sub.Status, &startsAt, &expiresAt, &sub.CreatedAt, &sub.UpdatedAt, &plan.ID, &plan.Name, &plan.Price, &plan.Currency, &plan.DurationDays, &plan.ServiceLimit, &plan.IsFeaturedAllowed, &plan.IsPriorityAllowed, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if startsAt.Valid {
		sub.StartsAt = &startsAt.Time
	}
	if expiresAt.Valid {
		sub.ExpiresAt = &expiresAt.Time
	}
	plan.Currency = strings.TrimSpace(plan.Currency)
	sub.Plan = &plan
	return &sub, nil
}

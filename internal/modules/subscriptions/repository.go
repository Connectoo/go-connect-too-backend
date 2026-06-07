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
		SELECT es.id, es.employee_id, es.plan_id, sp.name, es.status, es.starts_at, es.expires_at,
		       es.auto_renew, es.cancelled_at, es.cancellation_reason, es.created_at, es.updated_at,
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
	SELECT es.id, es.employee_id, es.plan_id, sp.name, es.status, es.starts_at, es.expires_at,
	       es.auto_renew, es.cancelled_at, es.cancellation_reason, es.created_at, es.updated_at,
	       sp.id, sp.name, sp.price, sp.currency, sp.duration_days, sp.service_limit,
	       sp.is_featured_allowed, sp.is_priority_allowed, sp.is_active, sp.created_at, sp.updated_at
	FROM employee_subscriptions es JOIN subscription_plans sp ON sp.id = es.plan_id
	WHERE es.employee_id = $1 AND es.status = 'active' AND es.starts_at <= $2 AND es.expires_at > $2
	ORDER BY es.expires_at DESC`
}

// Cancel marks an active subscription as cancelled.
func (r *Repository) Cancel(ctx context.Context, subscriptionID uuid.UUID, reason *string, at time.Time) (*EmployeeSubscription, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin cancel subscription tx: %w", err)
	}
	defer tx.Rollback()

	var employeeID, planID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		UPDATE employee_subscriptions
		SET status = 'cancelled', cancelled_at = $2, cancellation_reason = $3, auto_renew = false, updated_at = $2
		WHERE id = $1 AND status = 'active'
		RETURNING employee_id, plan_id`, subscriptionID, at, reason).Scan(&employeeID, &planID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoActive
		}
		return nil, fmt.Errorf("cancel subscription: %w", err)
	}

	if err := insertSubscriptionChangeInTx(ctx, tx, subscriptionID, employeeID, ChangeTypeCancel, &planID, nil, reason, at); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit cancel subscription tx: %w", err)
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT es.id, es.employee_id, es.plan_id, sp.name, es.status, es.starts_at, es.expires_at,
		       es.auto_renew, es.cancelled_at, es.cancellation_reason, es.created_at, es.updated_at,
		       sp.id, sp.name, sp.price, sp.currency, sp.duration_days, sp.service_limit,
		       sp.is_featured_allowed, sp.is_priority_allowed, sp.is_active, sp.created_at, sp.updated_at
		FROM employee_subscriptions es JOIN subscription_plans sp ON sp.id = es.plan_id
		WHERE es.id = $1`, subscriptionID)
	return scanSubscriptionWithPlan(row)
}

// CreatePendingPlanChange creates a pending subscription for a plan change.
func (r *Repository) CreatePendingPlanChange(ctx context.Context, employeeID, oldPlanID, newPlanID uuid.UUID, at time.Time) (*EmployeeSubscription, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin change plan tx: %w", err)
	}
	defer tx.Rollback()

	subscriptionID := uuid.New()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO employee_subscriptions (id, employee_id, plan_id, status, auto_renew, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', true, $4, $4)`, subscriptionID, employeeID, newPlanID, at)
	if err != nil {
		return nil, fmt.Errorf("insert pending plan change: %w", err)
	}

	if err := insertSubscriptionChangeInTx(ctx, tx, subscriptionID, employeeID, ChangeTypeChangePlan, &oldPlanID, &newPlanID, nil, at); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit change plan tx: %w", err)
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT es.id, es.employee_id, es.plan_id, sp.name, es.status, es.starts_at, es.expires_at,
		       es.auto_renew, es.cancelled_at, es.cancellation_reason, es.created_at, es.updated_at,
		       sp.id, sp.name, sp.price, sp.currency, sp.duration_days, sp.service_limit,
		       sp.is_featured_allowed, sp.is_priority_allowed, sp.is_active, sp.created_at, sp.updated_at
		FROM employee_subscriptions es JOIN subscription_plans sp ON sp.id = es.plan_id
		WHERE es.id = $1`, subscriptionID)
	return scanSubscriptionWithPlan(row)
}

// SetAutoRenew toggles auto-renew on the active subscription.
func (r *Repository) SetAutoRenew(ctx context.Context, employeeID uuid.UUID, autoRenew bool, at time.Time) (*EmployeeSubscription, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin auto renew tx: %w", err)
	}
	defer tx.Rollback()

	var subscriptionID, planID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		UPDATE employee_subscriptions
		SET auto_renew = $3, updated_at = $4
		WHERE employee_id = $1 AND status = 'active' AND starts_at <= $2 AND expires_at > $2
		RETURNING id, plan_id`, employeeID, at, autoRenew, at).Scan(&subscriptionID, &planID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoActive
		}
		return nil, fmt.Errorf("set auto renew: %w", err)
	}

	reason := fmt.Sprintf("auto_renew=%t", autoRenew)
	if err := insertSubscriptionChangeInTx(ctx, tx, subscriptionID, employeeID, ChangeTypeAutoRenew, &planID, &planID, &reason, at); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit auto renew tx: %w", err)
	}

	return r.CurrentByEmployeeID(ctx, employeeID, at)
}

func insertSubscriptionChangeInTx(
	ctx context.Context,
	tx *sql.Tx,
	subscriptionID, employeeID uuid.UUID,
	changeType string,
	oldPlanID, newPlanID *uuid.UUID,
	reason *string,
	at time.Time,
) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO subscription_changes (id, subscription_id, employee_id, change_type, old_plan_id, new_plan_id, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), subscriptionID, employeeID, changeType, oldPlanID, newPlanID, reason, at,
	)
	if err != nil {
		return fmt.Errorf("insert subscription change: %w", err)
	}
	return nil
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
	var startsAt, expiresAt, cancelledAt sql.NullTime
	var cancellationReason sql.NullString
	var plan Plan
	err := row.Scan(
		&sub.ID, &sub.EmployeeID, &sub.PlanID, &sub.PlanName, &sub.Status, &startsAt, &expiresAt,
		&sub.AutoRenew, &cancelledAt, &cancellationReason, &sub.CreatedAt, &sub.UpdatedAt,
		&plan.ID, &plan.Name, &plan.Price, &plan.Currency, &plan.DurationDays, &plan.ServiceLimit,
		&plan.IsFeaturedAllowed, &plan.IsPriorityAllowed, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if startsAt.Valid {
		sub.StartsAt = &startsAt.Time
	}
	if expiresAt.Valid {
		sub.ExpiresAt = &expiresAt.Time
	}
	if cancelledAt.Valid {
		sub.CancelledAt = &cancelledAt.Time
	}
	if cancellationReason.Valid {
		value := cancellationReason.String
		sub.CancellationReason = &value
	}
	plan.Currency = strings.TrimSpace(plan.Currency)
	sub.Plan = &plan
	return &sub, nil
}

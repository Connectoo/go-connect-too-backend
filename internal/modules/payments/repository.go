package payments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const paymentColumns = `id, employee_id, subscription_id, provider, provider_order_id, provider_payment_id, amount, currency, status, raw_response, created_at, updated_at`

func (r *Repository) CreateSubscriptionOrder(ctx context.Context, employeeID uuid.UUID, plan PlanSnapshot, order *GatewayOrder, at time.Time) (*Payment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin subscription order tx: %w", err)
	}
	defer tx.Rollback()

	subscriptionID := uuid.New()
	paymentID := uuid.New()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO employee_subscriptions (id, employee_id, plan_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', $4, $4)`, subscriptionID, employeeID, plan.ID, at)
	if err != nil {
		return nil, fmt.Errorf("insert pending subscription: %w", err)
	}

	row := tx.QueryRowContext(ctx, `
		INSERT INTO payments (id, employee_id, subscription_id, provider, provider_order_id, amount, currency, status, raw_response, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, $9, $9)
		RETURNING `+paymentColumns, paymentID, employeeID, subscriptionID, ProviderRazorpay, order.ProviderOrderID, order.Amount, order.Currency, order.RawResponse, at)
	payment, err := scanPayment(row)
	if err != nil {
		return nil, fmt.Errorf("insert pending payment: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit subscription order tx: %w", err)
	}
	return payment, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	payment, err := scanPayment(r.db.QueryRowContext(ctx, `SELECT `+paymentColumns+` FROM payments WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get payment by id: %w", err)
	}
	return payment, nil
}

func (r *Repository) GetByIDForEmployee(ctx context.Context, id, employeeID uuid.UUID) (*Payment, error) {
	payment, err := scanPayment(r.db.QueryRowContext(ctx, `SELECT `+paymentColumns+` FROM payments WHERE id = $1 AND employee_id = $2`, id, employeeID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get payment for employee: %w", err)
	}
	return payment, nil
}

func (r *Repository) CreateRefund(ctx context.Context, refund *Refund, at time.Time) (*Refund, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO refunds (id, payment_id, amount, reason, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id, payment_id, amount, reason, status, provider_refund_id, created_by, created_at, updated_at`,
		refund.ID, refund.PaymentID, refund.Amount, refund.Reason, refund.Status, refund.CreatedBy, at,
	)
	var created Refund
	var reason sql.NullString
	var providerRefundID sql.NullString
	if err := row.Scan(&created.ID, &created.PaymentID, &created.Amount, &reason, &created.Status, &providerRefundID, &created.CreatedBy, &created.CreatedAt, &created.UpdatedAt); err != nil {
		return nil, fmt.Errorf("insert refund: %w", err)
	}
	if reason.Valid {
		value := reason.String
		created.Reason = &value
	}
	if providerRefundID.Valid {
		value := providerRefundID.String
		created.ProviderRefundID = &value
	}
	return &created, nil
}

func (r *Repository) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]Payment, error) {
	return r.list(ctx, `SELECT `+paymentColumns+` FROM payments WHERE employee_id = $1 ORDER BY created_at DESC`, employeeID)
}

func (r *Repository) ListAll(ctx context.Context) ([]Payment, error) {
	return r.list(ctx, `SELECT `+paymentColumns+` FROM payments ORDER BY created_at DESC LIMIT 200`)
}

func (r *Repository) list(ctx context.Context, query string, args ...any) ([]Payment, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()
	out := []Payment{}
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		out = append(out, *payment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payments: %w", err)
	}
	return out, nil
}

func (r *Repository) InsertWebhookEvent(ctx context.Context, event *WebhookEvent) (bool, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO payment_webhook_events (id, provider, event_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`, event.ID, event.Provider, event.EventID, event.EventType, event.Payload, event.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return false, nil
		}
		return false, fmt.Errorf("insert webhook event: %w", err)
	}
	return true, nil
}

func (r *Repository) MarkWebhookProcessed(ctx context.Context, eventID uuid.UUID, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payment_webhook_events SET processed_at = $2 WHERE id = $1`, eventID, at)
	if err != nil {
		return fmt.Errorf("mark webhook processed: %w", err)
	}
	return nil
}

func (r *Repository) ActivateByProviderOrder(ctx context.Context, provider, providerOrderID, providerPaymentID string, raw []byte, at time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin payment activation tx: %w", err)
	}
	defer tx.Rollback()

	var payment Payment
	var durationDays int
	var providerPaymentIDNull sql.NullString
	var rawResponse []byte
	err = tx.QueryRowContext(ctx, `
		SELECT p.id, p.employee_id, p.subscription_id, p.provider, p.provider_order_id, p.provider_payment_id,
		       p.amount, p.currency, p.status, p.raw_response, p.created_at, p.updated_at, sp.duration_days
		FROM payments p
		JOIN employee_subscriptions es ON es.id = p.subscription_id
		JOIN subscription_plans sp ON sp.id = es.plan_id
		WHERE p.provider = $1 AND p.provider_order_id = $2
		FOR UPDATE OF p, es`, provider, providerOrderID).Scan(
		&payment.ID, &payment.EmployeeID, &payment.SubscriptionID, &payment.Provider, &payment.ProviderOrderID,
		&providerPaymentIDNull, &payment.Amount, &payment.Currency, &payment.Status, &rawResponse,
		&payment.CreatedAt, &payment.UpdatedAt, &durationDays,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("lock payment for activation: %w", err)
	}
	if payment.Status == StatusSuccess {
		return tx.Commit()
	}
	if payment.Status != StatusPending {
		return ErrPaymentNotPending
	}

	_, err = tx.ExecContext(ctx, `UPDATE employee_subscriptions SET status = 'expired', updated_at = $2 WHERE employee_id = $1 AND status = 'active'`, payment.EmployeeID, at)
	if err != nil {
		return fmt.Errorf("expire previous subscriptions: %w", err)
	}

	expiresAt := at.AddDate(0, 0, durationDays)
	_, err = tx.ExecContext(ctx, `UPDATE employee_subscriptions SET status = 'active', starts_at = $2, expires_at = $3, updated_at = $2 WHERE id = $1`, payment.SubscriptionID, at, expiresAt)
	if err != nil {
		return fmt.Errorf("activate subscription: %w", err)
	}

	_, err = tx.ExecContext(ctx, `UPDATE payments SET status = 'success', provider_payment_id = $3, raw_response = $4, updated_at = $5 WHERE provider = $1 AND provider_order_id = $2`, provider, providerOrderID, providerPaymentID, raw, at)
	if err != nil {
		return fmt.Errorf("mark payment success: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit payment activation tx: %w", err)
	}
	return nil
}

type rowScanner interface{ Scan(dest ...any) error }

func scanPayment(row rowScanner) (*Payment, error) {
	var payment Payment
	var providerPaymentID sql.NullString
	var raw []byte
	err := row.Scan(&payment.ID, &payment.EmployeeID, &payment.SubscriptionID, &payment.Provider, &payment.ProviderOrderID, &providerPaymentID, &payment.Amount, &payment.Currency, &payment.Status, &raw, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if providerPaymentID.Valid {
		payment.ProviderPaymentID = &providerPaymentID.String
	}
	payment.RawResponse = raw
	return &payment, nil
}

package admin

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

// Repository loads dashboard metrics.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an admin repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// DashboardSummary returns basic platform counts.
func (r *Repository) DashboardSummary(ctx context.Context) (*DashboardSummary, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM users WHERE role = 'customer'),
			(SELECT COUNT(*) FROM users WHERE role = 'employee'),
			(SELECT COUNT(*) FROM employee_profiles WHERE verification_status = $1),
			(SELECT COUNT(*) FROM bookings),
			(SELECT COUNT(*) FROM bookings WHERE status IN ('pending', 'accepted', 'in_progress')),
			(SELECT COUNT(*) FROM employee_services),
			(SELECT COUNT(*) FROM employee_services WHERE is_active = true),
			(SELECT COUNT(*) FROM payments),
			(SELECT COUNT(*) FROM payments WHERE status = 'completed'),
			COALESCE((SELECT SUM(amount) FROM payments WHERE status = 'completed'), 0),
			(SELECT COUNT(*) FROM employee_subscriptions WHERE status = 'active')`

	var summary DashboardSummary
	if err := r.db.QueryRowContext(ctx, query, employees.VerificationPending).Scan(
		&summary.TotalUsers,
		&summary.TotalCustomers,
		&summary.TotalEmployees,
		&summary.PendingEmployees,
		&summary.TotalBookings,
		&summary.ActiveBookings,
		&summary.TotalServices,
		&summary.ActiveServices,
		&summary.TotalPayments,
		&summary.CompletedPayments,
		&summary.TotalRevenue,
		&summary.ActiveSubscriptions,
	); err != nil {
		return nil, fmt.Errorf("dashboard summary: %w", err)
	}

	return &summary, nil
}

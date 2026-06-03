package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

// Repository runs analytics SQL aggregations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an analytics repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// RecordProfileView inserts a profile view event.
func (r *Repository) RecordProfileView(ctx context.Context, employeeID uuid.UUID, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO employee_profile_views (employee_id, viewed_at) VALUES ($1, $2)`,
		employeeID, at,
	)
	if err != nil {
		return fmt.Errorf("record profile view: %w", err)
	}
	return nil
}

// EmployeeProfileViews counts profile views in a date range.
func (r *Repository) EmployeeProfileViews(ctx context.Context, employeeID uuid.UUID, dr DateRange) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)::int
		FROM employee_profile_views
		WHERE employee_id = $1
		  AND viewed_at >= $2
		  AND viewed_at < $3`,
		employeeID, dr.From, dr.To,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("employee profile views: %w", err)
	}
	return count, nil
}

// EmployeeBookingCounts returns booking totals by status for an employee in range.
func (r *Repository) EmployeeBookingCounts(ctx context.Context, employeeID uuid.UUID, dr DateRange) (total, completed, cancelled int, err error) {
	query := `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE status = 'completed')::int,
			COUNT(*) FILTER (WHERE status = 'cancelled')::int
		FROM bookings
		WHERE employee_id = $1
		  AND created_at >= $2
		  AND created_at < $3`

	err = r.db.QueryRowContext(ctx, query, employeeID, dr.From, dr.To).Scan(&total, &completed, &cancelled)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("employee booking counts: %w", err)
	}
	return total, completed, cancelled, nil
}

// EmployeeEstimatedRevenue sums completed booking amounts in range.
func (r *Repository) EmployeeEstimatedRevenue(ctx context.Context, employeeID uuid.UUID, dr DateRange) (string, error) {
	var revenue sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_amount), 0)::text
		FROM bookings
		WHERE employee_id = $1
		  AND status = 'completed'
		  AND created_at >= $2
		  AND created_at < $3`,
		employeeID, dr.From, dr.To,
	).Scan(&revenue)
	if err != nil {
		return "", fmt.Errorf("employee estimated revenue: %w", err)
	}
	if revenue.Valid {
		return revenue.String, nil
	}
	return "0", nil
}

// EmployeeRatingTrend returns monthly average ratings for approved reviews.
func (r *Repository) EmployeeRatingTrend(ctx context.Context, employeeID uuid.UUID, dr DateRange) ([]RatingPeriod, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			date_trunc('month', created_at AT TIME ZONE 'UTC')::date,
			ROUND(AVG(rating)::numeric, 2)::float8,
			COUNT(*)::int
		FROM reviews
		WHERE employee_id = $1
		  AND status = 'approved'
		  AND created_at >= $2
		  AND created_at < $3
		GROUP BY 1
		ORDER BY 1`,
		employeeID, dr.From, dr.To,
	)
	if err != nil {
		return nil, fmt.Errorf("employee rating trend: %w", err)
	}
	defer rows.Close()

	return scanRatingPeriods(rows)
}

// EmployeeBookingsByStatus groups employee bookings by status in range.
func (r *Repository) EmployeeBookingsByStatus(ctx context.Context, employeeID uuid.UUID, dr DateRange) ([]StatusCount, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*)::int
		FROM bookings
		WHERE employee_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		GROUP BY status
		ORDER BY status`,
		employeeID, dr.From, dr.To,
	)
	if err != nil {
		return nil, fmt.Errorf("employee bookings by status: %w", err)
	}
	defer rows.Close()
	return scanStatusCounts(rows)
}

// EmployeeBookingsDaily returns daily booking counts for an employee.
func (r *Repository) EmployeeBookingsDaily(ctx context.Context, employeeID uuid.UUID, dr DateRange) ([]BookingDayCount, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT (created_at AT TIME ZONE 'UTC')::date, COUNT(*)::int
		FROM bookings
		WHERE employee_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		GROUP BY 1
		ORDER BY 1`,
		employeeID, dr.From, dr.To,
	)
	if err != nil {
		return nil, fmt.Errorf("employee bookings daily: %w", err)
	}
	defer rows.Close()
	return scanBookingDayCounts(rows)
}

// EmployeeReviewSummary returns aggregate review stats in range.
func (r *Repository) EmployeeReviewSummary(ctx context.Context, employeeID uuid.UUID, dr DateRange) (avg *float64, total int, err error) {
	var avgRating sql.NullFloat64
	err = r.db.QueryRowContext(ctx, `
		SELECT ROUND(AVG(rating)::numeric, 2)::float8, COUNT(*)::int
		FROM reviews
		WHERE employee_id = $1
		  AND status = 'approved'
		  AND created_at >= $2
		  AND created_at < $3`,
		employeeID, dr.From, dr.To,
	).Scan(&avgRating, &total)
	if err != nil {
		return nil, 0, fmt.Errorf("employee review summary: %w", err)
	}
	if avgRating.Valid {
		v := avgRating.Float64
		avg = &v
	}
	return avg, total, nil
}

// EmployeeRatingDistribution counts reviews per star rating in range.
func (r *Repository) EmployeeRatingDistribution(ctx context.Context, employeeID uuid.UUID, dr DateRange) (map[int]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT rating, COUNT(*)::int
		FROM reviews
		WHERE employee_id = $1
		  AND status = 'approved'
		  AND created_at >= $2
		  AND created_at < $3
		GROUP BY rating
		ORDER BY rating`,
		employeeID, dr.From, dr.To,
	)
	if err != nil {
		return nil, fmt.Errorf("employee rating distribution: %w", err)
	}
	defer rows.Close()

	out := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	for rows.Next() {
		var rating, count int
		if err := rows.Scan(&rating, &count); err != nil {
			return nil, fmt.Errorf("scan rating distribution: %w", err)
		}
		out[rating] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rating distribution: %w", err)
	}
	return out, nil
}

// AdminPlatformTotals returns snapshot platform counts (not date-filtered).
func (r *Repository) AdminPlatformTotals(ctx context.Context) (totalUsers, totalEmployees, approvedEmployees, activeSubscriptions int, mrr int64, err error) {
	query := `
		SELECT
			(SELECT COUNT(*)::int FROM users),
			(SELECT COUNT(*)::int FROM users WHERE role = 'employee'),
			(SELECT COUNT(*)::int FROM employee_profiles WHERE verification_status = $1),
			(SELECT COUNT(*)::int FROM employee_subscriptions WHERE status = 'active'),
			COALESCE((
				SELECT SUM(sp.price)::bigint
				FROM employee_subscriptions es
				JOIN subscription_plans sp ON sp.id = es.plan_id
				WHERE es.status = 'active'
			), 0)`

	err = r.db.QueryRowContext(ctx, query, employees.VerificationApproved).Scan(
		&totalUsers,
		&totalEmployees,
		&approvedEmployees,
		&activeSubscriptions,
		&mrr,
	)
	if err != nil {
		return 0, 0, 0, 0, 0, fmt.Errorf("admin platform totals: %w", err)
	}
	return totalUsers, totalEmployees, approvedEmployees, activeSubscriptions, mrr, nil
}

// AdminBookingVolume counts bookings created in range.
func (r *Repository) AdminBookingVolume(ctx context.Context, dr DateRange) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)::int
		FROM bookings
		WHERE created_at >= $1 AND created_at < $2`,
		dr.From, dr.To,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("admin booking volume: %w", err)
	}
	return count, nil
}

// AdminFailedPayments counts failed payments in range.
func (r *Repository) AdminFailedPayments(ctx context.Context, dr DateRange) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)::int
		FROM payments
		WHERE status = 'failed'
		  AND created_at >= $1
		  AND created_at < $2`,
		dr.From, dr.To,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("admin failed payments: %w", err)
	}
	return count, nil
}

// AdminSubscriptionRevenue sums successful subscription payments in range.
func (r *Repository) AdminSubscriptionRevenue(ctx context.Context, dr DateRange) (int64, error) {
	var total int64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)::bigint
		FROM payments
		WHERE status = 'success'
		  AND created_at >= $1
		  AND created_at < $2`,
		dr.From, dr.To,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("admin subscription revenue: %w", err)
	}
	return total, nil
}

// AdminBookingRevenue sums completed booking amounts in range.
func (r *Repository) AdminBookingRevenue(ctx context.Context, dr DateRange) (string, error) {
	var revenue sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_amount), 0)::text
		FROM bookings
		WHERE status = 'completed'
		  AND created_at >= $1
		  AND created_at < $2`,
		dr.From, dr.To,
	).Scan(&revenue)
	if err != nil {
		return "", fmt.Errorf("admin booking revenue: %w", err)
	}
	if revenue.Valid {
		return revenue.String, nil
	}
	return "0", nil
}

// AdminSubscriptionRevenueDaily returns daily successful payment totals.
func (r *Repository) AdminSubscriptionRevenueDaily(ctx context.Context, dr DateRange) ([]RevenueDay, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT (created_at AT TIME ZONE 'UTC')::date, COALESCE(SUM(amount), 0)::bigint
		FROM payments
		WHERE status = 'success'
		  AND created_at >= $1
		  AND created_at < $2
		GROUP BY 1
		ORDER BY 1`,
		dr.From, dr.To,
	)
	if err != nil {
		return nil, fmt.Errorf("admin subscription revenue daily: %w", err)
	}
	defer rows.Close()

	var items []RevenueDay
	for rows.Next() {
		var item RevenueDay
		if err := rows.Scan(&item.Date, &item.Amount); err != nil {
			return nil, fmt.Errorf("scan subscription revenue daily: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscription revenue daily: %w", err)
	}
	return items, nil
}

// AdminBookingsByStatus groups platform bookings by status in range.
func (r *Repository) AdminBookingsByStatus(ctx context.Context, dr DateRange) ([]StatusCount, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*)::int
		FROM bookings
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY status
		ORDER BY status`,
		dr.From, dr.To,
	)
	if err != nil {
		return nil, fmt.Errorf("admin bookings by status: %w", err)
	}
	defer rows.Close()
	return scanStatusCounts(rows)
}

// AdminBookingsDaily returns daily platform booking counts.
func (r *Repository) AdminBookingsDaily(ctx context.Context, dr DateRange) ([]BookingDayCount, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT (created_at AT TIME ZONE 'UTC')::date, COUNT(*)::int
		FROM bookings
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY 1
		ORDER BY 1`,
		dr.From, dr.To,
	)
	if err != nil {
		return nil, fmt.Errorf("admin bookings daily: %w", err)
	}
	defer rows.Close()
	return scanBookingDayCounts(rows)
}

// AdminPopularCategories returns categories ranked by bookings in range.
func (r *Repository) AdminPopularCategories(ctx context.Context, dr DateRange, limit int) ([]CategoryBookingCount, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.name, COUNT(b.id)::int AS booking_count
		FROM bookings b
		JOIN employee_services es ON es.id = b.service_id
		JOIN categories c ON c.id = es.category_id
		WHERE b.created_at >= $1 AND b.created_at < $2
		GROUP BY c.id, c.name
		ORDER BY booking_count DESC, c.name ASC
		LIMIT $3`,
		dr.From, dr.To, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("admin popular categories: %w", err)
	}
	defer rows.Close()

	var items []CategoryBookingCount
	for rows.Next() {
		var item CategoryBookingCount
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.BookingCount); err != nil {
			return nil, fmt.Errorf("scan popular categories: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate popular categories: %w", err)
	}
	return items, nil
}

func scanRatingPeriods(rows *sql.Rows) ([]RatingPeriod, error) {
	var items []RatingPeriod
	for rows.Next() {
		var item RatingPeriod
		if err := rows.Scan(&item.Period, &item.AverageRating, &item.ReviewCount); err != nil {
			return nil, fmt.Errorf("scan rating period: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rating periods: %w", err)
	}
	return items, nil
}

func scanStatusCounts(rows *sql.Rows) ([]StatusCount, error) {
	var items []StatusCount
	for rows.Next() {
		var item StatusCount
		if err := rows.Scan(&item.Status, &item.Count); err != nil {
			return nil, fmt.Errorf("scan status count: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate status counts: %w", err)
	}
	return items, nil
}

func scanBookingDayCounts(rows *sql.Rows) ([]BookingDayCount, error) {
	var items []BookingDayCount
	for rows.Next() {
		var item BookingDayCount
		if err := rows.Scan(&item.Date, &item.Count); err != nil {
			return nil, fmt.Errorf("scan booking day count: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate booking day counts: %w", err)
	}
	return items, nil
}

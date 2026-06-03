package analytics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

// EmployeeProfileLookup resolves an employee profile by user id.
type EmployeeProfileLookup interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*employees.Profile, error)
}

// AnalyticsStore runs aggregated analytics queries.
type AnalyticsStore interface {
	RecordProfileView(ctx context.Context, employeeID uuid.UUID, at time.Time) error
	EmployeeProfileViews(ctx context.Context, employeeID uuid.UUID, dr DateRange) (int, error)
	EmployeeBookingCounts(ctx context.Context, employeeID uuid.UUID, dr DateRange) (total, completed, cancelled int, err error)
	EmployeeEstimatedRevenue(ctx context.Context, employeeID uuid.UUID, dr DateRange) (string, error)
	EmployeeRatingTrend(ctx context.Context, employeeID uuid.UUID, dr DateRange) ([]RatingPeriod, error)
	EmployeeBookingsByStatus(ctx context.Context, employeeID uuid.UUID, dr DateRange) ([]StatusCount, error)
	EmployeeBookingsDaily(ctx context.Context, employeeID uuid.UUID, dr DateRange) ([]BookingDayCount, error)
	EmployeeReviewSummary(ctx context.Context, employeeID uuid.UUID, dr DateRange) (avg *float64, total int, err error)
	EmployeeRatingDistribution(ctx context.Context, employeeID uuid.UUID, dr DateRange) (map[int]int, error)
	AdminPlatformTotals(ctx context.Context) (totalUsers, totalEmployees, approvedEmployees, activeSubscriptions int, mrr int64, err error)
	AdminBookingVolume(ctx context.Context, dr DateRange) (int, error)
	AdminFailedPayments(ctx context.Context, dr DateRange) (int, error)
	AdminSubscriptionRevenue(ctx context.Context, dr DateRange) (int64, error)
	AdminBookingRevenue(ctx context.Context, dr DateRange) (string, error)
	AdminSubscriptionRevenueDaily(ctx context.Context, dr DateRange) ([]RevenueDay, error)
	AdminBookingsByStatus(ctx context.Context, dr DateRange) ([]StatusCount, error)
	AdminBookingsDaily(ctx context.Context, dr DateRange) ([]BookingDayCount, error)
	AdminPopularCategories(ctx context.Context, dr DateRange, limit int) ([]CategoryBookingCount, error)
}

// Service handles analytics business logic.
type Service struct {
	store    AnalyticsStore
	profiles EmployeeProfileLookup
	now      func() time.Time
}

// NewService creates an analytics service.
func NewService(store AnalyticsStore, profiles EmployeeProfileLookup) *Service {
	return &Service{
		store:    store,
		profiles: profiles,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

// RecordProfileView records a public profile view (best-effort).
func (s *Service) RecordProfileView(ctx context.Context, employeeID uuid.UUID) {
	if s.store == nil {
		return
	}
	_ = s.store.RecordProfileView(ctx, employeeID, s.now())
}

// EmployeeSummary returns overview metrics for the authenticated employee.
func (s *Service) EmployeeSummary(ctx context.Context, userID uuid.UUID, fromParam, toParam string) (*EmployeeSummaryResponse, error) {
	profile, dr, err := s.employeeRange(ctx, userID, fromParam, toParam)
	if err != nil {
		return nil, err
	}

	views, err := s.store.EmployeeProfileViews(ctx, profile.ID, dr)
	if err != nil {
		return nil, err
	}
	total, completed, cancelled, err := s.store.EmployeeBookingCounts(ctx, profile.ID, dr)
	if err != nil {
		return nil, err
	}
	revenue, err := s.store.EmployeeEstimatedRevenue(ctx, profile.ID, dr)
	if err != nil {
		return nil, err
	}
	trend, err := s.store.EmployeeRatingTrend(ctx, profile.ID, dr)
	if err != nil {
		return nil, err
	}

	return &EmployeeSummaryResponse{
		Period:                dr.ToResponse(),
		ProfileViews:          views,
		TotalBookings:         total,
		CompletedBookings:     completed,
		CancelledBookings:     cancelled,
		AverageResponseTimeMs: nil,
		EstimatedRevenue:      revenue,
		RatingTrend:           toRatingPoints(trend),
	}, nil
}

// EmployeeBookings returns booking analytics for the authenticated employee.
func (s *Service) EmployeeBookings(ctx context.Context, userID uuid.UUID, fromParam, toParam string) (*EmployeeBookingsResponse, error) {
	profile, dr, err := s.employeeRange(ctx, userID, fromParam, toParam)
	if err != nil {
		return nil, err
	}

	byStatus, err := s.store.EmployeeBookingsByStatus(ctx, profile.ID, dr)
	if err != nil {
		return nil, err
	}
	daily, err := s.store.EmployeeBookingsDaily(ctx, profile.ID, dr)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, item := range byStatus {
		total += item.Count
	}

	return &EmployeeBookingsResponse{
		Period:      dr.ToResponse(),
		Total:       total,
		ByStatus:    toStatusItems(byStatus),
		DailyVolume: toDailyItems(daily),
	}, nil
}

// EmployeeReviews returns review analytics for the authenticated employee.
func (s *Service) EmployeeReviews(ctx context.Context, userID uuid.UUID, fromParam, toParam string) (*EmployeeReviewsResponse, error) {
	profile, dr, err := s.employeeRange(ctx, userID, fromParam, toParam)
	if err != nil {
		return nil, err
	}

	avg, total, err := s.store.EmployeeReviewSummary(ctx, profile.ID, dr)
	if err != nil {
		return nil, err
	}
	trend, err := s.store.EmployeeRatingTrend(ctx, profile.ID, dr)
	if err != nil {
		return nil, err
	}
	distribution, err := s.store.EmployeeRatingDistribution(ctx, profile.ID, dr)
	if err != nil {
		return nil, err
	}

	return &EmployeeReviewsResponse{
		Period:             dr.ToResponse(),
		AverageRating:      avg,
		TotalReviews:       total,
		RatingTrend:        toRatingPoints(trend),
		RatingDistribution: distribution,
	}, nil
}

// AdminSummary returns platform overview metrics.
func (s *Service) AdminSummary(ctx context.Context, fromParam, toParam string) (*AdminSummaryResponse, error) {
	dr, err := ParseDateRange(fromParam, toParam, 30)
	if err != nil {
		return nil, err
	}

	totalUsers, totalEmployees, approved, activeSubs, mrr, err := s.store.AdminPlatformTotals(ctx)
	if err != nil {
		return nil, err
	}
	bookingVolume, err := s.store.AdminBookingVolume(ctx, dr)
	if err != nil {
		return nil, err
	}
	failedPayments, err := s.store.AdminFailedPayments(ctx, dr)
	if err != nil {
		return nil, err
	}

	return &AdminSummaryResponse{
		Period:                  dr.ToResponse(),
		TotalUsers:              totalUsers,
		TotalEmployees:          totalEmployees,
		ApprovedEmployees:       approved,
		ActiveSubscriptions:     activeSubs,
		MonthlyRecurringRevenue: mrr,
		BookingVolume:           bookingVolume,
		FailedPayments:          failedPayments,
		ChurnRate:               nil,
	}, nil
}

// AdminRevenue returns revenue analytics for admins.
func (s *Service) AdminRevenue(ctx context.Context, fromParam, toParam string) (*AdminRevenueResponse, error) {
	dr, err := ParseDateRange(fromParam, toParam, 30)
	if err != nil {
		return nil, err
	}

	subscriptionRevenue, err := s.store.AdminSubscriptionRevenue(ctx, dr)
	if err != nil {
		return nil, err
	}
	bookingRevenue, err := s.store.AdminBookingRevenue(ctx, dr)
	if err != nil {
		return nil, err
	}
	daily, err := s.store.AdminSubscriptionRevenueDaily(ctx, dr)
	if err != nil {
		return nil, err
	}

	return &AdminRevenueResponse{
		Period:              dr.ToResponse(),
		SubscriptionRevenue: subscriptionRevenue,
		BookingRevenue:      bookingRevenue,
		DailySubscription:   toDailyRevenueItems(daily),
	}, nil
}

// AdminBookings returns platform booking analytics.
func (s *Service) AdminBookings(ctx context.Context, fromParam, toParam string) (*AdminBookingsResponse, error) {
	dr, err := ParseDateRange(fromParam, toParam, 30)
	if err != nil {
		return nil, err
	}

	byStatus, err := s.store.AdminBookingsByStatus(ctx, dr)
	if err != nil {
		return nil, err
	}
	daily, err := s.store.AdminBookingsDaily(ctx, dr)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, item := range byStatus {
		total += item.Count
	}

	return &AdminBookingsResponse{
		Period:      dr.ToResponse(),
		Total:       total,
		ByStatus:    toStatusItems(byStatus),
		DailyVolume: toDailyItems(daily),
	}, nil
}

// AdminCategories returns popular categories by booking volume.
func (s *Service) AdminCategories(ctx context.Context, fromParam, toParam string) (*AdminCategoriesResponse, error) {
	dr, err := ParseDateRange(fromParam, toParam, 30)
	if err != nil {
		return nil, err
	}

	categories, err := s.store.AdminPopularCategories(ctx, dr, 10)
	if err != nil {
		return nil, err
	}

	return &AdminCategoriesResponse{
		Period:     dr.ToResponse(),
		Categories: toCategoryItems(categories),
	}, nil
}

func (s *Service) employeeRange(ctx context.Context, userID uuid.UUID, fromParam, toParam string) (*employees.Profile, DateRange, error) {
	dr, err := ParseDateRange(fromParam, toParam, 30)
	if err != nil {
		return nil, DateRange{}, err
	}
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, employees.ErrNotFound) {
			return nil, DateRange{}, ErrNotFound
		}
		return nil, DateRange{}, fmt.Errorf("load employee profile: %w", err)
	}
	return profile, dr, nil
}

func toRatingPoints(items []RatingPeriod) []RatingPoint {
	out := make([]RatingPoint, 0, len(items))
	for _, item := range items {
		out = append(out, RatingPoint{
			Period:        item.Period.Format("2006-01-02"),
			AverageRating: item.AverageRating,
			ReviewCount:   item.ReviewCount,
		})
	}
	return out
}

func toStatusItems(items []StatusCount) []StatusCountItem {
	out := make([]StatusCountItem, 0, len(items))
	for _, item := range items {
		out = append(out, StatusCountItem{Status: item.Status, Count: item.Count})
	}
	return out
}

func toDailyItems(items []BookingDayCount) []DailyCountItem {
	out := make([]DailyCountItem, 0, len(items))
	for _, item := range items {
		out = append(out, DailyCountItem{
			Date:  item.Date.Format("2006-01-02"),
			Count: item.Count,
		})
	}
	return out
}

func toDailyRevenueItems(items []RevenueDay) []DailyRevenueItem {
	out := make([]DailyRevenueItem, 0, len(items))
	for _, item := range items {
		out = append(out, DailyRevenueItem{
			Date:   item.Date.Format("2006-01-02"),
			Amount: item.Amount,
		})
	}
	return out
}

func toCategoryItems(items []CategoryBookingCount) []CategoryCountItem {
	out := make([]CategoryCountItem, 0, len(items))
	for _, item := range items {
		out = append(out, CategoryCountItem{
			CategoryID:   item.CategoryID.String(),
			CategoryName: item.CategoryName,
			BookingCount: item.BookingCount,
		})
	}
	return out
}

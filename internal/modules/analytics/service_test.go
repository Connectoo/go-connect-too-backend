package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

type mockAnalyticsStore struct {
	profileViews int
}

func (m *mockAnalyticsStore) RecordProfileView(context.Context, uuid.UUID, time.Time) error {
	return nil
}

func (m *mockAnalyticsStore) EmployeeProfileViews(context.Context, uuid.UUID, DateRange) (int, error) {
	return m.profileViews, nil
}

func (m *mockAnalyticsStore) EmployeeBookingCounts(context.Context, uuid.UUID, DateRange) (int, int, int, error) {
	return 10, 7, 2, nil
}

func (m *mockAnalyticsStore) EmployeeEstimatedRevenue(context.Context, uuid.UUID, DateRange) (string, error) {
	return "1500.00", nil
}

func (m *mockAnalyticsStore) EmployeeRatingTrend(context.Context, uuid.UUID, DateRange) ([]RatingPeriod, error) {
	return []RatingPeriod{
		{Period: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), AverageRating: 4.5, ReviewCount: 2},
	}, nil
}

func (m *mockAnalyticsStore) EmployeeBookingsByStatus(context.Context, uuid.UUID, DateRange) ([]StatusCount, error) {
	return []StatusCount{{Status: "completed", Count: 7}}, nil
}

func (m *mockAnalyticsStore) EmployeeBookingsDaily(context.Context, uuid.UUID, DateRange) ([]BookingDayCount, error) {
	return []BookingDayCount{{Date: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Count: 3}}, nil
}

func (m *mockAnalyticsStore) EmployeeReviewSummary(context.Context, uuid.UUID, DateRange) (*float64, int, error) {
	avg := 4.5
	return &avg, 2, nil
}

func (m *mockAnalyticsStore) EmployeeRatingDistribution(context.Context, uuid.UUID, DateRange) (map[int]int, error) {
	return map[int]int{1: 0, 2: 0, 3: 0, 4: 1, 5: 1}, nil
}

func (m *mockAnalyticsStore) AdminPlatformTotals(context.Context) (int, int, int, int, int64, error) {
	return 100, 40, 30, 12, 49900, nil
}

func (m *mockAnalyticsStore) AdminBookingVolume(context.Context, DateRange) (int, error) {
	return 25, nil
}

func (m *mockAnalyticsStore) AdminFailedPayments(context.Context, DateRange) (int, error) {
	return 1, nil
}

func (m *mockAnalyticsStore) AdminSubscriptionRevenue(context.Context, DateRange) (int64, error) {
	return 99900, nil
}

func (m *mockAnalyticsStore) AdminBookingRevenue(context.Context, DateRange) (string, error) {
	return "12000.00", nil
}

func (m *mockAnalyticsStore) AdminSubscriptionRevenueDaily(context.Context, DateRange) ([]RevenueDay, error) {
	return []RevenueDay{{Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Amount: 49900}}, nil
}

func (m *mockAnalyticsStore) AdminBookingsByStatus(context.Context, DateRange) ([]StatusCount, error) {
	return []StatusCount{{Status: "completed", Count: 20}}, nil
}

func (m *mockAnalyticsStore) AdminBookingsDaily(context.Context, DateRange) ([]BookingDayCount, error) {
	return []BookingDayCount{{Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Count: 5}}, nil
}

func (m *mockAnalyticsStore) AdminPopularCategories(context.Context, DateRange, int) ([]CategoryBookingCount, error) {
	return []CategoryBookingCount{
		{CategoryID: uuid.New(), CategoryName: "Cleaning", BookingCount: 8},
	}, nil
}

type mockEmployeeLookup struct {
	profile *employees.Profile
}

func (m mockEmployeeLookup) GetByUserID(context.Context, uuid.UUID) (*employees.Profile, error) {
	if m.profile == nil {
		return nil, employees.ErrNotFound
	}
	return m.profile, nil
}

func TestEmployeeSummary(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()
	svc := NewService(&mockAnalyticsStore{profileViews: 42}, mockEmployeeLookup{
		profile: &employees.Profile{ID: profileID, UserID: userID},
	})

	res, err := svc.EmployeeSummary(context.Background(), userID, "2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ProfileViews != 42 || res.TotalBookings != 10 || res.EstimatedRevenue != "1500.00" {
		t.Fatalf("unexpected summary: %+v", res)
	}
	if res.AverageResponseTimeMs != nil {
		t.Fatal("expected nil average response time placeholder")
	}
}

func TestEmployeeSummaryNotFound(t *testing.T) {
	svc := NewService(&mockAnalyticsStore{}, mockEmployeeLookup{})
	_, err := svc.EmployeeSummary(context.Background(), uuid.New(), "", "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAdminSummary(t *testing.T) {
	svc := NewService(&mockAnalyticsStore{}, mockEmployeeLookup{})
	res, err := svc.AdminSummary(context.Background(), "2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalUsers != 100 || res.MonthlyRecurringRevenue != 49900 || res.BookingVolume != 25 {
		t.Fatalf("unexpected admin summary: %+v", res)
	}
	if res.ChurnRate != nil {
		t.Fatal("expected nil churn rate placeholder")
	}
}

func TestParseDateRangeInvalidRejectedByService(t *testing.T) {
	svc := NewService(&mockAnalyticsStore{}, mockEmployeeLookup{
		profile: &employees.Profile{ID: uuid.New(), UserID: uuid.New()},
	})
	_, err := svc.AdminSummary(context.Background(), "bad", "2026-01-01")
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("expected ErrInvalidDateRange, got %v", err)
	}
}

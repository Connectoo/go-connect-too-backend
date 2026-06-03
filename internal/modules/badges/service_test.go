package badges

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockBadgeStore struct {
	items map[string]*Badge
}

func newMockBadgeStore() *mockBadgeStore {
	return &mockBadgeStore{items: make(map[string]*Badge)}
}

func key(employeeID uuid.UUID, badgeType string) string {
	return employeeID.String() + ":" + badgeType
}

func (m *mockBadgeStore) AwardIfNotExists(_ context.Context, employeeID uuid.UUID, badgeType string, at time.Time) (*Badge, error) {
	k := key(employeeID, badgeType)
	if existing, ok := m.items[k]; ok {
		copy := *existing
		return &copy, nil
	}
	badge := &Badge{
		ID:         uuid.New(),
		EmployeeID: employeeID,
		BadgeType:  badgeType,
		CreatedAt:  at,
	}
	m.items[k] = badge
	copy := *badge
	return &copy, nil
}

func (m *mockBadgeStore) ListByEmployeeID(_ context.Context, employeeID uuid.UUID) ([]Badge, error) {
	items := make([]Badge, 0)
	for _, badge := range m.items {
		if badge.EmployeeID == employeeID {
			items = append(items, *badge)
		}
	}
	return items, nil
}

func TestAwardVerifiedBookingReviewIdempotent(t *testing.T) {
	store := newMockBadgeStore()
	svc := NewService(store)
	svc.now = func() time.Time { return time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC) }
	employeeID := uuid.New()

	if err := svc.AwardVerifiedBookingReview(context.Background(), employeeID); err != nil {
		t.Fatalf("first award error = %v", err)
	}
	if err := svc.AwardVerifiedBookingReview(context.Background(), employeeID); err != nil {
		t.Fatalf("second award error = %v", err)
	}

	badges, err := svc.ListForEmployee(context.Background(), employeeID)
	if err != nil {
		t.Fatalf("ListForEmployee() error = %v", err)
	}
	if len(badges) != 1 || badges[0] != TypeVerifiedBookingReview {
		t.Fatalf("unexpected badges: %+v", badges)
	}
}

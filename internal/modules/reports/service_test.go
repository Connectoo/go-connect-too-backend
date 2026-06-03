package reports

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/bookings"
)

type mockBookingReader struct {
	bookings map[uuid.UUID]*bookings.Booking
}

func (m *mockBookingReader) GetByID(_ context.Context, bookingID uuid.UUID) (*bookings.Booking, error) {
	booking, ok := m.bookings[bookingID]
	if !ok {
		return nil, bookings.ErrNotFound
	}
	copy := *booking
	return &copy, nil
}

type mockReportStore struct {
	reports map[uuid.UUID]*Report
}

func newMockReportStore() *mockReportStore {
	return &mockReportStore{reports: make(map[uuid.UUID]*Report)}
}

func (m *mockReportStore) Create(_ context.Context, report *Report) (*Report, error) {
	copy := *report
	m.reports[report.ID] = &copy
	return &copy, nil
}

func (m *mockReportStore) GetByID(_ context.Context, id uuid.UUID) (*Report, error) {
	report, ok := m.reports[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *report
	return &copy, nil
}

func (m *mockReportStore) ListAdmin(_ context.Context, _ string, _, _ int) ([]Report, int, error) {
	items := make([]Report, 0, len(m.reports))
	for _, report := range m.reports {
		items = append(items, *report)
	}
	return items, len(items), nil
}

func (m *mockReportStore) Resolve(_ context.Context, id uuid.UUID, at time.Time) (*Report, error) {
	report, ok := m.reports[id]
	if !ok {
		return nil, ErrNotFound
	}
	updated := *report
	updated.Status = StatusResolved
	updated.UpdatedAt = at
	m.reports[id] = &updated
	return &updated, nil
}

func TestCreateValidation(t *testing.T) {
	svc := NewService(nil, newMockReportStore())
	reporterID := uuid.New()

	_, err := svc.Create(context.Background(), reporterID, CreateReportRequest{
		ReportedUserID: reporterID,
		Reason:         "spam",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want validation for self-report", err)
	}
}

func TestCreateSuccess(t *testing.T) {
	store := newMockReportStore()
	svc := NewService(&mockBookingReader{}, store)
	reporterID := uuid.New()
	reportedID := uuid.New()

	res, err := svc.Create(context.Background(), reporterID, CreateReportRequest{
		ReportedUserID: reportedID,
		Reason:         "harassment",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if res.Status != StatusOpen || res.Reason != "harassment" {
		t.Fatalf("unexpected report: %+v", res)
	}
}

func TestResolveSuccess(t *testing.T) {
	store := newMockReportStore()
	reportID := uuid.New()
	store.reports[reportID] = &Report{ID: reportID, Status: StatusOpen}

	svc := NewService(nil, store)
	res, err := svc.Resolve(context.Background(), reportID)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Status != StatusResolved {
		t.Fatalf("status = %q, want %q", res.Status, StatusResolved)
	}
}

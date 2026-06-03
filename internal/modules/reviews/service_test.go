package reviews

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/bookings"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

type mockCustomerStore struct {
	profile *customers.Profile
}

func (m *mockCustomerStore) GetByUserID(_ context.Context, userID uuid.UUID) (*customers.Profile, error) {
	if m.profile == nil || m.profile.UserID != userID {
		return nil, customers.ErrNotFound
	}
	copy := *m.profile
	return &copy, nil
}

type mockEmployeeStore struct {
	byUserID map[uuid.UUID]*employees.Profile
	byID     map[uuid.UUID]*employees.Profile
}

func (m *mockEmployeeStore) GetByUserID(_ context.Context, userID uuid.UUID) (*employees.Profile, error) {
	profile, ok := m.byUserID[userID]
	if !ok {
		return nil, employees.ErrNotFound
	}
	copy := *profile
	return &copy, nil
}

func (m *mockEmployeeStore) GetApprovedByID(_ context.Context, id uuid.UUID) (*employees.Profile, error) {
	profile, ok := m.byID[id]
	if !ok || profile.VerificationStatus != employees.VerificationApproved {
		return nil, employees.ErrNotFound
	}
	copy := *profile
	return &copy, nil
}

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

type mockReviewStore struct {
	reviews   map[uuid.UUID]*Review
	byBooking map[uuid.UUID]*Review
	replies   map[uuid.UUID]*Reply
}

func newMockReviewStore() *mockReviewStore {
	return &mockReviewStore{
		reviews:   make(map[uuid.UUID]*Review),
		byBooking: make(map[uuid.UUID]*Review),
		replies:   make(map[uuid.UUID]*Reply),
	}
}

func (m *mockReviewStore) Create(_ context.Context, review *Review) (*Review, error) {
	copy := *review
	m.reviews[review.ID] = &copy
	m.byBooking[review.BookingID] = &copy
	return &copy, nil
}

func (m *mockReviewStore) GetByID(_ context.Context, id uuid.UUID) (*Review, error) {
	review, ok := m.reviews[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *review
	return &copy, nil
}

func (m *mockReviewStore) GetByBookingID(_ context.Context, bookingID uuid.UUID) (*Review, error) {
	review, ok := m.byBooking[bookingID]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *review
	return &copy, nil
}

func (m *mockReviewStore) ListByEmployeeID(_ context.Context, employeeID uuid.UUID, status string, _, _ int) ([]Review, int, error) {
	items := make([]Review, 0)
	for _, review := range m.reviews {
		if review.EmployeeID != employeeID {
			continue
		}
		if status != "" && review.Status != status {
			continue
		}
		items = append(items, *review)
	}
	return items, len(items), nil
}

func (m *mockReviewStore) ListAdmin(_ context.Context, _ string, _, _ int) ([]Review, int, error) {
	items := make([]Review, 0, len(m.reviews))
	for _, review := range m.reviews {
		items = append(items, *review)
	}
	return items, len(items), nil
}

func (m *mockReviewStore) UpdateStatus(_ context.Context, id uuid.UUID, status string, at time.Time) (*Review, error) {
	review, ok := m.reviews[id]
	if !ok {
		return nil, ErrNotFound
	}
	updated := *review
	updated.Status = status
	updated.UpdatedAt = at
	m.reviews[id] = &updated
	m.byBooking[updated.BookingID] = &updated
	return &updated, nil
}

func (m *mockReviewStore) CreateReply(_ context.Context, reply *Reply) (*Reply, error) {
	if _, ok := m.replies[reply.ReviewID]; ok {
		return nil, ErrReplyAlreadyExists
	}
	copy := *reply
	m.replies[reply.ReviewID] = &copy
	return &copy, nil
}

func (m *mockReviewStore) GetReplyByReviewID(_ context.Context, reviewID uuid.UUID) (*Reply, error) {
	reply, ok := m.replies[reviewID]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *reply
	return &copy, nil
}

func (m *mockReviewStore) ListRepliesByReviewIDs(_ context.Context, reviewIDs []uuid.UUID) (map[uuid.UUID]Reply, error) {
	out := make(map[uuid.UUID]Reply)
	for _, id := range reviewIDs {
		if reply, ok := m.replies[id]; ok {
			out[id] = *reply
		}
	}
	return out, nil
}

type mockBadgeAwarder struct {
	awarded []uuid.UUID
}

func (m *mockBadgeAwarder) AwardVerifiedBookingReview(_ context.Context, employeeID uuid.UUID) error {
	m.awarded = append(m.awarded, employeeID)
	return nil
}

type mockRatingRefresher struct {
	refreshed []uuid.UUID
}

func (m *mockRatingRefresher) RefreshEmployeeRating(_ context.Context, employeeID uuid.UUID) error {
	m.refreshed = append(m.refreshed, employeeID)
	return nil
}

func newTestService(t *testing.T, store Store, bookings BookingReader) (*Service, *mockBadgeAwarder, *mockRatingRefresher) {
	t.Helper()
	badgeMock := &mockBadgeAwarder{}
	ratingMock := &mockRatingRefresher{}
	svc := NewService(
		&mockCustomerStore{profile: &customers.Profile{ID: uuid.New(), UserID: uuid.New()}},
		&mockEmployeeStore{byUserID: map[uuid.UUID]*employees.Profile{}, byID: map[uuid.UUID]*employees.Profile{}},
		bookings,
		store,
		badgeMock,
		ratingMock,
	)
	svc.now = func() time.Time { return time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC) }
	return svc, badgeMock, ratingMock
}

func TestCreateForBookingRequiresCompletedStatus(t *testing.T) {
	userID := uuid.New()
	customerID := uuid.New()
	bookingID := uuid.New()
	employeeID := uuid.New()

	store := newMockReviewStore()
	svc := NewService(
		&mockCustomerStore{profile: &customers.Profile{ID: customerID, UserID: userID}},
		nil,
		&mockBookingReader{bookings: map[uuid.UUID]*bookings.Booking{
			bookingID: {
				ID:         bookingID,
				CustomerID: customerID,
				EmployeeID: employeeID,
				Status:     bookings.StatusAccepted,
			},
		}},
		store,
		nil,
		nil,
	)

	_, err := svc.CreateForBooking(context.Background(), userID, bookingID, CreateReviewRequest{Rating: 5})
	if !errors.Is(err, ErrBookingNotCompleted) {
		t.Fatalf("error = %v, want %v", err, ErrBookingNotCompleted)
	}
}

func TestCreateForBookingSuccess(t *testing.T) {
	userID := uuid.New()
	customerID := uuid.New()
	bookingID := uuid.New()
	employeeID := uuid.New()

	store := newMockReviewStore()
	badgeMock := &mockBadgeAwarder{}
	svc := NewService(
		&mockCustomerStore{profile: &customers.Profile{ID: customerID, UserID: userID}},
		nil,
		&mockBookingReader{bookings: map[uuid.UUID]*bookings.Booking{
			bookingID: {
				ID:         bookingID,
				CustomerID: customerID,
				EmployeeID: employeeID,
				Status:     bookings.StatusCompleted,
			},
		}},
		store,
		badgeMock,
		nil,
	)

	res, err := svc.CreateForBooking(context.Background(), userID, bookingID, CreateReviewRequest{Rating: 4})
	if err != nil {
		t.Fatalf("CreateForBooking() error = %v", err)
	}
	if res.Status != StatusPending || res.Rating != 4 {
		t.Fatalf("unexpected review: %+v", res)
	}
	if len(badgeMock.awarded) != 1 || badgeMock.awarded[0] != employeeID {
		t.Fatalf("badge not awarded: %+v", badgeMock.awarded)
	}
}

func TestApproveRefreshesRating(t *testing.T) {
	store := newMockReviewStore()
	reviewID := uuid.New()
	employeeID := uuid.New()
	store.reviews[reviewID] = &Review{
		ID:         reviewID,
		EmployeeID: employeeID,
		Status:     StatusPending,
		Rating:     5,
	}

	ratingMock := &mockRatingRefresher{}
	svc := NewService(nil, nil, nil, store, nil, ratingMock)

	res, err := svc.Approve(context.Background(), reviewID)
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if res.Status != StatusApproved {
		t.Fatalf("status = %q, want %q", res.Status, StatusApproved)
	}
	if len(ratingMock.refreshed) != 1 || ratingMock.refreshed[0] != employeeID {
		t.Fatalf("rating not refreshed: %+v", ratingMock.refreshed)
	}
}

func TestReplyForbiddenForOtherEmployee(t *testing.T) {
	store := newMockReviewStore()
	reviewID := uuid.New()
	employeeUserID := uuid.New()
	otherEmployeeID := uuid.New()
	store.reviews[reviewID] = &Review{ID: reviewID, EmployeeID: otherEmployeeID}

	svc := NewService(nil, &mockEmployeeStore{
		byUserID: map[uuid.UUID]*employees.Profile{
			employeeUserID: {ID: uuid.New(), UserID: employeeUserID},
		},
	}, nil, store, nil, nil)

	_, err := svc.Reply(context.Background(), employeeUserID, reviewID, ReplyRequest{Reply: "Thanks"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("error = %v, want %v", err, ErrForbidden)
	}
}

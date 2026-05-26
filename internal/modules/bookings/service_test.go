package bookings

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/availability"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/services"
)

type mockCustomerStore struct {
	byUserID map[uuid.UUID]*customers.Profile
}

func (m *mockCustomerStore) GetByUserID(_ context.Context, userID uuid.UUID) (*customers.Profile, error) {
	profile, ok := m.byUserID[userID]
	if !ok {
		return nil, customers.ErrNotFound
	}
	copy := *profile
	return &copy, nil
}

type mockEmployeeStore struct {
	byUserID   map[uuid.UUID]*employees.Profile
	approvedID map[uuid.UUID]*employees.Profile
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
	profile, ok := m.approvedID[id]
	if !ok {
		return nil, employees.ErrNotFound
	}
	copy := *profile
	return &copy, nil
}

type mockServiceCatalog struct {
	byID map[uuid.UUID]*services.EmployeeService
}

func (m *mockServiceCatalog) GetPublicActiveByID(_ context.Context, serviceID uuid.UUID) (*services.EmployeeService, error) {
	service, ok := m.byID[serviceID]
	if !ok {
		return nil, services.ErrNotFound
	}
	copy := *service
	return &copy, nil
}

type mockBookingStore struct {
	bookings    map[uuid.UUID]*Booking
	available   bool
	overlapFail bool
}

func newMockBookingStore() *mockBookingStore {
	return &mockBookingStore{
		bookings:  make(map[uuid.UUID]*Booking),
		available: true,
	}
}

func (m *mockBookingStore) Create(_ context.Context, booking *Booking, _ uuid.UUID) (*Booking, error) {
	if m.overlapFail {
		return nil, ErrDoubleBooking
	}
	for _, existing := range m.bookings {
		if existing.EmployeeID == booking.EmployeeID &&
			existing.BookingDate.Equal(booking.BookingDate) &&
			existing.Status != StatusRejected &&
			existing.Status != StatusCancelled &&
			existing.Status != StatusCompleted &&
			existing.StartTime.Before(booking.EndTime) &&
			booking.StartTime.Before(existing.EndTime) {
			return nil, ErrDoubleBooking
		}
	}
	copy := *booking
	m.bookings[booking.ID] = &copy
	return &copy, nil
}

func (m *mockBookingStore) UpdateStatus(_ context.Context, bookingID uuid.UUID, newStatus string, _ uuid.UUID, reason, employeeNotes *string, at time.Time) (*Booking, error) {
	booking, ok := m.bookings[bookingID]
	if !ok {
		return nil, ErrNotFound
	}
	updated := *booking
	updated.Status = newStatus
	updated.UpdatedAt = at
	if employeeNotes != nil {
		updated.EmployeeNotes = employeeNotes
	}
	if reason != nil {
		_ = reason
	}
	m.bookings[bookingID] = &updated
	return &updated, nil
}

func (m *mockBookingStore) GetByID(_ context.Context, bookingID uuid.UUID) (*Booking, error) {
	booking, ok := m.bookings[bookingID]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *booking
	return &copy, nil
}

func (m *mockBookingStore) ListByCustomerID(context.Context, uuid.UUID) ([]Booking, error) {
	return []Booking{}, nil
}

func (m *mockBookingStore) ListByEmployeeID(context.Context, uuid.UUID) ([]Booking, error) {
	return []Booking{}, nil
}

func (m *mockBookingStore) ListAll(context.Context) ([]Booking, error) {
	return []Booking{}, nil
}

func (m *mockBookingStore) EmployeeIsAvailable(context.Context, uuid.UUID, int, availability.TimeOfDay, availability.TimeOfDay) (bool, error) {
	return m.available, nil
}

func newTestService(t *testing.T, customers CustomerProfileStore, employees EmployeeProfileStore, catalog ServiceCatalog, store Store) *Service {
	t.Helper()
	svc := NewService(customers, employees, catalog, store, NoopEventPublisher{})
	svc.now = func() time.Time {
		return time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	}
	return svc
}

func TestCreateBookingSuccess(t *testing.T) {
	customerUserID := uuid.New()
	customerID := uuid.New()
	employeeID := uuid.New()
	serviceID := uuid.New()

	svc := newTestService(t,
		&mockCustomerStore{byUserID: map[uuid.UUID]*customers.Profile{
			customerUserID: {ID: customerID, UserID: customerUserID},
		}},
		&mockEmployeeStore{approvedID: map[uuid.UUID]*employees.Profile{
			employeeID: {ID: employeeID, VerificationStatus: employees.VerificationApproved},
		}},
		&mockServiceCatalog{byID: map[uuid.UUID]*services.EmployeeService{
			serviceID: {ID: serviceID, EmployeeID: employeeID, Price: 49.99, IsActive: true},
		}},
		newMockBookingStore(),
	)

	res, err := svc.Create(context.Background(), customerUserID, CreateBookingRequest{
		ServiceID:   serviceID,
		BookingDate: "2026-05-28",
		StartTime:   availability.TimeOfDay{Hour: 10, Minute: 0},
		EndTime:     availability.TimeOfDay{Hour: 11, Minute: 0},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if res.Status != StatusPending || res.CustomerID != customerID || res.TotalAmount != 49.99 {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestDoubleBookingPrevention(t *testing.T) {
	customerUserID := uuid.New()
	customerID := uuid.New()
	employeeID := uuid.New()
	serviceID := uuid.New()
	store := newMockBookingStore()

	svc := newTestService(t,
		&mockCustomerStore{byUserID: map[uuid.UUID]*customers.Profile{
			customerUserID: {ID: customerID, UserID: customerUserID},
		}},
		&mockEmployeeStore{approvedID: map[uuid.UUID]*employees.Profile{
			employeeID: {ID: employeeID, VerificationStatus: employees.VerificationApproved},
		}},
		&mockServiceCatalog{byID: map[uuid.UUID]*services.EmployeeService{
			serviceID: {ID: serviceID, EmployeeID: employeeID, Price: 25, IsActive: true},
		}},
		store,
	)

	req := CreateBookingRequest{
		ServiceID:   serviceID,
		BookingDate: "2026-05-28",
		StartTime:   availability.TimeOfDay{Hour: 10, Minute: 0},
		EndTime:     availability.TimeOfDay{Hour: 11, Minute: 0},
	}
	if _, err := svc.Create(context.Background(), customerUserID, req); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err := svc.Create(context.Background(), customerUserID, req)
	if !errors.Is(err, ErrDoubleBooking) {
		t.Fatalf("second Create() error = %v, want %v", err, ErrDoubleBooking)
	}
}

func TestAcceptBooking(t *testing.T) {
	employeeUserID := uuid.New()
	employeeID := uuid.New()
	bookingID := uuid.New()

	store := newMockBookingStore()
	store.bookings[bookingID] = &Booking{
		ID:         bookingID,
		EmployeeID: employeeID,
		CustomerID: uuid.New(),
		Status:     StatusPending,
	}

	svc := newTestService(t,
		&mockCustomerStore{byUserID: map[uuid.UUID]*customers.Profile{}},
		&mockEmployeeStore{byUserID: map[uuid.UUID]*employees.Profile{
			employeeUserID: {ID: employeeID, UserID: employeeUserID},
		}},
		&mockServiceCatalog{byID: map[uuid.UUID]*services.EmployeeService{}},
		store,
	)

	res, err := svc.Accept(context.Background(), employeeUserID, bookingID, EmployeeActionRequest{})
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}
	if res.Status != StatusAccepted {
		t.Fatalf("status = %q, want %q", res.Status, StatusAccepted)
	}
}

func TestRejectBooking(t *testing.T) {
	employeeUserID := uuid.New()
	employeeID := uuid.New()
	bookingID := uuid.New()

	store := newMockBookingStore()
	store.bookings[bookingID] = &Booking{
		ID:         bookingID,
		EmployeeID: employeeID,
		Status:     StatusPending,
	}

	svc := newTestService(t,
		&mockCustomerStore{byUserID: map[uuid.UUID]*customers.Profile{}},
		&mockEmployeeStore{byUserID: map[uuid.UUID]*employees.Profile{
			employeeUserID: {ID: employeeID, UserID: employeeUserID},
		}},
		&mockServiceCatalog{byID: map[uuid.UUID]*services.EmployeeService{}},
		store,
	)

	res, err := svc.Reject(context.Background(), employeeUserID, bookingID, EmployeeActionRequest{Reason: strPtr("schedule conflict")})
	if err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if res.Status != StatusRejected {
		t.Fatalf("status = %q, want %q", res.Status, StatusRejected)
	}
}

func TestInvalidStatusTransition(t *testing.T) {
	employeeUserID := uuid.New()
	employeeID := uuid.New()
	bookingID := uuid.New()

	store := newMockBookingStore()
	store.bookings[bookingID] = &Booking{
		ID:         bookingID,
		EmployeeID: employeeID,
		Status:     StatusPending,
	}

	svc := newTestService(t,
		&mockCustomerStore{byUserID: map[uuid.UUID]*customers.Profile{}},
		&mockEmployeeStore{byUserID: map[uuid.UUID]*employees.Profile{
			employeeUserID: {ID: employeeID, UserID: employeeUserID},
		}},
		&mockServiceCatalog{byID: map[uuid.UUID]*services.EmployeeService{}},
		store,
	)

	_, err := svc.Complete(context.Background(), employeeUserID, bookingID, EmployeeActionRequest{})
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("Complete() error = %v, want %v", err, ErrInvalidStatusTransition)
	}
}

func TestUnauthorizedCustomerCancel(t *testing.T) {
	ownerUserID := uuid.New()
	otherUserID := uuid.New()
	customerID := uuid.New()
	bookingID := uuid.New()

	store := newMockBookingStore()
	store.bookings[bookingID] = &Booking{
		ID:         bookingID,
		CustomerID: customerID,
		Status:     StatusPending,
	}

	svc := newTestService(t,
		&mockCustomerStore{byUserID: map[uuid.UUID]*customers.Profile{
			ownerUserID: {ID: customerID, UserID: ownerUserID},
			otherUserID: {ID: uuid.New(), UserID: otherUserID},
		}},
		&mockEmployeeStore{},
		&mockServiceCatalog{},
		store,
	)

	_, err := svc.Cancel(context.Background(), otherUserID, bookingID, CancelBookingRequest{})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Cancel() error = %v, want %v", err, ErrForbidden)
	}
}

func TestUnauthorizedEmployeeAccept(t *testing.T) {
	ownerUserID := uuid.New()
	otherUserID := uuid.New()
	employeeID := uuid.New()
	bookingID := uuid.New()

	store := newMockBookingStore()
	store.bookings[bookingID] = &Booking{
		ID:         bookingID,
		EmployeeID: employeeID,
		Status:     StatusPending,
	}

	svc := newTestService(t,
		&mockCustomerStore{},
		&mockEmployeeStore{byUserID: map[uuid.UUID]*employees.Profile{
			ownerUserID: {ID: employeeID, UserID: ownerUserID},
			otherUserID: {ID: uuid.New(), UserID: otherUserID},
		}},
		&mockServiceCatalog{},
		store,
	)

	_, err := svc.Accept(context.Background(), otherUserID, bookingID, EmployeeActionRequest{})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Accept() error = %v, want %v", err, ErrForbidden)
	}
}

func strPtr(value string) *string {
	return &value
}

package bookings

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/availability"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/services"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

const maxNotesLength = 1000

// CustomerProfileStore resolves authenticated users to customer profiles.
type CustomerProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*customers.Profile, error)
}

// EmployeeProfileStore resolves authenticated users to employee profiles.
type EmployeeProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*employees.Profile, error)
	GetApprovedByID(ctx context.Context, id uuid.UUID) (*employees.Profile, error)
}

// ServiceCatalog loads bookable services.
type ServiceCatalog interface {
	GetPublicActiveByID(ctx context.Context, serviceID uuid.UUID) (*services.EmployeeService, error)
}

// AdminStore supports paginated admin booking listings.
type AdminStore interface {
	ListAdmin(ctx context.Context, filter AdminListFilter) ([]Booking, int, error)
}

// Store persists bookings.
type Store interface {
	Create(ctx context.Context, booking *Booking, changedByUserID uuid.UUID) (*Booking, error)
	UpdateStatus(ctx context.Context, bookingID uuid.UUID, newStatus string, changedByUserID uuid.UUID, reason, employeeNotes *string, at time.Time) (*Booking, error)
	GetByID(ctx context.Context, bookingID uuid.UUID) (*Booking, error)
	ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]Booking, error)
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]Booking, error)
	ListAll(ctx context.Context) ([]Booking, error)
	EmployeeIsAvailable(ctx context.Context, employeeID uuid.UUID, dayOfWeek int, start, end availability.TimeOfDay) (bool, error)
}

// Service handles booking business logic.
type Service struct {
	customers CustomerProfileStore
	employees EmployeeProfileStore
	services  ServiceCatalog
	store     Store
	events    EventPublisher
	now       func() time.Time
}

// NewService creates a booking service.
func NewService(
	customers CustomerProfileStore,
	employees EmployeeProfileStore,
	catalog ServiceCatalog,
	store Store,
	events EventPublisher,
) *Service {
	if events == nil {
		events = NoopEventPublisher{}
	}
	return &Service{
		customers: customers,
		employees: employees,
		services:  catalog,
		store:     store,
		events:    events,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// Create creates a booking for the authenticated customer.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateBookingRequest) (*BookingResponse, error) {
	customer, err := s.customers.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return nil, ErrCustomerProfileNotFound
		}
		return nil, err
	}
	return s.createBooking(ctx, userID, customer, req, nil)
}

// RebookPreview returns hints for rebooking a prior booking owned by the customer.
func (s *Service) RebookPreview(ctx context.Context, userID, sourceBookingID uuid.UUID) (*RebookPreviewResponse, error) {
	source, _, err := s.loadRebookSource(ctx, userID, sourceBookingID)
	if err != nil {
		return nil, err
	}

	serviceAvailable := false
	employeeAvailable := false
	var currentPrice *float64

	if service, err := s.services.GetPublicActiveByID(ctx, source.ServiceID); err == nil {
		serviceAvailable = true
		price := service.Price
		currentPrice = &price
		if _, err := s.employees.GetApprovedByID(ctx, service.EmployeeID); err == nil {
			employeeAvailable = true
		}
	}

	return &RebookPreviewResponse{
		SourceBookingID:   source.ID,
		SourceStatus:      source.Status,
		ServiceID:         source.ServiceID,
		EmployeeID:        source.EmployeeID,
		ServiceAvailable:  serviceAvailable,
		EmployeeAvailable: employeeAvailable,
		CanRebook:         serviceAvailable && employeeAvailable,
		CurrentPrice:      currentPrice,
		SuggestedDate:     source.BookingDate.Format("2006-01-02"),
		SuggestedStart:    source.StartTime.String(),
		SuggestedEnd:      source.EndTime.String(),
	}, nil
}

// Rebook creates a new booking from a completed or cancelled prior booking.
func (s *Service) Rebook(ctx context.Context, userID uuid.UUID, req RebookBookingRequest) (*BookingResponse, error) {
	if req.SourceBookingID == uuid.Nil {
		return nil, fmt.Errorf("%w: source_booking_id is required", ErrValidation)
	}

	source, customer, err := s.loadRebookSource(ctx, userID, req.SourceBookingID)
	if err != nil {
		return nil, err
	}

	createReq := CreateBookingRequest{
		ServiceID:     source.ServiceID,
		BookingDate:   req.BookingDate,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		CustomerNotes: req.CustomerNotes,
	}
	return s.createBooking(ctx, userID, customer, createReq, &source.ID)
}

func (s *Service) loadRebookSource(ctx context.Context, userID, sourceBookingID uuid.UUID) (*Booking, *customers.Profile, error) {
	customer, err := s.customers.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return nil, nil, ErrCustomerProfileNotFound
		}
		return nil, nil, err
	}

	source, err := s.store.GetByID(ctx, sourceBookingID)
	if err != nil {
		return nil, nil, err
	}
	if source.CustomerID != customer.ID {
		return nil, nil, ErrForbidden
	}
	if !isRebookEligibleStatus(source.Status) {
		return nil, nil, fmt.Errorf("%w: booking status is %s", ErrRebookNotAllowed, source.Status)
	}
	return source, customer, nil
}

func (s *Service) createBooking(
	ctx context.Context,
	userID uuid.UUID,
	customer *customers.Profile,
	req CreateBookingRequest,
	sourceBookingID *uuid.UUID,
) (*BookingResponse, error) {
	service, err := s.services.GetPublicActiveByID(ctx, req.ServiceID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}

	if _, err := s.employees.GetApprovedByID(ctx, service.EmployeeID); err != nil {
		if errors.Is(err, employees.ErrNotFound) {
			return nil, ErrEmployeeNotApproved
		}
		return nil, err
	}

	bookingDate, start, end, customerNotes, err := validateCreateRequest(req)
	if err != nil {
		return nil, err
	}
	if err := ensureBookingDateNotPast(bookingDate, s.now()); err != nil {
		return nil, err
	}

	available, err := s.store.EmployeeIsAvailable(ctx, service.EmployeeID, int(bookingDate.Weekday()), start, end)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, ErrEmployeeUnavailable
	}

	at := s.now()
	booking := &Booking{
		ID:              uuid.New(),
		CustomerID:      customer.ID,
		EmployeeID:      service.EmployeeID,
		ServiceID:       service.ID,
		BookingDate:     bookingDate,
		StartTime:       start,
		EndTime:         end,
		Status:          StatusPending,
		CustomerNotes:   customerNotes,
		TotalAmount:     service.Price,
		SourceBookingID: sourceBookingID,
		CreatedAt:       at,
		UpdatedAt:       at,
	}

	created, err := s.store.Create(ctx, booking, userID)
	if err != nil {
		if errors.Is(err, ErrDoubleBooking) {
			return nil, ErrDoubleBooking
		}
		return nil, err
	}

	s.events.Publish(ctx, BookingEvent{
		Type:       EventBookingCreated,
		BookingID:  created.ID,
		CustomerID: created.CustomerID,
		EmployeeID: created.EmployeeID,
		Status:     created.Status,
	})

	return toResponse(created), nil
}

func isRebookEligibleStatus(status string) bool {
	switch status {
	case StatusCompleted, StatusCancelled, StatusRejected, StatusNoShow:
		return true
	default:
		return false
	}
}

// ListForCustomer returns bookings owned by the authenticated customer.
func (s *Service) ListForCustomer(ctx context.Context, userID uuid.UUID) ([]BookingResponse, error) {
	customer, err := s.customers.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return nil, ErrCustomerProfileNotFound
		}
		return nil, err
	}

	items, err := s.store.ListByCustomerID(ctx, customer.ID)
	if err != nil {
		return nil, err
	}
	return toResponseList(items), nil
}

// GetForCustomer returns a booking when owned by the authenticated customer.
func (s *Service) GetForCustomer(ctx context.Context, userID, bookingID uuid.UUID) (*BookingResponse, error) {
	customer, err := s.customers.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return nil, ErrCustomerProfileNotFound
		}
		return nil, err
	}

	booking, err := s.store.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.CustomerID != customer.ID {
		return nil, ErrForbidden
	}
	return toResponse(booking), nil
}

// Cancel cancels a booking owned by the authenticated customer.
func (s *Service) Cancel(ctx context.Context, userID, bookingID uuid.UUID, req CancelBookingRequest) (*BookingResponse, error) {
	customer, err := s.customers.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return nil, ErrCustomerProfileNotFound
		}
		return nil, err
	}

	booking, err := s.store.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.CustomerID != customer.ID {
		return nil, ErrForbidden
	}

	reason, err := optionalNotes(req.Reason)
	if err != nil {
		return nil, err
	}

	updated, err := s.transition(ctx, booking, StatusCancelled, ActionCustomerCancel, userID, reason, nil)
	if err != nil {
		return nil, err
	}

	s.events.Publish(ctx, BookingEvent{
		Type:       EventBookingCancelled,
		BookingID:  updated.ID,
		CustomerID: updated.CustomerID,
		EmployeeID: updated.EmployeeID,
		Status:     updated.Status,
	})

	return toResponse(updated), nil
}

// ListForEmployee returns bookings assigned to the authenticated employee.
func (s *Service) ListForEmployee(ctx context.Context, userID uuid.UUID) ([]BookingResponse, error) {
	profile, err := s.employees.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, err := s.store.ListByEmployeeID(ctx, profile.ID)
	if err != nil {
		return nil, err
	}
	return toResponseList(items), nil
}

// Accept marks a pending booking as accepted.
func (s *Service) Accept(ctx context.Context, userID, bookingID uuid.UUID, req EmployeeActionRequest) (*BookingResponse, error) {
	return s.employeeTransition(ctx, userID, bookingID, StatusAccepted, ActionEmployeeAccept, req)
}

// Reject marks a pending booking as rejected.
func (s *Service) Reject(ctx context.Context, userID, bookingID uuid.UUID, req EmployeeActionRequest) (*BookingResponse, error) {
	return s.employeeTransition(ctx, userID, bookingID, StatusRejected, ActionEmployeeReject, req)
}

// Start marks an accepted booking as in progress.
func (s *Service) Start(ctx context.Context, userID, bookingID uuid.UUID, req EmployeeActionRequest) (*BookingResponse, error) {
	return s.employeeTransition(ctx, userID, bookingID, StatusInProgress, ActionEmployeeStart, req)
}

// Complete marks an in-progress booking as completed.
func (s *Service) Complete(ctx context.Context, userID, bookingID uuid.UUID, req EmployeeActionRequest) (*BookingResponse, error) {
	return s.employeeTransition(ctx, userID, bookingID, StatusCompleted, ActionEmployeeComplete, req)
}

// ListForAdmin returns paginated bookings for admin views.
func (s *Service) ListForAdmin(ctx context.Context, status string, page pagination.Params) (pagination.Result[BookingResponse], error) {
	adminStore, ok := s.store.(AdminStore)
	if !ok {
		items, err := s.store.ListAll(ctx)
		if err != nil {
			return pagination.Result[BookingResponse]{}, err
		}
		return pagination.NewResult(toResponseList(items), page, len(items)), nil
	}

	items, total, err := adminStore.ListAdmin(ctx, AdminListFilter{
		Status: status,
		Offset: page.Offset(),
		Limit:  page.Limit,
	})
	if err != nil {
		return pagination.Result[BookingResponse]{}, err
	}
	return pagination.NewResult(toResponseList(items), page, total), nil
}

// GetForAdmin returns a booking by id.
func (s *Service) GetForAdmin(ctx context.Context, bookingID uuid.UUID) (*BookingResponse, error) {
	booking, err := s.store.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	return toResponse(booking), nil
}

func (s *Service) employeeTransition(
	ctx context.Context,
	userID, bookingID uuid.UUID,
	newStatus string,
	action TransitionAction,
	req EmployeeActionRequest,
) (*BookingResponse, error) {
	profile, err := s.employees.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	booking, err := s.store.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.EmployeeID != profile.ID {
		return nil, ErrForbidden
	}

	employeeNotes, err := optionalNotes(req.EmployeeNotes)
	if err != nil {
		return nil, err
	}
	reason, err := optionalNotes(req.Reason)
	if err != nil {
		return nil, err
	}

	updated, err := s.transition(ctx, booking, newStatus, action, userID, reason, employeeNotes)
	if err != nil {
		return nil, err
	}

	s.events.Publish(ctx, employeeEvent(action, updated))
	return toResponse(updated), nil
}

func (s *Service) transition(
	ctx context.Context,
	booking *Booking,
	newStatus string,
	action TransitionAction,
	changedByUserID uuid.UUID,
	reason, employeeNotes *string,
) (*Booking, error) {
	if err := ValidateTransition(booking.Status, newStatus, action); err != nil {
		return nil, err
	}
	return s.store.UpdateStatus(ctx, booking.ID, newStatus, changedByUserID, reason, employeeNotes, s.now())
}

func employeeEvent(action TransitionAction, booking *Booking) BookingEvent {
	switch action {
	case ActionEmployeeAccept:
		return BookingEvent{Type: EventBookingAccepted, BookingID: booking.ID, CustomerID: booking.CustomerID, EmployeeID: booking.EmployeeID, Status: booking.Status}
	case ActionEmployeeReject:
		return BookingEvent{Type: EventBookingRejected, BookingID: booking.ID, CustomerID: booking.CustomerID, EmployeeID: booking.EmployeeID, Status: booking.Status}
	case ActionEmployeeStart:
		return BookingEvent{Type: EventBookingStarted, BookingID: booking.ID, CustomerID: booking.CustomerID, EmployeeID: booking.EmployeeID, Status: booking.Status}
	default:
		return BookingEvent{Type: EventBookingCompleted, BookingID: booking.ID, CustomerID: booking.CustomerID, EmployeeID: booking.EmployeeID, Status: booking.Status}
	}
}

func validateCreateRequest(req CreateBookingRequest) (time.Time, availability.TimeOfDay, availability.TimeOfDay, *string, error) {
	if req.ServiceID == uuid.Nil {
		return time.Time{}, availability.TimeOfDay{}, availability.TimeOfDay{}, nil, fmt.Errorf("%w: service_id is required", ErrValidation)
	}

	bookingDate, err := parseBookingDate(req.BookingDate)
	if err != nil {
		return time.Time{}, availability.TimeOfDay{}, availability.TimeOfDay{}, nil, err
	}
	if !req.StartTime.Before(req.EndTime) {
		return time.Time{}, availability.TimeOfDay{}, availability.TimeOfDay{}, nil, fmt.Errorf("%w: end_time must be after start_time", ErrValidation)
	}

	customerNotes, err := optionalNotes(req.CustomerNotes)
	if err != nil {
		return time.Time{}, availability.TimeOfDay{}, availability.TimeOfDay{}, nil, err
	}

	return bookingDate, req.StartTime, req.EndTime, customerNotes, nil
}

func parseBookingDate(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("%w: booking_date is required", ErrValidation)
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: booking_date must be YYYY-MM-DD", ErrValidation)
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
}

func ensureBookingDateNotPast(bookingDate, now time.Time) error {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if bookingDate.Before(today) {
		return fmt.Errorf("%w: booking_date cannot be in the past", ErrValidation)
	}
	return nil
}

func optionalNotes(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > maxNotesLength {
		return nil, fmt.Errorf("%w: notes exceed %d characters", ErrValidation, maxNotesLength)
	}
	return &trimmed, nil
}

func toResponseList(items []Booking) []BookingResponse {
	out := make([]BookingResponse, 0, len(items))
	for i := range items {
		out = append(out, *toResponse(&items[i]))
	}
	return out
}

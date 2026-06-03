package reports

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/bookings"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

const maxReasonLength = 100
const maxDescriptionLength = 2000

// BookingReader validates optional booking references.
type BookingReader interface {
	GetByID(ctx context.Context, bookingID uuid.UUID) (*bookings.Booking, error)
}

// Store persists reports.
type Store interface {
	Create(ctx context.Context, report *Report) (*Report, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Report, error)
	ListAdmin(ctx context.Context, status string, offset, limit int) ([]Report, int, error)
	Resolve(ctx context.Context, id uuid.UUID, at time.Time) (*Report, error)
}

// Service handles report business logic.
type Service struct {
	bookings BookingReader
	store    Store
	now      func() time.Time
}

// NewService creates a reports service.
func NewService(bookings BookingReader, store Store) *Service {
	return &Service{
		bookings: bookings,
		store:    store,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

// Create submits a report from the authenticated user.
func (s *Service) Create(ctx context.Context, reporterID uuid.UUID, req CreateReportRequest) (*ReportResponse, error) {
	if req.ReportedUserID == uuid.Nil {
		return nil, fmt.Errorf("%w: reported_user_id is required", ErrValidation)
	}
	if req.ReportedUserID == reporterID {
		return nil, fmt.Errorf("%w: cannot report yourself", ErrValidation)
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, fmt.Errorf("%w: reason is required", ErrValidation)
	}
	if utf8.RuneCountInString(reason) > maxReasonLength {
		return nil, fmt.Errorf("%w: reason exceeds %d characters", ErrValidation, maxReasonLength)
	}

	description, err := optionalDescription(req.Description)
	if err != nil {
		return nil, err
	}

	if req.BookingID != nil {
		booking, err := s.bookings.GetByID(ctx, *req.BookingID)
		if err != nil {
			if errors.Is(err, bookings.ErrNotFound) {
				return nil, fmt.Errorf("%w: booking not found", ErrValidation)
			}
			return nil, err
		}
		if booking.CustomerID == uuid.Nil && booking.EmployeeID == uuid.Nil {
			return nil, fmt.Errorf("%w: invalid booking reference", ErrValidation)
		}
	}

	at := s.now()
	report := &Report{
		ID:             uuid.New(),
		ReporterID:     reporterID,
		ReportedUserID: req.ReportedUserID,
		BookingID:      req.BookingID,
		Reason:         reason,
		Description:    description,
		Status:         StatusOpen,
		CreatedAt:      at,
		UpdatedAt:      at,
	}

	created, err := s.store.Create(ctx, report)
	if err != nil {
		return nil, err
	}
	return toResponse(created), nil
}

// ListForAdmin returns paginated reports for moderation.
func (s *Service) ListForAdmin(ctx context.Context, status string, page pagination.Params) (pagination.Result[ReportResponse], error) {
	items, total, err := s.store.ListAdmin(ctx, status, page.Offset(), page.Limit)
	if err != nil {
		return pagination.Result[ReportResponse]{}, err
	}

	out := make([]ReportResponse, 0, len(items))
	for i := range items {
		out = append(out, *toResponse(&items[i]))
	}
	return pagination.NewResult(out, page, total), nil
}

// Resolve marks a report as resolved.
func (s *Service) Resolve(ctx context.Context, reportID uuid.UUID) (*ReportResponse, error) {
	report, err := s.store.GetByID(ctx, reportID)
	if err != nil {
		return nil, err
	}
	if report.Status == StatusResolved {
		return toResponse(report), nil
	}

	updated, err := s.store.Resolve(ctx, reportID, s.now())
	if err != nil {
		return nil, err
	}
	return toResponse(updated), nil
}

func optionalDescription(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > maxDescriptionLength {
		return nil, fmt.Errorf("%w: description exceeds %d characters", ErrValidation, maxDescriptionLength)
	}
	return &trimmed, nil
}

func toResponse(report *Report) *ReportResponse {
	return &ReportResponse{
		ID:             report.ID,
		ReporterID:     report.ReporterID,
		ReportedUserID: report.ReportedUserID,
		BookingID:      report.BookingID,
		Reason:         report.Reason,
		Description:    report.Description,
		Status:         report.Status,
		CreatedAt:      report.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      report.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

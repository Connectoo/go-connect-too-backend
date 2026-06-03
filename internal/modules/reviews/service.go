package reviews

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/bookings"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

const maxCommentLength = 2000
const maxReplyLength = 2000

// CustomerProfileStore resolves customer profiles.
type CustomerProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*customers.Profile, error)
}

// EmployeeProfileStore resolves employee profiles.
type EmployeeProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*employees.Profile, error)
	GetApprovedByID(ctx context.Context, id uuid.UUID) (*employees.Profile, error)
}

// BookingReader loads bookings for review validation.
type BookingReader interface {
	GetByID(ctx context.Context, bookingID uuid.UUID) (*bookings.Booking, error)
}

// Store persists reviews and replies.
type Store interface {
	Create(ctx context.Context, review *Review) (*Review, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Review, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*Review, error)
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID, status string, offset, limit int) ([]Review, int, error)
	ListAdmin(ctx context.Context, status string, offset, limit int) ([]Review, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, at time.Time) (*Review, error)
	CreateReply(ctx context.Context, reply *Reply) (*Reply, error)
	GetReplyByReviewID(ctx context.Context, reviewID uuid.UUID) (*Reply, error)
	ListRepliesByReviewIDs(ctx context.Context, reviewIDs []uuid.UUID) (map[uuid.UUID]Reply, error)
}

// BadgeAwarder awards trust badges after verified reviews.
type BadgeAwarder interface {
	AwardVerifiedBookingReview(ctx context.Context, employeeID uuid.UUID) error
}

// RatingRefresher recalculates employee ratings after moderation.
type RatingRefresher interface {
	RefreshEmployeeRating(ctx context.Context, employeeID uuid.UUID) error
}

// NoopBadgeAwarder discards badge awards.
type NoopBadgeAwarder struct{}

// AwardVerifiedBookingReview implements BadgeAwarder.
func (NoopBadgeAwarder) AwardVerifiedBookingReview(context.Context, uuid.UUID) error { return nil }

// NoopRatingRefresher discards rating refresh calls.
type NoopRatingRefresher struct{}

// RefreshEmployeeRating implements RatingRefresher.
func (NoopRatingRefresher) RefreshEmployeeRating(context.Context, uuid.UUID) error { return nil }

// Service handles review business logic.
type Service struct {
	customers CustomerProfileStore
	employees EmployeeProfileStore
	bookings  BookingReader
	store     Store
	badges    BadgeAwarder
	ratings   RatingRefresher
	now       func() time.Time
}

// NewService creates a reviews service.
func NewService(
	customers CustomerProfileStore,
	employees EmployeeProfileStore,
	bookings BookingReader,
	store Store,
	badges BadgeAwarder,
	ratings RatingRefresher,
) *Service {
	if badges == nil {
		badges = NoopBadgeAwarder{}
	}
	if ratings == nil {
		ratings = NoopRatingRefresher{}
	}
	return &Service{
		customers: customers,
		employees: employees,
		bookings:  bookings,
		store:     store,
		badges:    badges,
		ratings:   ratings,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// CreateForBooking creates a review for a completed booking owned by the customer.
func (s *Service) CreateForBooking(ctx context.Context, userID, bookingID uuid.UUID, req CreateReviewRequest) (*ReviewResponse, error) {
	customer, err := s.customers.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return nil, ErrCustomerProfileNotFound
		}
		return nil, err
	}

	booking, err := s.bookings.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, bookings.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if booking.CustomerID != customer.ID {
		return nil, ErrForbidden
	}
	if booking.Status != bookings.StatusCompleted {
		return nil, ErrBookingNotCompleted
	}

	if _, err := s.store.GetByBookingID(ctx, bookingID); err == nil {
		return nil, ErrReviewAlreadyExists
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	comment, err := optionalText(req.Comment, maxCommentLength)
	if err != nil {
		return nil, err
	}
	if err := validateRating(req.Rating); err != nil {
		return nil, err
	}

	at := s.now()
	review := &Review{
		ID:         uuid.New(),
		BookingID:  bookingID,
		CustomerID: customer.ID,
		EmployeeID: booking.EmployeeID,
		Rating:     req.Rating,
		Comment:    comment,
		Status:     StatusPending,
		CreatedAt:  at,
		UpdatedAt:  at,
	}

	created, err := s.store.Create(ctx, review)
	if err != nil {
		return nil, err
	}

	if err := s.badges.AwardVerifiedBookingReview(ctx, booking.EmployeeID); err != nil {
		return nil, err
	}

	return toResponse(created, nil), nil
}

// ListForEmployeePublic returns approved reviews for a public employee profile.
func (s *Service) ListForEmployeePublic(ctx context.Context, employeeID uuid.UUID, page pagination.Params) (pagination.Result[ReviewResponse], error) {
	if _, err := s.employees.GetApprovedByID(ctx, employeeID); err != nil {
		return pagination.Result[ReviewResponse]{}, err
	}
	return s.listByEmployee(ctx, employeeID, StatusApproved, page)
}

// ListForEmployee returns reviews for the authenticated employee.
func (s *Service) ListForEmployee(ctx context.Context, userID uuid.UUID, page pagination.Params) (pagination.Result[ReviewResponse], error) {
	profile, err := s.employees.GetByUserID(ctx, userID)
	if err != nil {
		return pagination.Result[ReviewResponse]{}, err
	}
	return s.listByEmployee(ctx, profile.ID, "", page)
}

// Reply adds an employee reply to one of their reviews.
func (s *Service) Reply(ctx context.Context, userID, reviewID uuid.UUID, req ReplyRequest) (*ReviewResponse, error) {
	profile, err := s.employees.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	review, err := s.store.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if review.EmployeeID != profile.ID {
		return nil, ErrForbidden
	}

	replyText := strings.TrimSpace(req.Reply)
	if replyText == "" {
		return nil, fmt.Errorf("%w: reply is required", ErrValidation)
	}
	if utf8.RuneCountInString(replyText) > maxReplyLength {
		return nil, fmt.Errorf("%w: reply exceeds %d characters", ErrValidation, maxReplyLength)
	}

	at := s.now()
	reply, err := s.store.CreateReply(ctx, &Reply{
		ID:         uuid.New(),
		ReviewID:   reviewID,
		EmployeeID: profile.ID,
		Reply:      replyText,
		CreatedAt:  at,
		UpdatedAt:  at,
	})
	if err != nil {
		return nil, err
	}

	return toResponse(review, reply), nil
}

// ListForAdmin returns paginated reviews for moderation.
func (s *Service) ListForAdmin(ctx context.Context, status string, page pagination.Params) (pagination.Result[ReviewResponse], error) {
	items, total, err := s.store.ListAdmin(ctx, status, page.Offset(), page.Limit)
	if err != nil {
		return pagination.Result[ReviewResponse]{}, err
	}
	return pagination.NewResult(s.attachReplies(ctx, items), page, total), nil
}

// Approve marks a review as approved and refreshes employee rating.
func (s *Service) Approve(ctx context.Context, reviewID uuid.UUID) (*ReviewResponse, error) {
	review, err := s.store.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if review.Status == StatusApproved {
		return s.responseWithReply(ctx, review)
	}
	if review.Status == StatusHidden {
		return nil, ErrInvalidStatus
	}

	updated, err := s.store.UpdateStatus(ctx, reviewID, StatusApproved, s.now())
	if err != nil {
		return nil, err
	}
	if err := s.ratings.RefreshEmployeeRating(ctx, updated.EmployeeID); err != nil {
		return nil, err
	}
	return s.responseWithReply(ctx, updated)
}

// Hide marks a review as hidden and refreshes employee rating when needed.
func (s *Service) Hide(ctx context.Context, reviewID uuid.UUID) (*ReviewResponse, error) {
	review, err := s.store.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if review.Status == StatusHidden {
		return s.responseWithReply(ctx, review)
	}

	wasApproved := review.Status == StatusApproved
	updated, err := s.store.UpdateStatus(ctx, reviewID, StatusHidden, s.now())
	if err != nil {
		return nil, err
	}
	if wasApproved {
		if err := s.ratings.RefreshEmployeeRating(ctx, updated.EmployeeID); err != nil {
			return nil, err
		}
	}
	return s.responseWithReply(ctx, updated)
}

func (s *Service) listByEmployee(ctx context.Context, employeeID uuid.UUID, status string, page pagination.Params) (pagination.Result[ReviewResponse], error) {
	items, total, err := s.store.ListByEmployeeID(ctx, employeeID, status, page.Offset(), page.Limit)
	if err != nil {
		return pagination.Result[ReviewResponse]{}, err
	}
	return pagination.NewResult(s.attachReplies(ctx, items), page, total), nil
}

func (s *Service) attachReplies(ctx context.Context, items []Review) []ReviewResponse {
	if len(items) == 0 {
		return []ReviewResponse{}
	}

	ids := make([]uuid.UUID, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}

	replies, err := s.store.ListRepliesByReviewIDs(ctx, ids)
	if err != nil {
		replies = map[uuid.UUID]Reply{}
	}

	out := make([]ReviewResponse, 0, len(items))
	for i := range items {
		var reply *Reply
		if value, ok := replies[items[i].ID]; ok {
			copy := value
			reply = &copy
		}
		out = append(out, *toResponse(&items[i], reply))
	}
	return out
}

func (s *Service) responseWithReply(ctx context.Context, review *Review) (*ReviewResponse, error) {
	reply, err := s.store.GetReplyByReviewID(ctx, review.ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if errors.Is(err, ErrNotFound) {
		reply = nil
	}
	return toResponse(review, reply), nil
}

func validateRating(rating int) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("%w: rating must be between 1 and 5", ErrValidation)
	}
	return nil
}

func optionalText(value *string, maxLen int) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > maxLen {
		return nil, fmt.Errorf("%w: text exceeds %d characters", ErrValidation, maxLen)
	}
	return &trimmed, nil
}

func toResponse(review *Review, reply *Reply) *ReviewResponse {
	res := &ReviewResponse{
		ID:         review.ID,
		BookingID:  review.BookingID,
		CustomerID: review.CustomerID,
		EmployeeID: review.EmployeeID,
		Rating:     review.Rating,
		Comment:    review.Comment,
		Status:     review.Status,
		CreatedAt:  review.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  review.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if reply != nil {
		res.Reply = &ReplyResponse{
			ID:         reply.ID,
			ReviewID:   reply.ReviewID,
			EmployeeID: reply.EmployeeID,
			Reply:      reply.Reply,
			CreatedAt:  reply.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:  reply.UpdatedAt.UTC().Format(time.RFC3339),
		}
	}
	return res
}

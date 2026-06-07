package support

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

const maxSubjectLength = 255
const maxMessageLength = 2000

type CustomerProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*customers.Profile, error)
}

type Store interface {
	CreateTicket(ctx context.Context, ticket *Ticket, message *Message) (*Ticket, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Ticket, error)
	ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]Ticket, error)
	ListAdmin(ctx context.Context, status string, offset, limit int) ([]Ticket, int, error)
	UpdateTicket(ctx context.Context, id uuid.UUID, status, priority *string, at time.Time) (*Ticket, error)
	AddMessage(ctx context.Context, message *Message, at time.Time) (*Message, error)
	ListMessages(ctx context.Context, ticketID uuid.UUID) ([]Message, error)
}

type Service struct {
	customers CustomerProfileStore
	store     Store
	now       func() time.Time
}

func NewService(customers CustomerProfileStore, store Store) *Service {
	return &Service{
		customers: customers,
		store:     store,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateTicketRequest) (*TicketResponse, error) {
	customer, err := s.customers.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	subject, err := validateSubject(req.Subject)
	if err != nil {
		return nil, err
	}
	messageText, err := validateMessage(req.Message)
	if err != nil {
		return nil, err
	}

	at := s.now()
	ticketID := uuid.New()
	created, err := s.store.CreateTicket(ctx, &Ticket{
		ID:         ticketID,
		CustomerID: customer.ID,
		Subject:    subject,
		Status:     StatusOpen,
		Priority:   PriorityNormal,
		CreatedAt:  at,
		UpdatedAt:  at,
	}, &Message{
		ID:        uuid.New(),
		TicketID:  ticketID,
		SenderID:  userID,
		Message:   messageText,
		IsStaff:   false,
		CreatedAt: at,
	})
	if err != nil {
		return nil, err
	}
	return toTicketResponse(created), nil
}

func (s *Service) ListForCustomer(ctx context.Context, userID uuid.UUID) ([]TicketResponse, error) {
	customer, err := s.customers.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	items, err := s.store.ListByCustomerID(ctx, customer.ID)
	if err != nil {
		return nil, err
	}
	return toTicketResponses(items), nil
}

func (s *Service) ListForAdmin(ctx context.Context, status string, page pagination.Params) (pagination.Result[TicketResponse], error) {
	items, total, err := s.store.ListAdmin(ctx, status, page.Offset(), page.Limit)
	if err != nil {
		return pagination.Result[TicketResponse]{}, err
	}
	return pagination.NewResult(toTicketResponses(items), page, total), nil
}

func (s *Service) GetForAdmin(ctx context.Context, ticketID uuid.UUID) (*TicketDetailResponse, error) {
	ticket, err := s.store.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	messages, err := s.store.ListMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	out := make([]MessageResponse, 0, len(messages))
	for i := range messages {
		out = append(out, toMessageResponse(&messages[i]))
	}
	return &TicketDetailResponse{
		TicketResponse: *toTicketResponse(ticket),
		Messages:       out,
	}, nil
}

func (s *Service) UpdateForAdmin(ctx context.Context, ticketID uuid.UUID, req UpdateTicketRequest) (*TicketResponse, error) {
	status, err := optionalStatus(req.Status)
	if err != nil {
		return nil, err
	}
	priority, err := optionalPriority(req.Priority)
	if err != nil {
		return nil, err
	}
	updated, err := s.store.UpdateTicket(ctx, ticketID, status, priority, s.now())
	if err != nil {
		return nil, err
	}
	return toTicketResponse(updated), nil
}

func (s *Service) AddMessageForAdmin(ctx context.Context, adminUserID, ticketID uuid.UUID, req AddMessageRequest) (*MessageResponse, error) {
	messageText, err := validateMessage(req.Message)
	if err != nil {
		return nil, err
	}
	if _, err := s.store.GetByID(ctx, ticketID); err != nil {
		return nil, err
	}
	at := s.now()
	created, err := s.store.AddMessage(ctx, &Message{
		ID:        uuid.New(),
		TicketID:  ticketID,
		SenderID:  adminUserID,
		Message:   messageText,
		IsStaff:   true,
		CreatedAt: at,
	}, at)
	if err != nil {
		return nil, err
	}
	res := toMessageResponse(created)
	return &res, nil
}

func validateSubject(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%w: subject is required", ErrValidation)
	}
	if utf8.RuneCountInString(trimmed) > maxSubjectLength {
		return "", fmt.Errorf("%w: subject is too long", ErrValidation)
	}
	return trimmed, nil
}

func validateMessage(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%w: message is required", ErrValidation)
	}
	if utf8.RuneCountInString(trimmed) > maxMessageLength {
		return "", fmt.Errorf("%w: message is too long", ErrValidation)
	}
	return trimmed, nil
}

func optionalStatus(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	switch trimmed {
	case StatusOpen, StatusInProgress, StatusResolved, StatusClosed:
		return &trimmed, nil
	default:
		return nil, fmt.Errorf("%w: invalid status", ErrValidation)
	}
}

func optionalPriority(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	switch trimmed {
	case PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent:
		return &trimmed, nil
	default:
		return nil, fmt.Errorf("%w: invalid priority", ErrValidation)
	}
}

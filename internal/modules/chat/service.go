package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/events"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

const maxMessageLength = 2000

// CustomerProfileStore resolves customer profiles.
type CustomerProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*customers.Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*customers.Profile, error)
}

// EmployeeProfileStore resolves employee profiles.
type EmployeeProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*employees.Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*employees.Profile, error)
}

// Store persists chat data.
type Store interface {
	EnsureForBooking(ctx context.Context, bookingID, customerID, employeeID uuid.UUID, at time.Time) (*Conversation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	ListByParticipant(ctx context.Context, customerID, employeeID *uuid.UUID) ([]Conversation, error)
	HasBookingBetween(ctx context.Context, customerID, employeeID uuid.UUID) (bool, error)
	CreateMessage(ctx context.Context, message *Message, at time.Time) (*Message, error)
	ListMessages(ctx context.Context, conversationID uuid.UUID, offset, limit int) ([]Message, int, error)
}

// EventPublisher emits chat-related platform events.
type EventPublisher interface {
	Publish(ctx context.Context, event events.Event)
}

// NoopEventPublisher discards events.
type NoopEventPublisher struct{}

// Publish implements EventPublisher.
func (NoopEventPublisher) Publish(context.Context, events.Event) {}

// Service handles chat business logic.
type Service struct {
	customers CustomerProfileStore
	employees EmployeeProfileStore
	store     Store
	events    EventPublisher
	now       func() time.Time
}

// NewService creates a chat service.
func NewService(
	customers CustomerProfileStore,
	employees EmployeeProfileStore,
	store Store,
	publisher EventPublisher,
) *Service {
	if publisher == nil {
		publisher = NoopEventPublisher{}
	}
	return &Service{
		customers: customers,
		employees: employees,
		store:     store,
		events:    publisher,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// EnsureConversationForBooking creates a conversation when a booking exists.
func (s *Service) EnsureConversationForBooking(ctx context.Context, bookingID, customerID, employeeID uuid.UUID) (*Conversation, error) {
	return s.store.EnsureForBooking(ctx, bookingID, customerID, employeeID, s.now())
}

// ListConversations returns conversations for the authenticated user.
func (s *Service) ListConversations(ctx context.Context, userID uuid.UUID) ([]ConversationResponse, error) {
	customerID, employeeID, err := s.participantProfileIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, err := s.store.ListByParticipant(ctx, customerID, employeeID)
	if err != nil {
		return nil, err
	}
	return toConversationResponses(items), nil
}

// ListMessages returns paginated messages when the user belongs to the conversation.
func (s *Service) ListMessages(ctx context.Context, userID, conversationID uuid.UUID, page pagination.Params) (pagination.Result[MessageResponse], error) {
	conversation, err := s.authorizeConversation(ctx, userID, conversationID)
	if err != nil {
		return pagination.Result[MessageResponse]{}, err
	}
	_ = conversation

	items, total, err := s.store.ListMessages(ctx, conversationID, page.Offset(), page.Limit)
	if err != nil {
		return pagination.Result[MessageResponse]{}, err
	}
	return pagination.NewResult(toMessageResponses(items), page, total), nil
}

// SendMessage stores a message and emits message.sent.
func (s *Service) SendMessage(ctx context.Context, userID, conversationID uuid.UUID, req SendMessageRequest) (*MessageResponse, error) {
	conversation, err := s.authorizeConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	messageText, err := validateMessage(req.Message)
	if err != nil {
		return nil, err
	}

	at := s.now()
	created, err := s.store.CreateMessage(ctx, &Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       userID,
		Message:        messageText,
		CreatedAt:      at,
	}, at)
	if err != nil {
		return nil, err
	}

	recipientUserID, err := s.recipientUserID(ctx, conversation, userID)
	if err != nil {
		return nil, err
	}

	s.events.Publish(ctx, events.Event{
		Type: events.TypeMessageSent,
		Payload: map[string]any{
			"conversation_id":   conversation.ID.String(),
			"message_id":        created.ID.String(),
			"sender_id":         userID.String(),
			"recipient_user_id": recipientUserID.String(),
			"message":           created.Message,
			"customer_id":       conversation.CustomerID.String(),
			"employee_id":       conversation.EmployeeID.String(),
		},
	})

	res := toMessageResponse(created)
	return &res, nil
}

func (s *Service) authorizeConversation(ctx context.Context, userID, conversationID uuid.UUID) (*Conversation, error) {
	conversation, err := s.store.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	customer, err := s.customers.GetByID(ctx, conversation.CustomerID)
	if err != nil {
		if errors.Is(err, customers.ErrNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	employee, err := s.employees.GetByID(ctx, conversation.EmployeeID)
	if err != nil {
		if errors.Is(err, employees.ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	if userID != customer.UserID && userID != employee.UserID {
		return nil, ErrForbidden
	}

	allowed, err := s.store.HasBookingBetween(ctx, conversation.CustomerID, conversation.EmployeeID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrChatNotAllowed
	}

	return conversation, nil
}

func (s *Service) participantProfileIDs(ctx context.Context, userID uuid.UUID) (*uuid.UUID, *uuid.UUID, error) {
	if customer, err := s.customers.GetByUserID(ctx, userID); err == nil {
		id := customer.ID
		return &id, nil, nil
	} else if !errors.Is(err, customers.ErrNotFound) {
		return nil, nil, err
	}

	if employee, err := s.employees.GetByUserID(ctx, userID); err == nil {
		id := employee.ID
		return nil, &id, nil
	} else if !errors.Is(err, employees.ErrNotFound) {
		return nil, nil, err
	}

	return nil, nil, ErrForbidden
}

func (s *Service) recipientUserID(ctx context.Context, conversation *Conversation, senderID uuid.UUID) (uuid.UUID, error) {
	customer, err := s.customers.GetByID(ctx, conversation.CustomerID)
	if err != nil {
		return uuid.Nil, err
	}
	if senderID == customer.UserID {
		employee, err := s.employees.GetByID(ctx, conversation.EmployeeID)
		if err != nil {
			return uuid.Nil, err
		}
		return employee.UserID, nil
	}

	return customer.UserID, nil
}

func validateMessage(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("%w: message is required", ErrValidation)
	}
	if utf8.RuneCountInString(trimmed) > maxMessageLength {
		return "", fmt.Errorf("%w: message exceeds maximum length", ErrValidation)
	}
	return trimmed, nil
}

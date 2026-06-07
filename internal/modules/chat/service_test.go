package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/events"
)

type mockCustomerStore struct {
	profiles map[uuid.UUID]*customers.Profile
}

func (m *mockCustomerStore) GetByUserID(_ context.Context, userID uuid.UUID) (*customers.Profile, error) {
	for _, profile := range m.profiles {
		if profile.UserID == userID {
			copy := *profile
			return &copy, nil
		}
	}
	return nil, customers.ErrNotFound
}

func (m *mockCustomerStore) GetByID(_ context.Context, id uuid.UUID) (*customers.Profile, error) {
	profile, ok := m.profiles[id]
	if !ok {
		return nil, customers.ErrNotFound
	}
	copy := *profile
	return &copy, nil
}

type mockEmployeeStore struct {
	profiles map[uuid.UUID]*employees.Profile
}

func (m *mockEmployeeStore) GetByUserID(_ context.Context, userID uuid.UUID) (*employees.Profile, error) {
	for _, profile := range m.profiles {
		if profile.UserID == userID {
			copy := *profile
			return &copy, nil
		}
	}
	return nil, employees.ErrNotFound
}

func (m *mockEmployeeStore) GetByID(_ context.Context, id uuid.UUID) (*employees.Profile, error) {
	profile, ok := m.profiles[id]
	if !ok {
		return nil, employees.ErrNotFound
	}
	copy := *profile
	return &copy, nil
}

type mockChatStore struct {
	conversations map[uuid.UUID]*Conversation
	messages      []Message
	hasBooking    bool
}

func (m *mockChatStore) EnsureForBooking(_ context.Context, bookingID, customerID, employeeID uuid.UUID, at time.Time) (*Conversation, error) {
	item := &Conversation{
		ID:         uuid.New(),
		CustomerID: customerID,
		EmployeeID: employeeID,
		BookingID:  &bookingID,
		CreatedAt:  at,
		UpdatedAt:  at,
	}
	m.conversations[item.ID] = item
	return item, nil
}

func (m *mockChatStore) GetByID(_ context.Context, id uuid.UUID) (*Conversation, error) {
	item, ok := m.conversations[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *item
	return &copy, nil
}

func (m *mockChatStore) ListByParticipant(_ context.Context, customerID, employeeID *uuid.UUID) ([]Conversation, error) {
	items := make([]Conversation, 0)
	for _, conversation := range m.conversations {
		if customerID != nil && conversation.CustomerID == *customerID {
			items = append(items, *conversation)
		}
		if employeeID != nil && conversation.EmployeeID == *employeeID {
			items = append(items, *conversation)
		}
	}
	return items, nil
}

func (m *mockChatStore) HasBookingBetween(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return m.hasBooking, nil
}

func (m *mockChatStore) CreateMessage(_ context.Context, message *Message, _ time.Time) (*Message, error) {
	copy := *message
	m.messages = append(m.messages, copy)
	return &copy, nil
}

func (m *mockChatStore) MarkMessageRead(_ context.Context, _, messageID, _ uuid.UUID, at time.Time) (*Message, error) {
	for i, message := range m.messages {
		if message.ID == messageID {
			readAt := at
			updated := message
			updated.ReadAt = &readAt
			m.messages[i] = updated
			return &updated, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockChatStore) ListMessages(_ context.Context, conversationID uuid.UUID, offset, limit int) ([]Message, int, error) {
	filtered := make([]Message, 0)
	for _, message := range m.messages {
		if message.ConversationID == conversationID {
			filtered = append(filtered, message)
		}
	}
	total := len(filtered)
	if offset >= total {
		return []Message{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

type recordingPublisher struct {
	events []events.Event
}

func (r *recordingPublisher) Publish(_ context.Context, event events.Event) {
	r.events = append(r.events, event)
}

func TestSendMessagePublishesEvent(t *testing.T) {
	customerUserID := uuid.New()
	employeeUserID := uuid.New()
	customerID := uuid.New()
	employeeID := uuid.New()
	conversationID := uuid.New()

	customersStore := &mockCustomerStore{profiles: map[uuid.UUID]*customers.Profile{
		customerID: {ID: customerID, UserID: customerUserID},
	}}
	employeesStore := &mockEmployeeStore{profiles: map[uuid.UUID]*employees.Profile{
		employeeID: {ID: employeeID, UserID: employeeUserID},
	}}
	chatStore := &mockChatStore{
		conversations: map[uuid.UUID]*Conversation{
			conversationID: {
				ID:         conversationID,
				CustomerID: customerID,
				EmployeeID: employeeID,
			},
		},
		hasBooking: true,
	}
	publisher := &recordingPublisher{}
	svc := NewService(customersStore, employeesStore, chatStore, publisher)

	_, err := svc.SendMessage(context.Background(), customerUserID, conversationID, SendMessageRequest{
		Message: "Hello",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("events = %d, want 1", len(publisher.events))
	}
	if publisher.events[0].Type != events.TypeMessageSent {
		t.Fatalf("event type = %q, want %q", publisher.events[0].Type, events.TypeMessageSent)
	}
}

func TestSendMessageForbiddenForOutsider(t *testing.T) {
	customerUserID := uuid.New()
	conversationID := uuid.New()
	customerID := uuid.New()
	employeeID := uuid.New()

	svc := NewService(
		&mockCustomerStore{profiles: map[uuid.UUID]*customers.Profile{
			customerID: {ID: customerID, UserID: customerUserID},
		}},
		&mockEmployeeStore{profiles: map[uuid.UUID]*employees.Profile{
			employeeID: {ID: employeeID, UserID: uuid.New()},
		}},
		&mockChatStore{
			conversations: map[uuid.UUID]*Conversation{
				conversationID: {ID: conversationID, CustomerID: customerID, EmployeeID: employeeID},
			},
			hasBooking: true,
		},
		NoopEventPublisher{},
	)

	_, err := svc.SendMessage(context.Background(), uuid.New(), conversationID, SendMessageRequest{Message: "Hello"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

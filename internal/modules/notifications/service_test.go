package notifications

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

type mockStore struct {
	notifications []Notification
}

func (m *mockStore) Create(_ context.Context, notification *Notification) (*Notification, error) {
	copy := *notification
	m.notifications = append(m.notifications, copy)
	return &copy, nil
}

func (m *mockStore) ListByUserID(_ context.Context, userID uuid.UUID, offset, limit int) ([]Notification, int, error) {
	filtered := make([]Notification, 0)
	for _, item := range m.notifications {
		if item.UserID == userID {
			filtered = append(filtered, item)
		}
	}
	total := len(filtered)
	if offset >= total {
		return []Notification{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

func (m *mockStore) GetByID(_ context.Context, id uuid.UUID) (*Notification, error) {
	for i := range m.notifications {
		if m.notifications[i].ID == id {
			copy := m.notifications[i]
			return &copy, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockStore) MarkRead(_ context.Context, id, userID uuid.UUID, _ time.Time) (*Notification, error) {
	for i := range m.notifications {
		if m.notifications[i].ID == id && m.notifications[i].UserID == userID {
			copy := m.notifications[i]
			return &copy, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockStore) MarkAllRead(_ context.Context, _ uuid.UUID, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockStore) UpsertDeviceToken(_ context.Context, token *DeviceToken) (*DeviceToken, error) {
	copy := *token
	return &copy, nil
}

func TestRegisterDeviceTokenValidation(t *testing.T) {
	svc := NewService(&mockStore{})
	userID := uuid.New()

	_, err := svc.RegisterDeviceToken(context.Background(), userID, RegisterDeviceTokenRequest{
		Platform: "desktop",
		Token:    "abc",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateNotificationSuccess(t *testing.T) {
	svc := NewService(&mockStore{})
	userID := uuid.New()

	res, err := svc.Create(context.Background(), CreateInput{
		UserID: userID,
		Type:   "booking.created",
		Title:  "Booking created",
		Body:   "Your booking was created",
		Data:   map[string]any{"booking_id": uuid.New().String()},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if res.Title != "Booking created" {
		t.Fatalf("title = %q, want Booking created", res.Title)
	}
}

func TestListNotificationsForUser(t *testing.T) {
	store := &mockStore{}
	svc := NewService(store)
	userID := uuid.New()
	otherUserID := uuid.New()

	_, err := svc.Create(context.Background(), CreateInput{
		UserID: userID,
		Type:   "booking.created",
		Title:  "Mine",
		Body:   "Body",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, err = svc.Create(context.Background(), CreateInput{
		UserID: otherUserID,
		Type:   "booking.created",
		Title:  "Other",
		Body:   "Body",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	result, err := svc.List(context.Background(), userID, pagination.Params{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(result.Items))
	}
	if result.Items[0].Title != "Mine" {
		t.Fatalf("title = %q, want Mine", result.Items[0].Title)
	}
}

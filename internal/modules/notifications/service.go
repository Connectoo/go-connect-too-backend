package notifications

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

// Store persists notifications and device tokens.
type Store interface {
	Create(ctx context.Context, notification *Notification) (*Notification, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]Notification, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID, at time.Time) (*Notification, error)
	MarkAllRead(ctx context.Context, userID uuid.UUID, at time.Time) (int64, error)
	UpsertDeviceToken(ctx context.Context, token *DeviceToken) (*DeviceToken, error)
}

// CreateInput carries data for a new notification.
type CreateInput struct {
	UserID uuid.UUID
	Type   string
	Title  string
	Body   string
	Data   map[string]any
}

// Service handles notification business logic.
type Service struct {
	store Store
	now   func() time.Time
}

// NewService creates a notification service.
func NewService(store Store) *Service {
	return &Service{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

// Create stores a notification for a user.
func (s *Service) Create(ctx context.Context, input CreateInput) (*NotificationResponse, error) {
	title := strings.TrimSpace(input.Title)
	body := strings.TrimSpace(input.Body)
	eventType := strings.TrimSpace(input.Type)
	if input.UserID == uuid.Nil || title == "" || body == "" || eventType == "" {
		return nil, fmt.Errorf("%w: user_id, type, title, and body are required", ErrValidation)
	}

	at := s.now()
	created, err := s.store.Create(ctx, &Notification{
		ID:        uuid.New(),
		UserID:    input.UserID,
		Type:      eventType,
		Title:     title,
		Body:      body,
		Data:      input.Data,
		CreatedAt: at,
	})
	if err != nil {
		return nil, err
	}

	res := toNotificationResponse(created)
	return &res, nil
}

// List returns paginated notifications for the authenticated user.
func (s *Service) List(ctx context.Context, userID uuid.UUID, page pagination.Params) (pagination.Result[NotificationResponse], error) {
	items, total, err := s.store.ListByUserID(ctx, userID, page.Offset(), page.Limit)
	if err != nil {
		return pagination.Result[NotificationResponse]{}, err
	}
	return pagination.NewResult(toNotificationResponses(items), page, total), nil
}

// MarkRead marks one notification as read for the authenticated user.
func (s *Service) MarkRead(ctx context.Context, userID, notificationID uuid.UUID) (*NotificationResponse, error) {
	item, err := s.store.MarkRead(ctx, notificationID, userID, s.now())
	if err != nil {
		return nil, err
	}
	res := toNotificationResponse(item)
	return &res, nil
}

// MarkAllRead marks all notifications as read for the authenticated user.
func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.store.MarkAllRead(ctx, userID, s.now())
}

// RegisterDeviceToken registers or refreshes a push device token.
func (s *Service) RegisterDeviceToken(ctx context.Context, userID uuid.UUID, req RegisterDeviceTokenRequest) (*DeviceTokenResponse, error) {
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	token := strings.TrimSpace(req.Token)
	if platform == "" || token == "" {
		return nil, fmt.Errorf("%w: platform and token are required", ErrValidation)
	}
	if platform != PlatformIOS && platform != PlatformAndroid && platform != PlatformWeb {
		return nil, fmt.Errorf("%w: platform must be ios, android, or web", ErrValidation)
	}

	at := s.now()
	created, err := s.store.UpsertDeviceToken(ctx, &DeviceToken{
		ID:        uuid.New(),
		UserID:    userID,
		Platform:  platform,
		Token:     token,
		IsActive:  true,
		CreatedAt: at,
		UpdatedAt: at,
	})
	if err != nil {
		return nil, err
	}

	res := toDeviceTokenResponse(created)
	return &res, nil
}

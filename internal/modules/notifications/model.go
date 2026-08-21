package notifications

import (
	"time"

	"github.com/google/uuid"
)

// Notification is a stored in-app notification for a user.
type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Title     string
	Body      string
	Data      map[string]any
	ReadAt    *time.Time
	CreatedAt time.Time
}

// DeviceToken stores a push notification device token.
type DeviceToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Platform  string
	Token     string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformWeb     = "web"
)

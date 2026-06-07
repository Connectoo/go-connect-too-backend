package notifications

import (
	"time"

	"github.com/google/uuid"
)

const (
	DeliveryChannelPush   = "push"
	DeliveryStatusSent    = "sent"
	DeliveryStatusFailed  = "failed"
	DeliveryStatusSkipped = "skipped"
)

// Delivery tracks outbound notification delivery attempts.
type Delivery struct {
	ID             uuid.UUID
	NotificationID *uuid.UUID
	UserID         uuid.UUID
	Channel        string
	Provider       string
	Status         string
	ErrorMessage   *string
	CreatedAt      time.Time
}

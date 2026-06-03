package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Type identifies a platform event.
type Type string

const (
	TypeBookingCreated       Type = "booking.created"
	TypeBookingAccepted      Type = "booking.accepted"
	TypeBookingRejected      Type = "booking.rejected"
	TypeBookingCancelled     Type = "booking.cancelled"
	TypeBookingCompleted     Type = "booking.completed"
	TypeMessageSent          Type = "message.sent"
	TypePaymentSuccess       Type = "payment.success"
	TypeSubscriptionExpiring Type = "subscription.expiring"
	TypeKYCApproved          Type = "kyc.approved"
	TypeKYCRejected          Type = "kyc.rejected"
)

// Event is an in-app domain event published through the dispatcher.
type Event struct {
	Type      Type
	Payload   map[string]any
	CreatedAt time.Time
}

// StringPayload returns a string field from the payload when present.
func (e Event) StringPayload(key string) string {
	value, ok := e.Payload[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

// UUIDPayload parses a UUID field from the payload.
func (e Event) UUIDPayload(key string) (uuid.UUID, bool) {
	raw := e.StringPayload(key)
	if raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// MarshalPayload serializes payload data for storage.
func MarshalPayload(payload map[string]any) ([]byte, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	return json.Marshal(payload)
}

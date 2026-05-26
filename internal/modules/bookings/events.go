package bookings

import (
	"context"

	"github.com/google/uuid"
)

// EventType identifies a booking notification event (delivery wired in a later phase).
type EventType string

const (
	EventBookingCreated   EventType = "booking.created"
	EventBookingAccepted  EventType = "booking.accepted"
	EventBookingRejected  EventType = "booking.rejected"
	EventBookingStarted   EventType = "booking.started"
	EventBookingCompleted EventType = "booking.completed"
	EventBookingCancelled EventType = "booking.cancelled"
)

// BookingEvent is a placeholder notification payload for future push/WebSocket delivery.
type BookingEvent struct {
	Type       EventType
	BookingID  uuid.UUID
	CustomerID uuid.UUID
	EmployeeID uuid.UUID
	Status     string
}

// EventPublisher emits booking notification events without blocking business logic.
type EventPublisher interface {
	Publish(ctx context.Context, event BookingEvent)
}

// NoopEventPublisher discards events until the notifications module is implemented.
type NoopEventPublisher struct{}

// Publish implements EventPublisher.
func (NoopEventPublisher) Publish(context.Context, BookingEvent) {}

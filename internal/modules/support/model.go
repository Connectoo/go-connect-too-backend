package support

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"

	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// Ticket is a customer support request.
type Ticket struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Subject    string
	Status     string
	Priority   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Message is a message on a support ticket.
type Message struct {
	ID        uuid.UUID
	TicketID  uuid.UUID
	SenderID  uuid.UUID
	Message   string
	IsStaff   bool
	CreatedAt time.Time
}

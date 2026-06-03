package chat

import (
	"time"

	"github.com/google/uuid"
)

// Conversation links a customer and employee, optionally to a booking.
type Conversation struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	EmployeeID uuid.UUID
	BookingID  *uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Message is a chat message within a conversation.
type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Message        string
	ReadAt         *time.Time
	CreatedAt      time.Time
}

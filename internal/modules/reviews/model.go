package reviews

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusHidden   = "hidden"
)

// Review is a customer rating for a completed booking.
type Review struct {
	ID           uuid.UUID
	BookingID    uuid.UUID
	CustomerID   uuid.UUID
	EmployeeID   uuid.UUID
	Rating       int
	Comment      *string
	ReviewImages []string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Reply is an employee response to a review.
type Reply struct {
	ID         uuid.UUID
	ReviewID   uuid.UUID
	EmployeeID uuid.UUID
	Reply      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

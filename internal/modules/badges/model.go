package badges

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeVerifiedBookingReview = "verified_booking_review"
)

// Badge is a trust indicator awarded to an employee.
type Badge struct {
	ID         uuid.UUID
	EmployeeID uuid.UUID
	BadgeType  string
	CreatedAt  time.Time
}

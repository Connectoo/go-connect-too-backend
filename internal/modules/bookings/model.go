package bookings

import (
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/availability"
)

// Booking is a service appointment between a customer and an employee.
type Booking struct {
	ID              uuid.UUID
	CustomerID      uuid.UUID
	EmployeeID      uuid.UUID
	ServiceID       uuid.UUID
	BookingDate     time.Time
	StartTime       availability.TimeOfDay
	EndTime         availability.TimeOfDay
	Status          string
	CustomerNotes   *string
	EmployeeNotes   *string
	TotalAmount     float64
	SourceBookingID *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// StatusHistory records a booking status change.
type StatusHistory struct {
	ID              uuid.UUID
	BookingID       uuid.UUID
	OldStatus       *string
	NewStatus       string
	ChangedByUserID uuid.UUID
	Reason          *string
	CreatedAt       time.Time
}

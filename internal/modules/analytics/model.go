package analytics

import (
	"time"

	"github.com/google/uuid"
)

// BookingDayCount is a daily booking aggregate.
type BookingDayCount struct {
	Date  time.Time
	Count int
}

// StatusCount groups bookings by status.
type StatusCount struct {
	Status string
	Count  int
}

// RatingPeriod is average rating for a time bucket.
type RatingPeriod struct {
	Period        time.Time
	AverageRating float64
	ReviewCount   int
}

// CategoryBookingCount ranks categories by bookings in a period.
type CategoryBookingCount struct {
	CategoryID   uuid.UUID
	CategoryName string
	BookingCount int
}

// RevenueDay is daily revenue from successful subscription payments.
type RevenueDay struct {
	Date   time.Time
	Amount int64
}

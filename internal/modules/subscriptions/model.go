package subscriptions

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending   = "pending"
	StatusActive    = "active"
	StatusExpired   = "expired"
	StatusCancelled = "cancelled"
)

type Plan struct {
	ID                uuid.UUID
	Name              string
	Price             int64
	Currency          string
	DurationDays      int
	ServiceLimit      int
	IsFeaturedAllowed bool
	IsPriorityAllowed bool
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type EmployeeSubscription struct {
	ID         uuid.UUID
	EmployeeID uuid.UUID
	PlanID     uuid.UUID
	PlanName   string
	Status     string
	StartsAt   *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Plan       *Plan
}

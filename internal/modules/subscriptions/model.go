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
	ID                 uuid.UUID
	EmployeeID         uuid.UUID
	PlanID             uuid.UUID
	PlanName           string
	Status             string
	StartsAt           *time.Time
	ExpiresAt          *time.Time
	AutoRenew          bool
	CancelledAt        *time.Time
	CancellationReason *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Plan               *Plan
}

// SubscriptionChange records subscription lifecycle audit events.
type SubscriptionChange struct {
	ID             uuid.UUID
	SubscriptionID uuid.UUID
	EmployeeID     uuid.UUID
	ChangeType     string
	OldPlanID      *uuid.UUID
	NewPlanID      *uuid.UUID
	Reason         *string
	CreatedAt      time.Time
}

const (
	ChangeTypeCancel     = "cancel"
	ChangeTypeChangePlan = "change_plan"
	ChangeTypeAutoRenew  = "auto_renew"
)

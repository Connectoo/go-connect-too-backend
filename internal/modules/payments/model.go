package payments

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

type Payment struct {
	ID                uuid.UUID
	EmployeeID        uuid.UUID
	SubscriptionID    uuid.UUID
	Provider          string
	ProviderOrderID   string
	ProviderPaymentID *string
	Amount            int64
	Currency          string
	Status            string
	RawResponse       []byte
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Refund struct {
	ID               uuid.UUID
	PaymentID        uuid.UUID
	Amount           int64
	Reason           *string
	Status           string
	ProviderRefundID *string
	CreatedBy        uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

const (
	RefundStatusPending   = "pending"
	RefundStatusCompleted = "completed"
	RefundStatusFailed    = "failed"
)

type WebhookEvent struct {
	ID          uuid.UUID
	Provider    string
	EventID     string
	EventType   string
	Payload     []byte
	ProcessedAt *time.Time
	CreatedAt   time.Time
}

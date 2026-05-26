package payments

import "github.com/google/uuid"

type PlanSnapshot struct {
	ID           uuid.UUID
	Name         string
	Price        int64
	Currency     string
	DurationDays int
}

type CreateSubscriptionOrderResponse struct {
	PaymentID       uuid.UUID `json:"payment_id"`
	SubscriptionID  uuid.UUID `json:"subscription_id"`
	Provider        string    `json:"provider"`
	ProviderOrderID string    `json:"provider_order_id"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	RazorpayKeyID   string    `json:"razorpay_key_id,omitempty"`
}

type PaymentResponse struct {
	ID                uuid.UUID `json:"id"`
	EmployeeID        uuid.UUID `json:"employee_id"`
	SubscriptionID    uuid.UUID `json:"subscription_id"`
	Provider          string    `json:"provider"`
	ProviderOrderID   string    `json:"provider_order_id"`
	ProviderPaymentID *string   `json:"provider_payment_id,omitempty"`
	Amount            int64     `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	CreatedAt         string    `json:"created_at"`
	UpdatedAt         string    `json:"updated_at"`
}

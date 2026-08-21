package subscriptions

import "github.com/google/uuid"

type CreateOrderRequest struct {
	PlanID uuid.UUID `json:"plan_id"`
}

type CreatePlanRequest struct {
	Name              string `json:"name"`
	Price             int64  `json:"price"`
	Currency          string `json:"currency"`
	DurationDays      int    `json:"duration_days"`
	ServiceLimit      int    `json:"service_limit"`
	IsFeaturedAllowed bool   `json:"is_featured_allowed"`
	IsPriorityAllowed bool   `json:"is_priority_allowed"`
	IsActive          *bool  `json:"is_active,omitempty"`
}

type UpdatePlanRequest = CreatePlanRequest

type PlanResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Price             int64     `json:"price"`
	Currency          string    `json:"currency"`
	DurationDays      int       `json:"duration_days"`
	ServiceLimit      int       `json:"service_limit"`
	IsFeaturedAllowed bool      `json:"is_featured_allowed"`
	IsPriorityAllowed bool      `json:"is_priority_allowed"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         string    `json:"created_at"`
	UpdatedAt         string    `json:"updated_at"`
}

type SubscriptionResponse struct {
	ID                 uuid.UUID     `json:"id"`
	EmployeeID         uuid.UUID     `json:"employee_id"`
	PlanID             uuid.UUID     `json:"plan_id"`
	PlanName           string        `json:"plan_name"`
	Status             string        `json:"status"`
	StartsAt           *string       `json:"starts_at,omitempty"`
	ExpiresAt          *string       `json:"expires_at,omitempty"`
	AutoRenew          bool          `json:"auto_renew"`
	CancelledAt        *string       `json:"cancelled_at,omitempty"`
	CancellationReason *string       `json:"cancellation_reason,omitempty"`
	Plan               *PlanResponse `json:"plan,omitempty"`
	CreatedAt          string        `json:"created_at"`
	UpdatedAt          string        `json:"updated_at"`
}

type CancelSubscriptionRequest struct {
	Reason *string `json:"reason,omitempty"`
}

type ChangePlanRequest struct {
	PlanID uuid.UUID `json:"plan_id"`
}

type AutoRenewRequest struct {
	AutoRenew bool `json:"auto_renew"`
}

type VerifyPaymentRequest struct {
	PaymentID         uuid.UUID `json:"payment_id"`
	ProviderOrderID   string    `json:"provider_order_id"`
	ProviderPaymentID string    `json:"provider_payment_id"`
	Signature         string    `json:"signature"`
}

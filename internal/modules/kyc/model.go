package kyc

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// Record is an employee KYC submission.
type Record struct {
	ID              uuid.UUID
	EmployeeID      uuid.UUID
	IDProofURL      string
	AddressProofURL string
	Status          string
	RejectionReason *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

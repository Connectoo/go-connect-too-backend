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
	ReviewedBy      *uuid.UUID
	ReviewedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AdminListItem is a KYC record with linked employee and user metadata.
type AdminListItem struct {
	Record
	EmployeeDisplayName *string
	UserName            string
	UserEmail           string
}

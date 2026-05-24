package kyc

import "github.com/google/uuid"

// SubmitRequest creates or resubmits employee KYC documents.
type SubmitRequest struct {
	IDProofURL      string `json:"id_proof_url"`
	AddressProofURL string `json:"address_proof_url"`
}

// Response is the public KYC payload.
type Response struct {
	ID              uuid.UUID `json:"id"`
	EmployeeID      uuid.UUID `json:"employee_id"`
	IDProofURL      string    `json:"id_proof_url"`
	AddressProofURL string    `json:"address_proof_url"`
	Status          string    `json:"status"`
	RejectionReason *string   `json:"rejection_reason,omitempty"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

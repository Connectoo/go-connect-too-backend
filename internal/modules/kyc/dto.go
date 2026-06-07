package kyc

import "github.com/google/uuid"

// SubmitRequest creates or resubmits employee KYC documents.
type SubmitRequest struct {
	IDProofURL         string     `json:"id_proof_url"`
	AddressProofURL    string     `json:"address_proof_url"`
	IDProofFileID      *uuid.UUID `json:"id_proof_file_id,omitempty"`
	AddressProofFileID *uuid.UUID `json:"address_proof_file_id,omitempty"`
}

// RejectRequest carries admin rejection reason.
type RejectRequest struct {
	Reason string `json:"reason"`
}

// Response is the public KYC payload.
type Response struct {
	ID              uuid.UUID  `json:"id"`
	EmployeeID      uuid.UUID  `json:"employee_id"`
	IDProofURL      string     `json:"id_proof_url"`
	AddressProofURL string     `json:"address_proof_url"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt      *string    `json:"reviewed_at,omitempty"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

// AdminResponse includes employee and user metadata for admin views.
type AdminResponse struct {
	Response
	EmployeeDisplayName *string `json:"employee_display_name,omitempty"`
	UserName            string  `json:"user_name"`
	UserEmail           string  `json:"user_email"`
}

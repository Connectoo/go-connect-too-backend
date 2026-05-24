package kyc

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EmployeeLookup resolves an employee profile for an authenticated user.
type EmployeeLookup interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (employeeID uuid.UUID, err error)
}

// RecordStore loads and submits KYC records.
type RecordStore interface {
	GetByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*Record, error)
	Submit(ctx context.Context, employeeID uuid.UUID, idProofURL, addressProofURL string, at time.Time) (*Record, error)
}

// Service handles employee KYC business logic.
type Service struct {
	employees EmployeeLookup
	records   RecordStore
	now       func() time.Time
}

// NewService creates a KYC service.
func NewService(employees EmployeeLookup, records RecordStore) *Service {
	return &Service{
		employees: employees,
		records:   records,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// Submit creates or resubmits the authenticated employee's KYC documents.
func (s *Service) Submit(ctx context.Context, userID uuid.UUID, req SubmitRequest) (*Response, error) {
	idProofURL, err := validateDocumentURL(req.IDProofURL, "id_proof_url")
	if err != nil {
		return nil, err
	}
	addressProofURL, err := validateDocumentURL(req.AddressProofURL, "address_proof_url")
	if err != nil {
		return nil, err
	}

	employeeID, err := s.employees.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	record, err := s.records.Submit(ctx, employeeID, idProofURL, addressProofURL, s.now())
	if err != nil {
		return nil, err
	}

	return toResponse(record), nil
}

// Get returns the authenticated employee's KYC record.
func (s *Service) Get(ctx context.Context, userID uuid.UUID) (*Response, error) {
	employeeID, err := s.employees.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	record, err := s.records.GetByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	return toResponse(record), nil
}

func validateDocumentURL(raw, field string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("%w: %s is required", ErrValidation, field)
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%w: %s must be a valid URL", ErrValidation, field)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%w: %s must use http or https", ErrValidation, field)
	}

	return trimmed, nil
}

func toResponse(record *Record) *Response {
	return &Response{
		ID:              record.ID,
		EmployeeID:      record.EmployeeID,
		IDProofURL:      record.IDProofURL,
		AddressProofURL: record.AddressProofURL,
		Status:          record.Status,
		RejectionReason: record.RejectionReason,
		CreatedAt:       record.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       record.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

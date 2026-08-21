package kyc

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/events"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/storage"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

// EmployeeLookup resolves an employee profile for an authenticated user.
type EmployeeLookup interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (employeeID uuid.UUID, err error)
}

// EmployeeUserLookup resolves the user id for an employee profile.
type EmployeeUserLookup interface {
	GetUserIDByEmployeeID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error)
}

// RecordStore loads and submits KYC records.
type RecordStore interface {
	GetByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*Record, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Record, error)
	Submit(ctx context.Context, employeeID uuid.UUID, idProofURL, addressProofURL string, at time.Time) (*Record, error)
	ListForAdmin(ctx context.Context, filter AdminListFilter) ([]AdminListItem, int, error)
	GetAdminByID(ctx context.Context, id uuid.UUID) (*AdminListItem, error)
	Approve(ctx context.Context, id, reviewerID uuid.UUID, at time.Time) (*Record, error)
	Reject(ctx context.Context, id, reviewerID uuid.UUID, reason string, at time.Time) (*Record, error)
}

// VerificationSync updates employee profile verification status.
type VerificationSync interface {
	UpdateVerificationStatus(ctx context.Context, employeeID uuid.UUID, status string, at time.Time) error
}

// FileResolver resolves owned uploaded file keys for a user.
type FileResolver interface {
	ResolveFileURL(ctx context.Context, userID, fileID uuid.UUID) (string, error)
}

// EventPublisher publishes platform events.
type EventPublisher interface {
	Publish(ctx context.Context, event events.Event)
}

// Service handles employee KYC business logic.
type Service struct {
	employees    EmployeeLookup
	employeeUser EmployeeUserLookup
	records      RecordStore
	verification VerificationSync
	files        FileResolver
	events       EventPublisher
	now          func() time.Time
}

// NewService creates a KYC service.
func NewService(employees EmployeeLookup, records RecordStore, opts ...ServiceOption) *Service {
	s := &Service{
		employees: employees,
		records:   records,
		now:       func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ServiceOption configures optional KYC service dependencies.
type ServiceOption func(*Service)

// WithVerificationSync configures employee verification status sync.
func WithVerificationSync(sync VerificationSync) ServiceOption {
	return func(s *Service) { s.verification = sync }
}

// WithEmployeeUserLookup configures employee-to-user resolution.
func WithEmployeeUserLookup(lookup EmployeeUserLookup) ServiceOption {
	return func(s *Service) { s.employeeUser = lookup }
}

// WithFileResolver configures uploaded file resolution.
func WithFileResolver(resolver FileResolver) ServiceOption {
	return func(s *Service) { s.files = resolver }
}

// WithEventPublisher configures platform event publishing.
func WithEventPublisher(publisher EventPublisher) ServiceOption {
	return func(s *Service) { s.events = publisher }
}

// Submit creates or resubmits the authenticated employee's KYC documents.
func (s *Service) Submit(ctx context.Context, userID uuid.UUID, req SubmitRequest) (*Response, error) {
	idProofURL, err := s.resolveDocument(ctx, userID, req.IDProofURL, req.IDProofFileID, "id_proof_url")
	if err != nil {
		return nil, err
	}
	addressProofURL, err := s.resolveDocument(ctx, userID, req.AddressProofURL, req.AddressProofFileID, "address_proof_url")
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

// ListForAdmin returns paginated KYC records for admin review.
func (s *Service) ListForAdmin(ctx context.Context, status string, page pagination.Params) (pagination.Result[AdminResponse], error) {
	items, total, err := s.records.ListForAdmin(ctx, AdminListFilter{
		Status: status,
		Offset: page.Offset(),
		Limit:  page.Limit,
	})
	if err != nil {
		return pagination.Result[AdminResponse]{}, err
	}

	out := make([]AdminResponse, 0, len(items))
	for i := range items {
		out = append(out, toAdminResponse(&items[i]))
	}
	return pagination.NewResult(out, page, total), nil
}

// GetForAdmin returns one KYC record for admin review.
func (s *Service) GetForAdmin(ctx context.Context, id uuid.UUID) (*AdminResponse, error) {
	item, err := s.records.GetAdminByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := toAdminResponse(item)
	return &res, nil
}

// Approve approves a pending KYC submission and syncs employee verification status.
func (s *Service) Approve(ctx context.Context, reviewerID, kycID uuid.UUID) (*AdminResponse, error) {
	record, err := s.records.Approve(ctx, kycID, reviewerID, s.now())
	if err != nil {
		return nil, err
	}

	if s.verification != nil {
		if err := s.verification.UpdateVerificationStatus(ctx, record.EmployeeID, employees.VerificationApproved, s.now()); err != nil {
			return nil, err
		}
	}

	s.publishKYCEvent(ctx, record, events.TypeKYCApproved, nil)

	item, err := s.records.GetAdminByID(ctx, kycID)
	if err != nil {
		return nil, err
	}
	res := toAdminResponse(item)
	return &res, nil
}

// Reject rejects a pending KYC submission with a reason.
func (s *Service) Reject(ctx context.Context, reviewerID, kycID uuid.UUID, req RejectRequest) (*AdminResponse, error) {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, fmt.Errorf("%w: reason is required", ErrValidation)
	}

	record, err := s.records.Reject(ctx, kycID, reviewerID, reason, s.now())
	if err != nil {
		return nil, err
	}

	s.publishKYCEvent(ctx, record, events.TypeKYCRejected, &reason)

	item, err := s.records.GetAdminByID(ctx, kycID)
	if err != nil {
		return nil, err
	}
	res := toAdminResponse(item)
	return &res, nil
}

func (s *Service) resolveDocument(ctx context.Context, userID uuid.UUID, rawURL string, fileID *uuid.UUID, field string) (string, error) {
	if fileID != nil {
		if s.files == nil {
			return "", fmt.Errorf("%w: file storage is not configured", ErrValidation)
		}
		url, err := s.files.ResolveFileURL(ctx, userID, *fileID)
		if err != nil {
			return "", mapStorageFileError(err)
		}
		return url, nil
	}
	return validateDocumentURL(rawURL, field)
}

func mapStorageFileError(err error) error {
	if errors.Is(err, storage.ErrNotFound) {
		return ErrFileNotFound
	}
	if errors.Is(err, storage.ErrForbidden) {
		return ErrFileNotOwned
	}
	return err
}

func (s *Service) publishKYCEvent(ctx context.Context, record *Record, eventType events.Type, reason *string) {
	if s.events == nil || s.employeeUser == nil {
		return
	}

	userID, err := s.employeeUser.GetUserIDByEmployeeID(ctx, record.EmployeeID)
	if err != nil {
		return
	}

	payload := map[string]any{
		"user_id":     userID.String(),
		"kyc_id":      record.ID.String(),
		"employee_id": record.EmployeeID.String(),
	}
	if reason != nil {
		payload["reason"] = *reason
	}

	s.events.Publish(ctx, events.Event{
		Type:    eventType,
		Payload: payload,
	})
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
	res := &Response{
		ID:              record.ID,
		EmployeeID:      record.EmployeeID,
		IDProofURL:      record.IDProofURL,
		AddressProofURL: record.AddressProofURL,
		Status:          record.Status,
		RejectionReason: record.RejectionReason,
		ReviewedBy:      record.ReviewedBy,
		CreatedAt:       record.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       record.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if record.ReviewedAt != nil {
		formatted := record.ReviewedAt.UTC().Format(time.RFC3339)
		res.ReviewedAt = &formatted
	}
	return res
}

func toAdminResponse(item *AdminListItem) AdminResponse {
	return AdminResponse{
		Response:            *toResponse(&item.Record),
		EmployeeDisplayName: item.EmployeeDisplayName,
		UserName:            item.UserName,
		UserEmail:           item.UserEmail,
	}
}

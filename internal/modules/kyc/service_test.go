package kyc

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

type mockEmployees struct {
	employeeID uuid.UUID
}

func (m mockEmployees) GetByUserID(context.Context, uuid.UUID) (uuid.UUID, error) {
	return m.employeeID, nil
}

type mockEmployeeUser struct {
	userID uuid.UUID
}

func (m mockEmployeeUser) GetUserIDByEmployeeID(context.Context, uuid.UUID) (uuid.UUID, error) {
	return m.userID, nil
}

type mockVerification struct {
	status string
	called bool
}

func (m *mockVerification) UpdateVerificationStatus(_ context.Context, _ uuid.UUID, status string, _ time.Time) error {
	m.called = true
	m.status = status
	return nil
}

type mockRecordStore struct {
	record      *Record
	latest      *Record
	approveErr  error
	rejectErr   error
	approveCall bool
	rejectCall  bool
}

func (m *mockRecordStore) GetByEmployeeID(context.Context, uuid.UUID) (*Record, error) {
	return m.record, nil
}

func (m *mockRecordStore) GetByID(context.Context, uuid.UUID) (*Record, error) {
	return m.record, nil
}

func (m *mockRecordStore) Submit(context.Context, uuid.UUID, string, string, time.Time) (*Record, error) {
	return m.record, nil
}

func (m *mockRecordStore) ListForAdmin(context.Context, AdminListFilter) ([]AdminListItem, int, error) {
	return nil, 0, nil
}

func (m *mockRecordStore) GetAdminByID(_ context.Context, id uuid.UUID) (*AdminListItem, error) {
	source := m.record
	if m.latest != nil {
		source = m.latest
	}
	if source == nil {
		return nil, ErrNotFound
	}
	return &AdminListItem{
		Record:    *source,
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}, nil
}

func (m *mockRecordStore) Approve(_ context.Context, id, reviewerID uuid.UUID, at time.Time) (*Record, error) {
	m.approveCall = true
	if m.approveErr != nil {
		return nil, m.approveErr
	}
	approved := *m.record
	approved.Status = StatusApproved
	approved.ReviewedBy = &reviewerID
	approved.ReviewedAt = &at
	m.latest = &approved
	return &approved, nil
}

func (m *mockRecordStore) Reject(_ context.Context, id, reviewerID uuid.UUID, reason string, at time.Time) (*Record, error) {
	m.rejectCall = true
	if m.rejectErr != nil {
		return nil, m.rejectErr
	}
	rejected := *m.record
	rejected.Status = StatusRejected
	rejected.RejectionReason = &reason
	rejected.ReviewedBy = &reviewerID
	rejected.ReviewedAt = &at
	m.latest = &rejected
	return &rejected, nil
}

func TestService_Approve_syncsVerificationAndReturnsRecord(t *testing.T) {
	employeeID := uuid.New()
	kycID := uuid.New()
	reviewerID := uuid.New()
	store := &mockRecordStore{
		record: &Record{
			ID:         kycID,
			EmployeeID: employeeID,
			Status:     StatusPending,
		},
	}
	verification := &mockVerification{}

	svc := NewService(
		mockEmployees{employeeID: employeeID},
		store,
		WithVerificationSync(verification),
		WithEmployeeUserLookup(mockEmployeeUser{userID: uuid.New()}),
	)

	res, err := svc.Approve(context.Background(), reviewerID, kycID)
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if !store.approveCall {
		t.Fatal("Approve() did not call repository Approve")
	}
	if !verification.called {
		t.Fatal("Approve() did not sync employee verification status")
	}
	if verification.status != employees.VerificationApproved {
		t.Fatalf("verification status = %q, want %q", verification.status, employees.VerificationApproved)
	}
	if res.Status != StatusApproved {
		t.Fatalf("response status = %q, want %q", res.Status, StatusApproved)
	}
}

func TestService_Reject_requiresReason(t *testing.T) {
	svc := NewService(mockEmployees{employeeID: uuid.New()}, &mockRecordStore{
		record: &Record{ID: uuid.New(), Status: StatusPending},
	})

	_, err := svc.Reject(context.Background(), uuid.New(), uuid.New(), RejectRequest{Reason: "  "})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("Reject() error = %v, want ErrValidation", err)
	}
}

func TestService_Reject_setsRejectedStatus(t *testing.T) {
	kycID := uuid.New()
	store := &mockRecordStore{
		record: &Record{ID: kycID, EmployeeID: uuid.New(), Status: StatusPending},
	}
	svc := NewService(mockEmployees{employeeID: uuid.New()}, store,
		WithEmployeeUserLookup(mockEmployeeUser{userID: uuid.New()}),
	)

	res, err := svc.Reject(context.Background(), uuid.New(), kycID, RejectRequest{Reason: "blurry documents"})
	if err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if !store.rejectCall {
		t.Fatal("Reject() did not call repository Reject")
	}
	if res.Status != StatusRejected {
		t.Fatalf("response status = %q, want %q", res.Status, StatusRejected)
	}
	if res.RejectionReason == nil || strings.TrimSpace(*res.RejectionReason) == "" {
		t.Fatal("Reject() missing rejection reason")
	}
}

func TestService_Approve_invalidStatus(t *testing.T) {
	store := &mockRecordStore{
		record:     &Record{ID: uuid.New(), Status: StatusApproved},
		approveErr: ErrInvalidStatus,
	}
	svc := NewService(mockEmployees{employeeID: uuid.New()}, store)

	_, err := svc.Approve(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("Approve() error = %v, want ErrInvalidStatus", err)
	}
}

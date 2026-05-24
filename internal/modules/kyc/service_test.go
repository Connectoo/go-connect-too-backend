package kyc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockEmployeeLookup struct {
	byUserID map[uuid.UUID]uuid.UUID
	err      error
}

func (m *mockEmployeeLookup) GetByUserID(_ context.Context, userID uuid.UUID) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	employeeID, ok := m.byUserID[userID]
	if !ok {
		return uuid.Nil, ErrNotFound
	}
	return employeeID, nil
}

type mockRecordStore struct {
	byEmployeeID map[uuid.UUID]*Record
}

func newMockRecordStore() *mockRecordStore {
	return &mockRecordStore{byEmployeeID: make(map[uuid.UUID]*Record)}
}

func (m *mockRecordStore) GetByEmployeeID(_ context.Context, employeeID uuid.UUID) (*Record, error) {
	record, ok := m.byEmployeeID[employeeID]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *record
	return &copy, nil
}

func (m *mockRecordStore) Submit(_ context.Context, employeeID uuid.UUID, idProofURL, addressProofURL string, at time.Time) (*Record, error) {
	if existing, ok := m.byEmployeeID[employeeID]; ok {
		if existing.Status != StatusRejected {
			return nil, ErrAlreadyExists
		}
		updated := *existing
		updated.IDProofURL = idProofURL
		updated.AddressProofURL = addressProofURL
		updated.Status = StatusPending
		updated.RejectionReason = nil
		updated.UpdatedAt = at
		m.byEmployeeID[employeeID] = &updated
		copy := updated
		return &copy, nil
	}

	record := &Record{
		ID:              uuid.New(),
		EmployeeID:      employeeID,
		IDProofURL:      idProofURL,
		AddressProofURL: addressProofURL,
		Status:          StatusPending,
		CreatedAt:       at,
		UpdatedAt:       at,
	}
	m.byEmployeeID[employeeID] = record
	copy := *record
	return &copy, nil
}

func newTestService(t *testing.T, employees EmployeeLookup, records RecordStore) *Service {
	t.Helper()
	svc := NewService(employees, records)
	svc.now = func() time.Time {
		return time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	}
	return svc
}

func TestSubmitSuccess(t *testing.T) {
	userID := uuid.New()
	employeeID := uuid.New()
	svc := newTestService(t, &mockEmployeeLookup{
		byUserID: map[uuid.UUID]uuid.UUID{userID: employeeID},
	}, newMockRecordStore())

	res, err := svc.Submit(context.Background(), userID, SubmitRequest{
		IDProofURL:      "https://cdn.example.com/id.pdf",
		AddressProofURL: "https://cdn.example.com/address.pdf",
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if res.EmployeeID != employeeID || res.Status != StatusPending {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestSubmitValidation(t *testing.T) {
	userID := uuid.New()
	svc := newTestService(t, &mockEmployeeLookup{
		byUserID: map[uuid.UUID]uuid.UUID{userID: uuid.New()},
	}, newMockRecordStore())

	_, err := svc.Submit(context.Background(), userID, SubmitRequest{
		IDProofURL:      "not-a-url",
		AddressProofURL: "https://cdn.example.com/address.pdf",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want %v", err, ErrValidation)
	}
}

func TestSubmitAlreadyExists(t *testing.T) {
	userID := uuid.New()
	employeeID := uuid.New()
	store := newMockRecordStore()
	store.byEmployeeID[employeeID] = &Record{
		ID:              uuid.New(),
		EmployeeID:      employeeID,
		IDProofURL:      "https://cdn.example.com/id-old.pdf",
		AddressProofURL: "https://cdn.example.com/address-old.pdf",
		Status:          StatusPending,
	}

	svc := newTestService(t, &mockEmployeeLookup{
		byUserID: map[uuid.UUID]uuid.UUID{userID: employeeID},
	}, store)

	_, err := svc.Submit(context.Background(), userID, SubmitRequest{
		IDProofURL:      "https://cdn.example.com/id.pdf",
		AddressProofURL: "https://cdn.example.com/address.pdf",
	})
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("error = %v, want %v", err, ErrAlreadyExists)
	}
}

func TestSubmitResubmitAfterRejection(t *testing.T) {
	userID := uuid.New()
	employeeID := uuid.New()
	rejectionReason := "blurry document"
	store := newMockRecordStore()
	store.byEmployeeID[employeeID] = &Record{
		ID:              uuid.New(),
		EmployeeID:      employeeID,
		IDProofURL:      "https://cdn.example.com/id-old.pdf",
		AddressProofURL: "https://cdn.example.com/address-old.pdf",
		Status:          StatusRejected,
		RejectionReason: &rejectionReason,
	}

	svc := newTestService(t, &mockEmployeeLookup{
		byUserID: map[uuid.UUID]uuid.UUID{userID: employeeID},
	}, store)

	res, err := svc.Submit(context.Background(), userID, SubmitRequest{
		IDProofURL:      "https://cdn.example.com/id-new.pdf",
		AddressProofURL: "https://cdn.example.com/address-new.pdf",
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if res.Status != StatusPending || res.RejectionReason != nil {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestGetSuccess(t *testing.T) {
	userID := uuid.New()
	employeeID := uuid.New()
	store := newMockRecordStore()
	store.byEmployeeID[employeeID] = &Record{
		ID:              uuid.New(),
		EmployeeID:      employeeID,
		IDProofURL:      "https://cdn.example.com/id.pdf",
		AddressProofURL: "https://cdn.example.com/address.pdf",
		Status:          StatusPending,
		CreatedAt:       time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC),
	}

	svc := newTestService(t, &mockEmployeeLookup{
		byUserID: map[uuid.UUID]uuid.UUID{userID: employeeID},
	}, store)

	res, err := svc.Get(context.Background(), userID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if res.EmployeeID != employeeID || res.Status != StatusPending {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestGetNotFound(t *testing.T) {
	userID := uuid.New()
	svc := newTestService(t, &mockEmployeeLookup{
		byUserID: map[uuid.UUID]uuid.UUID{userID: uuid.New()},
	}, newMockRecordStore())

	_, err := svc.Get(context.Background(), userID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrNotFound)
	}
}

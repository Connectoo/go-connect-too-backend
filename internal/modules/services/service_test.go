package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

type mockProfileStore struct {
	profile *employees.Profile
	err     error
}

func (m *mockProfileStore) GetByUserID(_ context.Context, _ uuid.UUID) (*employees.Profile, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.profile, nil
}

type mockLimitStore struct {
	limit int
	err   error
}

func (m mockLimitStore) CurrentServiceLimit(_ context.Context, _ uuid.UUID, _ time.Time) (int, error) {
	return m.limit, m.err
}

func TestService_Create_enforcesSubscriptionLimit(t *testing.T) {
	categoryID := uuid.New()
	userID := uuid.New()
	employeeID := uuid.New()
	store := newMockStore(categoryID)
	store.services[uuid.New()] = EmployeeService{
		ID:              uuid.New(),
		EmployeeID:      employeeID,
		CategoryID:      categoryID,
		Title:           "House Cleaning",
		Price:           1200,
		DurationMinutes: 90,
		IsActive:        true,
	}

	svc := NewService(&mockProfileStore{profile: completeProfile(employeeID)}, store, mockLimitStore{limit: 1})
	_, err := svc.Create(context.Background(), userID, CreateServiceRequest{
		CategoryID:      categoryID,
		Title:           "Deep Cleaning",
		Price:           1500,
		DurationMinutes: 120,
		IsActive:        true,
	})
	if !errors.Is(err, ErrServiceLimit) {
		t.Fatalf("Create() error = %v, want ErrServiceLimit", err)
	}
}

type mockStore struct {
	categories map[uuid.UUID]bool
	services   map[uuid.UUID]EmployeeService
}

func newMockStore(categoryID uuid.UUID) *mockStore {
	return &mockStore{
		categories: map[uuid.UUID]bool{categoryID: true},
		services:   map[uuid.UUID]EmployeeService{},
	}
}

func (m *mockStore) CategoryExists(_ context.Context, categoryID uuid.UUID) (bool, error) {
	return m.categories[categoryID], nil
}

func (m *mockStore) ListPublicActive(_ context.Context, _ *uuid.UUID, _ int) ([]EmployeeService, error) {
	return []EmployeeService{}, nil
}

func (m *mockStore) GetPublicActiveByID(_ context.Context, _ uuid.UUID) (*EmployeeService, error) {
	return nil, ErrNotFound
}

func (m *mockStore) ListActiveByEmployeeProfileID(_ context.Context, _ uuid.UUID) ([]EmployeeService, error) {
	return []EmployeeService{}, nil
}

func (m *mockStore) ListByEmployeeID(_ context.Context, employeeID uuid.UUID) ([]EmployeeService, error) {
	out := []EmployeeService{}
	for _, service := range m.services {
		if service.EmployeeID == employeeID {
			out = append(out, service)
		}
	}
	return out, nil
}

func (m *mockStore) CountActiveByEmployeeID(_ context.Context, employeeID uuid.UUID) (int, error) {
	count := 0
	for _, service := range m.services {
		if service.EmployeeID == employeeID && service.IsActive {
			count++
		}
	}
	return count, nil
}

func (m *mockStore) CountActiveExcludingID(_ context.Context, employeeID, serviceID uuid.UUID) (int, error) {
	count := 0
	for _, service := range m.services {
		if service.EmployeeID == employeeID && service.ID != serviceID && service.IsActive {
			count++
		}
	}
	return count, nil
}

func (m *mockStore) Create(_ context.Context, service *EmployeeService) (*EmployeeService, error) {
	copy := *service
	m.services[copy.ID] = copy
	return &copy, nil
}

func (m *mockStore) Update(_ context.Context, employeeID, serviceID uuid.UUID, service *EmployeeService, at time.Time) (*EmployeeService, error) {
	existing, ok := m.services[serviceID]
	if !ok || existing.EmployeeID != employeeID {
		return nil, ErrNotFound
	}

	existing.CategoryID = service.CategoryID
	existing.Title = service.Title
	existing.Description = service.Description
	existing.Price = service.Price
	existing.DurationMinutes = service.DurationMinutes
	existing.IsActive = service.IsActive
	existing.UpdatedAt = at
	m.services[serviceID] = existing
	return &existing, nil
}

func (m *mockStore) Delete(_ context.Context, employeeID, serviceID uuid.UUID) error {
	existing, ok := m.services[serviceID]
	if !ok || existing.EmployeeID != employeeID {
		return ErrNotFound
	}
	delete(m.services, serviceID)
	return nil
}

func (m *mockStore) UpdateStatus(_ context.Context, employeeID, serviceID uuid.UUID, isActive bool, at time.Time) (*EmployeeService, error) {
	existing, ok := m.services[serviceID]
	if !ok || existing.EmployeeID != employeeID {
		return nil, ErrNotFound
	}
	existing.IsActive = isActive
	existing.UpdatedAt = at
	m.services[serviceID] = existing
	return &existing, nil
}

func completeProfile(employeeID uuid.UUID) *employees.Profile {
	displayName := "Cleaner"
	phone := "+15551234567"
	return &employees.Profile{
		ID:          employeeID,
		DisplayName: &displayName,
		Phone:       &phone,
	}
}

func incompleteProfile(employeeID uuid.UUID) *employees.Profile {
	return &employees.Profile{ID: employeeID}
}

func TestService_Create(t *testing.T) {
	categoryID := uuid.New()
	userID := uuid.New()
	employeeID := uuid.New()

	tests := []struct {
		name    string
		profile *employees.Profile
		req     CreateServiceRequest
		wantErr error
	}{
		{
			name:    "inactive service allowed with incomplete profile",
			profile: incompleteProfile(employeeID),
			req: CreateServiceRequest{
				CategoryID:      categoryID,
				Title:           "House Cleaning",
				Price:           1200,
				DurationMinutes: 90,
				IsActive:        false,
			},
		},
		{
			name:    "active service requires completed profile",
			profile: incompleteProfile(employeeID),
			req: CreateServiceRequest{
				CategoryID:      categoryID,
				Title:           "House Cleaning",
				Price:           1200,
				DurationMinutes: 90,
				IsActive:        true,
			},
			wantErr: ErrProfileIncomplete,
		},
		{
			name:    "completed profile can create active service",
			profile: completeProfile(employeeID),
			req: CreateServiceRequest{
				CategoryID:      categoryID,
				Title:           "House Cleaning",
				Price:           1200,
				DurationMinutes: 90,
				IsActive:        true,
			},
		},
		{
			name:    "missing title fails validation",
			profile: completeProfile(employeeID),
			req: CreateServiceRequest{
				CategoryID:      categoryID,
				Title:           "   ",
				Price:           1200,
				DurationMinutes: 90,
			},
			wantErr: ErrValidation,
		},
		{
			name:    "unknown category rejected",
			profile: completeProfile(employeeID),
			req: CreateServiceRequest{
				CategoryID:      uuid.New(),
				Title:           "House Cleaning",
				Price:           1200,
				DurationMinutes: 90,
			},
			wantErr: ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockProfileStore{profile: tt.profile}, newMockStore(categoryID))
			svc.now = func() time.Time { return time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC) }

			res, err := svc.Create(context.Background(), userID, tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && res == nil {
				t.Fatal("Create() response is nil")
			}
		})
	}
}

func TestService_UpdateStatus(t *testing.T) {
	categoryID := uuid.New()
	userID := uuid.New()
	employeeID := uuid.New()
	serviceID := uuid.New()
	store := newMockStore(categoryID)
	store.services[serviceID] = EmployeeService{
		ID:              serviceID,
		EmployeeID:      employeeID,
		CategoryID:      categoryID,
		Title:           "House Cleaning",
		Price:           1200,
		DurationMinutes: 90,
		IsActive:        false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	svc := NewService(&mockProfileStore{profile: incompleteProfile(employeeID)}, store)
	_, err := svc.UpdateStatus(context.Background(), userID, serviceID, UpdateServiceStatusRequest{IsActive: true})
	if !errors.Is(err, ErrProfileIncomplete) {
		t.Fatalf("UpdateStatus() error = %v, want ErrProfileIncomplete", err)
	}

	svc = NewService(&mockProfileStore{profile: completeProfile(employeeID)}, store)
	res, err := svc.UpdateStatus(context.Background(), userID, serviceID, UpdateServiceStatusRequest{IsActive: true})
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if !res.IsActive {
		t.Fatal("IsActive = false, want true")
	}
}

func TestService_Delete_notOwned(t *testing.T) {
	categoryID := uuid.New()
	userID := uuid.New()
	employeeID := uuid.New()
	otherEmployeeID := uuid.New()
	serviceID := uuid.New()
	store := newMockStore(categoryID)
	store.services[serviceID] = EmployeeService{
		ID:              serviceID,
		EmployeeID:      otherEmployeeID,
		CategoryID:      categoryID,
		Title:           "House Cleaning",
		Price:           1200,
		DurationMinutes: 90,
	}

	svc := NewService(&mockProfileStore{profile: completeProfile(employeeID)}, store)
	err := svc.Delete(context.Background(), userID, serviceID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete() error = %v, want ErrNotFound", err)
	}
}

package availability

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

type mockStore struct {
	slots map[uuid.UUID]Availability
}

func newMockStore() *mockStore {
	return &mockStore{slots: map[uuid.UUID]Availability{}}
}

func (m *mockStore) ListByEmployeeID(_ context.Context, employeeID uuid.UUID) ([]Availability, error) {
	out := []Availability{}
	for _, slot := range m.slots {
		if slot.EmployeeID == employeeID {
			out = append(out, slot)
		}
	}
	return out, nil
}

func (m *mockStore) Create(_ context.Context, slot *Availability) (*Availability, error) {
	for _, existing := range m.slots {
		if existing.EmployeeID == slot.EmployeeID && existing.DayOfWeek == slot.DayOfWeek && overlaps(existing, *slot) {
			return nil, ErrOverlap
		}
	}
	copy := *slot
	m.slots[copy.ID] = copy
	return &copy, nil
}

func (m *mockStore) Update(_ context.Context, employeeID, slotID uuid.UUID, slot *Availability, at time.Time) (*Availability, error) {
	existing, ok := m.slots[slotID]
	if !ok || existing.EmployeeID != employeeID {
		return nil, ErrNotFound
	}
	for id, other := range m.slots {
		if id == slotID {
			continue
		}
		if other.EmployeeID == employeeID && other.DayOfWeek == slot.DayOfWeek && overlaps(other, Availability{StartTime: slot.StartTime, EndTime: slot.EndTime}) {
			return nil, ErrOverlap
		}
	}
	existing.DayOfWeek = slot.DayOfWeek
	existing.StartTime = slot.StartTime
	existing.EndTime = slot.EndTime
	existing.IsAvailable = slot.IsAvailable
	existing.UpdatedAt = at
	m.slots[slotID] = existing
	return &existing, nil
}

func (m *mockStore) Delete(_ context.Context, employeeID, slotID uuid.UUID) error {
	existing, ok := m.slots[slotID]
	if !ok || existing.EmployeeID != employeeID {
		return ErrNotFound
	}
	delete(m.slots, slotID)
	return nil
}

func overlaps(a, b Availability) bool {
	return a.StartTime.Before(b.EndTime) && b.StartTime.Before(a.EndTime)
}

func mustTime(t *testing.T, value string) TimeOfDay {
	t.Helper()
	parsed, err := ParseTimeOfDay(value)
	if err != nil {
		t.Fatalf("ParseTimeOfDay(%q) error = %v", value, err)
	}
	return parsed
}

func profileFor(employeeID uuid.UUID) *employees.Profile {
	return &employees.Profile{ID: employeeID, UserID: uuid.New()}
}

func TestService_Create(t *testing.T) {
	employeeID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name    string
		seed    func(store *mockStore)
		req     CreateAvailabilityRequest
		wantErr error
	}{
		{
			name: "success",
			req: CreateAvailabilityRequest{
				DayOfWeek: 1,
				StartTime: mustTime(t, "09:00"),
				EndTime:   mustTime(t, "12:00"),
			},
		},
		{
			name: "invalid day rejected",
			req: CreateAvailabilityRequest{
				DayOfWeek: 7,
				StartTime: mustTime(t, "09:00"),
				EndTime:   mustTime(t, "12:00"),
			},
			wantErr: ErrValidation,
		},
		{
			name: "start not before end rejected",
			req: CreateAvailabilityRequest{
				DayOfWeek: 1,
				StartTime: mustTime(t, "12:00"),
				EndTime:   mustTime(t, "12:00"),
			},
			wantErr: ErrValidation,
		},
		{
			name: "overlap rejected",
			seed: func(store *mockStore) {
				id := uuid.New()
				store.slots[id] = Availability{
					ID:         id,
					EmployeeID: employeeID,
					DayOfWeek:  1,
					StartTime:  mustTime(t, "09:00"),
					EndTime:    mustTime(t, "11:00"),
				}
			},
			req: CreateAvailabilityRequest{
				DayOfWeek: 1,
				StartTime: mustTime(t, "10:00"),
				EndTime:   mustTime(t, "12:00"),
			},
			wantErr: ErrOverlap,
		},
		{
			name: "adjacent slot allowed",
			seed: func(store *mockStore) {
				id := uuid.New()
				store.slots[id] = Availability{
					ID:         id,
					EmployeeID: employeeID,
					DayOfWeek:  1,
					StartTime:  mustTime(t, "09:00"),
					EndTime:    mustTime(t, "10:00"),
				}
			},
			req: CreateAvailabilityRequest{
				DayOfWeek: 1,
				StartTime: mustTime(t, "10:00"),
				EndTime:   mustTime(t, "11:00"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockStore()
			if tt.seed != nil {
				tt.seed(store)
			}
			svc := NewService(&mockProfileStore{profile: profileFor(employeeID)}, store)
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

func TestService_Update_notOwned(t *testing.T) {
	employeeID := uuid.New()
	otherEmployeeID := uuid.New()
	userID := uuid.New()
	slotID := uuid.New()

	store := newMockStore()
	store.slots[slotID] = Availability{
		ID:         slotID,
		EmployeeID: otherEmployeeID,
		DayOfWeek:  1,
		StartTime:  mustTime(t, "09:00"),
		EndTime:    mustTime(t, "10:00"),
	}

	svc := NewService(&mockProfileStore{profile: profileFor(employeeID)}, store)
	_, err := svc.Update(context.Background(), userID, slotID, UpdateAvailabilityRequest{
		DayOfWeek: 1,
		StartTime: mustTime(t, "10:00"),
		EndTime:   mustTime(t, "11:00"),
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestService_Update_success(t *testing.T) {
	employeeID := uuid.New()
	userID := uuid.New()
	slotID := uuid.New()

	store := newMockStore()
	store.slots[slotID] = Availability{
		ID:         slotID,
		EmployeeID: employeeID,
		DayOfWeek:  1,
		StartTime:  mustTime(t, "09:00"),
		EndTime:    mustTime(t, "10:00"),
	}

	svc := NewService(&mockProfileStore{profile: profileFor(employeeID)}, store)
	svc.now = func() time.Time { return time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC) }

	res, err := svc.Update(context.Background(), userID, slotID, UpdateAvailabilityRequest{
		DayOfWeek:   2,
		StartTime:   mustTime(t, "13:00"),
		EndTime:     mustTime(t, "14:00"),
		IsAvailable: true,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if res.DayOfWeek != 2 {
		t.Fatalf("DayOfWeek = %d, want 2", res.DayOfWeek)
	}
}

func TestService_Delete_notOwned(t *testing.T) {
	employeeID := uuid.New()
	otherEmployeeID := uuid.New()
	userID := uuid.New()
	slotID := uuid.New()

	store := newMockStore()
	store.slots[slotID] = Availability{
		ID:         slotID,
		EmployeeID: otherEmployeeID,
		DayOfWeek:  1,
		StartTime:  mustTime(t, "09:00"),
		EndTime:    mustTime(t, "10:00"),
	}

	svc := NewService(&mockProfileStore{profile: profileFor(employeeID)}, store)
	err := svc.Delete(context.Background(), userID, slotID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestParseTimeOfDay(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   bool
	}{
		{"valid", "09:30", false},
		{"midnight", "00:00", false},
		{"hour out of range", "24:00", true},
		{"minute out of range", "12:60", true},
		{"missing minute", "12", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTimeOfDay(tt.input)
			if (err != nil) != tt.err {
				t.Fatalf("ParseTimeOfDay(%q) err = %v, want err = %v", tt.input, err, tt.err)
			}
		})
	}
}

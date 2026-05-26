package employees

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockProfileStore struct {
	byUserID map[uuid.UUID]*Profile
	byID     map[uuid.UUID]*Profile
}

func newMockProfileStore() *mockProfileStore {
	return &mockProfileStore{
		byUserID: make(map[uuid.UUID]*Profile),
		byID:     make(map[uuid.UUID]*Profile),
	}
}

func (m *mockProfileStore) seed(profile *Profile) {
	copy := *profile
	if copy.Languages == nil {
		copy.Languages = []string{}
	}
	if copy.Skills == nil {
		copy.Skills = []string{}
	}
	m.byUserID[profile.UserID] = &copy
	m.byID[profile.ID] = &copy
}

func (m *mockProfileStore) GetByUserID(_ context.Context, userID uuid.UUID) (*Profile, error) {
	profile, ok := m.byUserID[userID]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *profile
	return &copy, nil
}

func (m *mockProfileStore) GetByID(_ context.Context, id uuid.UUID) (*Profile, error) {
	profile, ok := m.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *profile
	return &copy, nil
}

func (m *mockProfileStore) GetApprovedByID(_ context.Context, id uuid.UUID) (*Profile, error) {
	profile, ok := m.byID[id]
	if !ok || profile.VerificationStatus != VerificationApproved {
		return nil, ErrNotFound
	}
	copy := *profile
	return &copy, nil
}

func (m *mockProfileStore) UpdateByUserID(_ context.Context, userID uuid.UUID, profile *Profile, at time.Time) (*Profile, error) {
	existing, ok := m.byUserID[userID]
	if !ok {
		return nil, ErrNotFound
	}

	updated := *existing
	updated.DisplayName = profile.DisplayName
	updated.Phone = profile.Phone
	updated.Bio = profile.Bio
	updated.ExperienceYears = profile.ExperienceYears
	updated.ProfilePhotoURL = profile.ProfilePhotoURL
	updated.LocationText = profile.LocationText
	updated.Latitude = profile.Latitude
	updated.Longitude = profile.Longitude
	updated.ServiceAreaRadiusKm = profile.ServiceAreaRadiusKm
	updated.Languages = profile.Languages
	updated.Skills = profile.Skills
	updated.VerificationStatus = profile.VerificationStatus
	updated.UpdatedAt = at

	m.byUserID[userID] = &updated
	m.byID[updated.ID] = &updated
	copy := updated
	return &copy, nil
}

func (m *mockProfileStore) UpdateVerificationStatus(_ context.Context, id uuid.UUID, status string, at time.Time) (*Profile, error) {
	profile, ok := m.byID[id]
	if !ok {
		return nil, ErrNotFound
	}

	updated := *profile
	updated.VerificationStatus = status
	updated.UpdatedAt = at

	m.byID[id] = &updated
	m.byUserID[updated.UserID] = &updated
	copy := updated
	return &copy, nil
}

func newTestService(t *testing.T, store ProfileStore) *Service {
	t.Helper()
	svc := NewService(store)
	svc.now = func() time.Time {
		return time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	}
	return svc
}

func TestGetProfileSuccess(t *testing.T) {
	store := newMockProfileStore()
	userID := uuid.New()
	profileID := uuid.New()
	store.seed(&Profile{
		ID:                 profileID,
		UserID:             userID,
		VerificationStatus: VerificationPending,
		CreatedAt:          time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
		UpdatedAt:          time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
	})

	svc := newTestService(t, store)
	res, err := svc.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if res.ID != profileID || res.VerificationStatus != VerificationPending {
		t.Fatalf("unexpected profile: %+v", res)
	}
}

func TestGetProfileNotFound(t *testing.T) {
	svc := newTestService(t, newMockProfileStore())

	_, err := svc.GetProfile(context.Background(), uuid.New())
	if err != ErrNotFound {
		t.Fatalf("error = %v, want %v", err, ErrNotFound)
	}
}

func TestUpdateProfileValidation(t *testing.T) {
	store := newMockProfileStore()
	userID := uuid.New()
	store.seed(&Profile{
		ID:                 uuid.New(),
		UserID:             userID,
		VerificationStatus: VerificationApproved,
	})

	svc := newTestService(t, store)
	_, err := svc.UpdateProfile(context.Background(), userID, UpdateProfileRequest{})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want %v", err, ErrValidation)
	}
}

func TestUpdateProfileResetsVerification(t *testing.T) {
	store := newMockProfileStore()
	userID := uuid.New()
	store.seed(&Profile{
		ID:                 uuid.New(),
		UserID:             userID,
		VerificationStatus: VerificationApproved,
	})

	svc := newTestService(t, store)
	res, err := svc.UpdateProfile(context.Background(), userID, UpdateProfileRequest{
		DisplayName:     "Alex Worker",
		Phone:           "+15550100",
		ExperienceYears: 3,
		Languages:       []string{"English"},
		Skills:          []string{"Plumbing"},
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if res.DisplayName == nil || *res.DisplayName != "Alex Worker" {
		t.Fatalf("unexpected display name: %+v", res.DisplayName)
	}
	if res.VerificationStatus != VerificationPending {
		t.Fatalf("verification_status = %q, want %q", res.VerificationStatus, VerificationPending)
	}
}

func TestApproveProfileSuccess(t *testing.T) {
	store := newMockProfileStore()
	profileID := uuid.New()
	store.seed(&Profile{
		ID:                 profileID,
		UserID:             uuid.New(),
		VerificationStatus: VerificationPending,
	})

	svc := newTestService(t, store)
	res, err := svc.ApproveProfile(context.Background(), profileID)
	if err != nil {
		t.Fatalf("ApproveProfile() error = %v", err)
	}
	if res.VerificationStatus != VerificationApproved {
		t.Fatalf("verification_status = %q, want %q", res.VerificationStatus, VerificationApproved)
	}
}

func TestRejectProfileNotFound(t *testing.T) {
	svc := newTestService(t, newMockProfileStore())

	_, err := svc.RejectProfile(context.Background(), uuid.New())
	if err != ErrNotFound {
		t.Fatalf("error = %v, want %v", err, ErrNotFound)
	}
}

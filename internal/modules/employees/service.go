package employees

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProfileStore loads and updates employee profiles.
type ProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	UpdateByUserID(ctx context.Context, userID uuid.UUID, profile *Profile, at time.Time) (*Profile, error)
	UpdateVerificationStatus(ctx context.Context, id uuid.UUID, status string, at time.Time) (*Profile, error)
}

// Service handles employee profile business logic.
type Service struct {
	profiles ProfileStore
	now      func() time.Time
}

// NewService creates an employee profile service.
func NewService(profiles ProfileStore) *Service {
	return &Service{
		profiles: profiles,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

// GetProfile returns the authenticated employee's profile.
func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*ProfileResponse, error) {
	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toProfileResponse(profile), nil
}

// UpdateProfile updates the authenticated employee's profile and resets verification to pending.
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*ProfileResponse, error) {
	if err := validateUpdateProfile(req); err != nil {
		return nil, err
	}

	displayName := strings.TrimSpace(req.DisplayName)
	phone := strings.TrimSpace(req.Phone)

	updated, err := s.profiles.UpdateByUserID(ctx, userID, &Profile{
		DisplayName:         &displayName,
		Phone:               &phone,
		Bio:                 trimOptionalString(req.Bio),
		ExperienceYears:     req.ExperienceYears,
		ProfilePhotoURL:     trimOptionalString(req.ProfilePhotoURL),
		LocationText:        trimOptionalString(req.LocationText),
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		ServiceAreaRadiusKm: req.ServiceAreaRadiusKm,
		Languages:           normalizeStringList(req.Languages),
		Skills:              normalizeStringList(req.Skills),
		VerificationStatus:  VerificationPending,
	}, s.now())
	if err != nil {
		return nil, err
	}

	return toProfileResponse(updated), nil
}

// ApproveProfile marks an employee profile as approved.
func (s *Service) ApproveProfile(ctx context.Context, profileID uuid.UUID) (*ProfileResponse, error) {
	profile, err := s.profiles.UpdateVerificationStatus(ctx, profileID, VerificationApproved, s.now())
	if err != nil {
		return nil, err
	}
	return toProfileResponse(profile), nil
}

// RejectProfile marks an employee profile as rejected.
func (s *Service) RejectProfile(ctx context.Context, profileID uuid.UUID) (*ProfileResponse, error) {
	profile, err := s.profiles.UpdateVerificationStatus(ctx, profileID, VerificationRejected, s.now())
	if err != nil {
		return nil, err
	}
	return toProfileResponse(profile), nil
}

func validateUpdateProfile(req UpdateProfileRequest) error {
	if strings.TrimSpace(req.DisplayName) == "" {
		return fmt.Errorf("%w: display_name is required", ErrValidation)
	}
	if strings.TrimSpace(req.Phone) == "" {
		return fmt.Errorf("%w: phone is required", ErrValidation)
	}
	if req.ExperienceYears < 0 {
		return fmt.Errorf("%w: experience_years must be zero or greater", ErrValidation)
	}
	if req.ServiceAreaRadiusKm != nil && *req.ServiceAreaRadiusKm < 0 {
		return fmt.Errorf("%w: service_area_radius_km must be zero or greater", ErrValidation)
	}
	if req.Latitude != nil && (*req.Latitude < -90 || *req.Latitude > 90) {
		return fmt.Errorf("%w: latitude must be between -90 and 90", ErrValidation)
	}
	if req.Longitude != nil && (*req.Longitude < -180 || *req.Longitude > 180) {
		return fmt.Errorf("%w: longitude must be between -180 and 180", ErrValidation)
	}
	return nil
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return []string{}
	}
	return out
}

func toProfileResponse(profile *Profile) *ProfileResponse {
	return &ProfileResponse{
		ID:                  profile.ID,
		UserID:              profile.UserID,
		DisplayName:         profile.DisplayName,
		Phone:               profile.Phone,
		Bio:                 profile.Bio,
		ExperienceYears:     profile.ExperienceYears,
		ProfilePhotoURL:     profile.ProfilePhotoURL,
		LocationText:        profile.LocationText,
		Latitude:            profile.Latitude,
		Longitude:           profile.Longitude,
		ServiceAreaRadiusKm: profile.ServiceAreaRadiusKm,
		Languages:           profile.Languages,
		Skills:              profile.Skills,
		VerificationStatus:  profile.VerificationStatus,
		CreatedAt:           profile.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           profile.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

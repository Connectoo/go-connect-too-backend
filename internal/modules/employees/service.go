package employees

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/storage"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

// BadgeReader loads trust badges for public profiles.
type BadgeReader interface {
	ListForEmployee(ctx context.Context, employeeID uuid.UUID) ([]string, error)
}

// ProfileViewRecorder records public profile view events for analytics.
type ProfileViewRecorder interface {
	RecordProfileView(ctx context.Context, employeeID uuid.UUID)
}

// ProfileStore loads and updates employee profiles.
type ProfileStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	GetApprovedByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	UpdateByUserID(ctx context.Context, userID uuid.UUID, profile *Profile, at time.Time) (*Profile, error)
	UpdateVerificationStatus(ctx context.Context, id uuid.UUID, status string, at time.Time) (*Profile, error)
	ListAdmin(ctx context.Context, filter AdminListFilter) ([]AdminListItem, int, error)
	GetAdminByID(ctx context.Context, id uuid.UUID) (*AdminListItem, error)
}

// UserStatusStore updates linked user account status.
type UserStatusStore interface {
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, at time.Time) error
}

// KYCStatusChecker verifies employee KYC approval before profile approval.
type KYCStatusChecker interface {
	GetStatusByEmployeeID(ctx context.Context, employeeID uuid.UUID) (string, error)
}

// Service handles employee profile business logic.
type Service struct {
	profiles     ProfileStore
	users        UserStatusStore
	kyc          KYCStatusChecker
	files        ProfileFileResolver
	badges       BadgeReader
	profileViews ProfileViewRecorder
	now          func() time.Time
}

// ProfileFileResolver resolves owned uploaded file keys for profile photos.
type ProfileFileResolver interface {
	ResolveFileURL(ctx context.Context, userID, fileID uuid.UUID) (string, error)
}

// NewService creates an employee profile service.
func NewService(profiles ProfileStore, users ...UserStatusStore) *Service {
	var userStore UserStatusStore
	if len(users) > 0 {
		userStore = users[0]
	}
	return &Service{
		profiles: profiles,
		users:    userStore,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

// WithKYCChecker configures KYC status checks for profile approval.
func (s *Service) WithKYCChecker(checker KYCStatusChecker) *Service {
	s.kyc = checker
	return s
}

// WithProfileFiles configures uploaded file resolution for profile photos.
func (s *Service) WithProfileFiles(resolver ProfileFileResolver) *Service {
	s.files = resolver
	return s
}

// WithBadges configures badge loading for public profiles.
func (s *Service) WithBadges(badges BadgeReader) *Service {
	s.badges = badges
	return s
}

// WithProfileViews configures analytics profile view recording.
func (s *Service) WithProfileViews(recorder ProfileViewRecorder) *Service {
	s.profileViews = recorder
	return s
}

// GetPublicProfile returns an approved employee profile for marketplace browsing.
func (s *Service) GetPublicProfile(ctx context.Context, profileID uuid.UUID) (*PublicProfileResponse, error) {
	profile, err := s.profiles.GetApprovedByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	res := toPublicProfileResponse(profile)
	if s.badges != nil {
		badges, err := s.badges.ListForEmployee(ctx, profileID)
		if err != nil {
			return nil, err
		}
		res.Badges = badges
	}
	if s.profileViews != nil {
		s.profileViews.RecordProfileView(ctx, profileID)
	}
	return res, nil
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

	profilePhotoURL, err := s.resolveProfilePhoto(ctx, userID, req.ProfilePhotoURL, req.ProfilePhotoFileID)
	if err != nil {
		return nil, err
	}

	updated, err := s.profiles.UpdateByUserID(ctx, userID, &Profile{
		DisplayName:         &displayName,
		Phone:               &phone,
		Bio:                 trimOptionalString(req.Bio),
		ExperienceYears:     req.ExperienceYears,
		ProfilePhotoURL:     profilePhotoURL,
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
	if s.kyc != nil {
		status, err := s.kyc.GetStatusByEmployeeID(ctx, profileID)
		if err != nil {
			return nil, err
		}
		if status != "approved" {
			return nil, ErrKYCNotApproved
		}
	}

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

// ListAdmin returns paginated employee profiles for admin views.
func (s *Service) ListAdmin(ctx context.Context, verificationStatus, query string, page pagination.Params) (pagination.Result[AdminEmployeeResponse], error) {
	items, total, err := s.profiles.ListAdmin(ctx, AdminListFilter{
		VerificationStatus: verificationStatus,
		Query:              query,
		Offset:             page.Offset(),
		Limit:              page.Limit,
	})
	if err != nil {
		return pagination.Result[AdminEmployeeResponse]{}, err
	}

	out := make([]AdminEmployeeResponse, 0, len(items))
	for i := range items {
		out = append(out, toAdminEmployeeResponse(&items[i]))
	}
	return pagination.NewResult(out, page, total), nil
}

// GetAdmin returns an employee profile with user account fields.
func (s *Service) GetAdmin(ctx context.Context, profileID uuid.UUID) (*AdminEmployeeResponse, error) {
	item, err := s.profiles.GetAdminByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	res := toAdminEmployeeResponse(item)
	return &res, nil
}

// SuspendProfile suspends the linked user account for an employee profile.
func (s *Service) SuspendProfile(ctx context.Context, profileID uuid.UUID) (*AdminEmployeeResponse, error) {
	if s.users == nil {
		return nil, fmt.Errorf("%w: user status store not configured", ErrValidation)
	}

	profile, err := s.profiles.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if err := s.users.UpdateStatus(ctx, profile.UserID, users.StatusSuspended, s.now()); err != nil {
		return nil, err
	}
	return s.GetAdmin(ctx, profileID)
}

func toAdminEmployeeResponse(item *AdminListItem) AdminEmployeeResponse {
	return AdminEmployeeResponse{
		ProfileResponse: *toProfileResponse(&item.Profile),
		UserName:        item.UserName,
		UserEmail:       item.UserEmail,
		UserStatus:      item.UserStatus,
	}
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

func (s *Service) resolveProfilePhoto(ctx context.Context, userID uuid.UUID, rawURL *string, fileID *uuid.UUID) (*string, error) {
	if fileID != nil {
		if s.files == nil {
			return nil, fmt.Errorf("%w: file storage is not configured", ErrValidation)
		}
		url, err := s.files.ResolveFileURL(ctx, userID, *fileID)
		if err != nil {
			return nil, mapStorageFileError(err)
		}
		return &url, nil
	}
	return trimOptionalString(rawURL), nil
}

func mapStorageFileError(err error) error {
	if errors.Is(err, storage.ErrNotFound) {
		return fmt.Errorf("%w: uploaded file not found", ErrValidation)
	}
	if errors.Is(err, storage.ErrForbidden) {
		return fmt.Errorf("%w: uploaded file not owned by user", ErrValidation)
	}
	return err
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

func toPublicProfileResponse(profile *Profile) *PublicProfileResponse {
	return &PublicProfileResponse{
		ID:                  profile.ID,
		DisplayName:         profile.DisplayName,
		Bio:                 profile.Bio,
		ExperienceYears:     profile.ExperienceYears,
		ProfilePhotoURL:     profile.ProfilePhotoURL,
		LocationText:        profile.LocationText,
		Latitude:            profile.Latitude,
		Longitude:           profile.Longitude,
		ServiceAreaRadiusKm: profile.ServiceAreaRadiusKm,
		Languages:           profile.Languages,
		Skills:              profile.Skills,
		Rating:              profile.AverageRating,
		TotalReviews:        profile.TotalReviews,
	}
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

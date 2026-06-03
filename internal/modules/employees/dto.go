package employees

import "github.com/google/uuid"

// UpdateProfileRequest updates the authenticated employee profile.
type UpdateProfileRequest struct {
	DisplayName         string   `json:"display_name"`
	Phone               string   `json:"phone"`
	Bio                 *string  `json:"bio,omitempty"`
	ExperienceYears     int      `json:"experience_years"`
	ProfilePhotoURL     *string  `json:"profile_photo_url,omitempty"`
	LocationText        *string  `json:"location_text,omitempty"`
	Latitude            *float64 `json:"latitude,omitempty"`
	Longitude           *float64 `json:"longitude,omitempty"`
	ServiceAreaRadiusKm *float64 `json:"service_area_radius_km,omitempty"`
	Languages           []string `json:"languages"`
	Skills              []string `json:"skills"`
}

// ProfileResponse is the authenticated employee profile payload.
type ProfileResponse struct {
	ID                  uuid.UUID `json:"id"`
	UserID              uuid.UUID `json:"user_id"`
	DisplayName         *string   `json:"display_name,omitempty"`
	Phone               *string   `json:"phone,omitempty"`
	Bio                 *string   `json:"bio,omitempty"`
	ExperienceYears     int       `json:"experience_years"`
	ProfilePhotoURL     *string   `json:"profile_photo_url,omitempty"`
	LocationText        *string   `json:"location_text,omitempty"`
	Latitude            *float64  `json:"latitude,omitempty"`
	Longitude           *float64  `json:"longitude,omitempty"`
	ServiceAreaRadiusKm *float64  `json:"service_area_radius_km,omitempty"`
	Languages           []string  `json:"languages"`
	Skills              []string  `json:"skills"`
	VerificationStatus  string    `json:"verification_status"`
	CreatedAt           string    `json:"created_at"`
	UpdatedAt           string    `json:"updated_at"`
}

// PublicProfileResponse is the marketplace-visible employee profile.
type PublicProfileResponse struct {
	ID                  uuid.UUID `json:"id"`
	DisplayName         *string   `json:"display_name,omitempty"`
	Bio                 *string   `json:"bio,omitempty"`
	ExperienceYears     int       `json:"experience_years"`
	ProfilePhotoURL     *string   `json:"profile_photo_url,omitempty"`
	LocationText        *string   `json:"location_text,omitempty"`
	Latitude            *float64  `json:"latitude,omitempty"`
	Longitude           *float64  `json:"longitude,omitempty"`
	ServiceAreaRadiusKm *float64  `json:"service_area_radius_km,omitempty"`
	Languages           []string  `json:"languages"`
	Skills              []string  `json:"skills"`
	Rating              *float64  `json:"rating,omitempty"`
	TotalReviews        int       `json:"total_reviews"`
	Badges              []string  `json:"badges,omitempty"`
}

// AdminEmployeeResponse is an employee profile with linked user account fields.
type AdminEmployeeResponse struct {
	ProfileResponse
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
	UserStatus string `json:"user_status"`
}

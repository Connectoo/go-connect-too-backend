package employees

import (
	"time"

	"github.com/google/uuid"
)

const (
	VerificationPending  = "pending"
	VerificationApproved = "approved"
	VerificationRejected = "rejected"
)

// Profile is an employee workforce profile.
type Profile struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	DisplayName         *string
	Phone               *string
	Bio                 *string
	ExperienceYears     int
	ProfilePhotoURL     *string
	LocationText        *string
	Latitude            *float64
	Longitude           *float64
	ServiceAreaRadiusKm *float64
	Languages           []string
	Skills              []string
	VerificationStatus  string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

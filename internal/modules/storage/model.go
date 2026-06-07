package storage

import (
	"time"

	"github.com/google/uuid"
)

const (
	PurposeKYCIDProof      = "kyc_id_proof"
	PurposeKYCAddressProof = "kyc_address_proof"
	PurposeProfilePhoto    = "profile_photo"
)

// UploadedFile is a user-owned object stored in S3.
type UploadedFile struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ObjectKey   string
	ContentType string
	SizeBytes   *int64
	Purpose     string
	CreatedAt   time.Time
	DeletedAt   *time.Time
}

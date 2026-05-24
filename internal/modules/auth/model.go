package auth

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken is a persisted refresh session.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

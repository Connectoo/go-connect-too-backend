package customers

import (
	"time"

	"github.com/google/uuid"
)

// Profile links a customer account to marketplace-specific data (expanded in later phases).
type Profile struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

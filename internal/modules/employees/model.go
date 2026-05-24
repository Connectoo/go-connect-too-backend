package employees

import (
	"time"

	"github.com/google/uuid"
)

// Profile links an employee account to optional workforce metadata (expanded in later phases).
type Profile struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

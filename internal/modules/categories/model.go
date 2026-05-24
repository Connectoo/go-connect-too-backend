package categories

import (
	"time"

	"github.com/google/uuid"
)

// Category is a service marketplace category row.
type Category struct {
	ID          uuid.UUID
	Name        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

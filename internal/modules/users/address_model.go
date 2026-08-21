package users

import (
	"time"

	"github.com/google/uuid"
)

// Address is a saved customer address.
type Address struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Label       string
	AddressLine string
	City        string
	State       string
	Country     string
	Pincode     string
	Latitude    *float64
	Longitude   *float64
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

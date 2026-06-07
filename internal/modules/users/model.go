package users

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleCustomer = "customer"
	RoleEmployee = "employee"
	RoleAdmin    = "admin"

	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusSuspended = "suspended"
)

// User is a platform account.
type User struct {
	ID              uuid.UUID
	Name            string
	Email           string
	Phone           *string
	PasswordHash    string
	Role            string
	Status          string
	EmailVerifiedAt *time.Time
	DeactivatedAt   *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

package services

import (
	"time"

	"github.com/google/uuid"
)

// EmployeeService is a service listing owned by an employee profile.
type EmployeeService struct {
	ID              uuid.UUID
	EmployeeID      uuid.UUID
	CategoryID      uuid.UUID
	Title           string
	Description     *string
	Price           float64
	DurationMinutes int
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

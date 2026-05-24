package services

import "github.com/google/uuid"

// CreateServiceRequest creates an employee service listing.
type CreateServiceRequest struct {
	CategoryID      uuid.UUID `json:"category_id"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	Price           float64   `json:"price"`
	DurationMinutes int       `json:"duration_minutes"`
	IsActive        bool      `json:"is_active"`
}

// UpdateServiceRequest replaces editable service listing fields.
type UpdateServiceRequest struct {
	CategoryID      uuid.UUID `json:"category_id"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	Price           float64   `json:"price"`
	DurationMinutes int       `json:"duration_minutes"`
	IsActive        bool      `json:"is_active"`
}

// UpdateServiceStatusRequest activates or deactivates a service listing.
type UpdateServiceStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// ServiceResponse is the public employee service payload.
type ServiceResponse struct {
	ID              uuid.UUID `json:"id"`
	EmployeeID      uuid.UUID `json:"employee_id"`
	CategoryID      uuid.UUID `json:"category_id"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	Price           float64   `json:"price"`
	DurationMinutes int       `json:"duration_minutes"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

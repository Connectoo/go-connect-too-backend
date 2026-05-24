package availability

import "github.com/google/uuid"

// CreateAvailabilityRequest creates a weekly availability slot.
type CreateAvailabilityRequest struct {
	DayOfWeek   int       `json:"day_of_week"`
	StartTime   TimeOfDay `json:"start_time"`
	EndTime     TimeOfDay `json:"end_time"`
	IsAvailable *bool     `json:"is_available,omitempty"`
}

// UpdateAvailabilityRequest replaces editable availability slot fields.
type UpdateAvailabilityRequest struct {
	DayOfWeek   int       `json:"day_of_week"`
	StartTime   TimeOfDay `json:"start_time"`
	EndTime     TimeOfDay `json:"end_time"`
	IsAvailable bool      `json:"is_available"`
}

// AvailabilityResponse is the public availability slot payload.
type AvailabilityResponse struct {
	ID          uuid.UUID `json:"id"`
	EmployeeID  uuid.UUID `json:"employee_id"`
	DayOfWeek   int       `json:"day_of_week"`
	StartTime   TimeOfDay `json:"start_time"`
	EndTime     TimeOfDay `json:"end_time"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

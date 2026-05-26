package search

import "github.com/google/uuid"

// ServiceSearchParams filters marketplace service search.
type ServiceSearchParams struct {
	Query      string
	CategoryID *uuid.UUID
	Location   string
	MinPrice   *float64
	MaxPrice   *float64
	Sort       string
	Limit      int
}

// EmployeeSearchParams filters marketplace employee search.
type EmployeeSearchParams struct {
	Query      string
	CategoryID *uuid.UUID
	Latitude   *float64
	Longitude  *float64
	RadiusKm   *float64
	Sort       string
	Limit      int
}

// ServiceSearchItem is a service search result row.
type ServiceSearchItem struct {
	ID                  uuid.UUID `json:"id"`
	EmployeeID          uuid.UUID `json:"employee_id"`
	CategoryID          uuid.UUID `json:"category_id"`
	Title               string    `json:"title"`
	Description         *string   `json:"description,omitempty"`
	Price               float64   `json:"price"`
	DurationMinutes     int       `json:"duration_minutes"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           string    `json:"created_at"`
	UpdatedAt           string    `json:"updated_at"`
	EmployeeDisplayName *string   `json:"employee_display_name,omitempty"`
	EmployeeLocation    *string   `json:"employee_location,omitempty"`
	Rating              *float64  `json:"rating"`
}

// EmployeeSearchItem is an employee search result row.
type EmployeeSearchItem struct {
	ID                  uuid.UUID `json:"id"`
	DisplayName         *string   `json:"display_name,omitempty"`
	Bio                 *string   `json:"bio,omitempty"`
	ExperienceYears     int       `json:"experience_years"`
	ProfilePhotoURL     *string   `json:"profile_photo_url,omitempty"`
	LocationText        *string   `json:"location_text,omitempty"`
	Latitude            *float64  `json:"latitude,omitempty"`
	Longitude           *float64  `json:"longitude,omitempty"`
	ServiceAreaRadiusKm *float64  `json:"service_area_radius_km,omitempty"`
	Languages           []string  `json:"languages"`
	Skills              []string  `json:"skills"`
	DistanceKm          *float64  `json:"distance_km,omitempty"`
	Rating              *float64  `json:"rating"`
}

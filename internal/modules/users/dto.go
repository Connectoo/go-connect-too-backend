package users

import "github.com/google/uuid"

// ProfileResponse is the authenticated user's profile.
type ProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// UpdateProfileRequest updates editable user fields.
type UpdateProfileRequest struct {
	Name  string  `json:"name"`
	Phone *string `json:"phone,omitempty"`
}

// CreateAddressRequest creates a saved address.
type CreateAddressRequest struct {
	Label       string   `json:"label"`
	AddressLine string   `json:"address_line"`
	City        string   `json:"city"`
	State       string   `json:"state"`
	Country     string   `json:"country"`
	Pincode     string   `json:"pincode"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	IsDefault   bool     `json:"is_default"`
}

// UpdateAddressRequest replaces address fields.
type UpdateAddressRequest struct {
	Label       string   `json:"label"`
	AddressLine string   `json:"address_line"`
	City        string   `json:"city"`
	State       string   `json:"state"`
	Country     string   `json:"country"`
	Pincode     string   `json:"pincode"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	IsDefault   bool     `json:"is_default"`
}

// AddressResponse is a saved address payload.
type AddressResponse struct {
	ID          uuid.UUID `json:"id"`
	Label       string    `json:"label"`
	AddressLine string    `json:"address_line"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Country     string    `json:"country"`
	Pincode     string    `json:"pincode"`
	Latitude    *float64  `json:"latitude,omitempty"`
	Longitude   *float64  `json:"longitude,omitempty"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

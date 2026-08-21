package public

import (
	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/categories"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/services"
)

// HomeResponse is the public homepage payload.
type HomeResponse struct {
	FeaturedCategories []categories.CategoryResponse `json:"featured_categories"`
	FeaturedProviders  []ProviderResponse            `json:"featured_providers"`
	FeaturedServices   []services.ServiceResponse    `json:"featured_services"`
	Stats              HomeStats                     `json:"stats"`
}

// HomeStats holds basic public marketplace counts.
type HomeStats struct {
	CategoriesCount int `json:"categories_count"`
	ProvidersCount  int `json:"providers_count"`
	ServicesCount   int `json:"services_count"`
}

// ProviderResponse is a public provider profile payload.
type ProviderResponse struct {
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
	Rating              *float64  `json:"rating,omitempty"`
	TotalReviews        int       `json:"total_reviews"`
}

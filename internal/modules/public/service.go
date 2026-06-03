package public

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/categories"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/services"
)

// CategoryReader loads active categories.
type CategoryReader interface {
	ListActive(ctx context.Context) ([]categories.Category, error)
	CountActive(ctx context.Context) (int, error)
}

// ProviderReader loads public provider profiles.
type ProviderReader interface {
	ListApprovedProviders(ctx context.Context, limit int) ([]employees.Profile, error)
	CountApprovedProviders(ctx context.Context) (int, error)
}

// ServiceReader loads public services.
type ServiceReader interface {
	ListPublicActive(ctx context.Context, categoryID *uuid.UUID, limit int) ([]services.EmployeeService, error)
	GetPublicActiveByID(ctx context.Context, serviceID uuid.UUID) (*services.EmployeeService, error)
	CountActivePublic(ctx context.Context) (int, error)
}

// EmployeeReader loads approved employee profiles by id.
type EmployeeReader interface {
	GetApprovedByID(ctx context.Context, id uuid.UUID) (*employees.Profile, error)
}

// Service handles public website business logic.
type Service struct {
	categories CategoryReader
	providers  ProviderReader
	services   ServiceReader
	employees  EmployeeReader
}

// NewService creates a public website service.
func NewService(
	categories CategoryReader,
	providers ProviderReader,
	services ServiceReader,
	employees EmployeeReader,
) *Service {
	return &Service{
		categories: categories,
		providers:  providers,
		services:   services,
		employees:  employees,
	}
}

// GetHome returns homepage content for the public website.
func (s *Service) GetHome(ctx context.Context) (*HomeResponse, error) {
	categoryItems, err := s.categories.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	featuredCategories := make([]categories.CategoryResponse, 0, min(len(categoryItems), 6))
	for i := range categoryItems {
		if i >= 6 {
			break
		}
		featuredCategories = append(featuredCategories, *toCategoryResponse(&categoryItems[i]))
	}

	providerProfiles, err := s.providers.ListApprovedProviders(ctx, 6)
	if err != nil {
		return nil, err
	}
	featuredProviders := make([]ProviderResponse, 0, len(providerProfiles))
	for i := range providerProfiles {
		featuredProviders = append(featuredProviders, toProviderResponse(&providerProfiles[i]))
	}

	serviceItems, err := s.services.ListPublicActive(ctx, nil, 6)
	if err != nil {
		return nil, err
	}
	featuredServices := make([]services.ServiceResponse, 0, len(serviceItems))
	for i := range serviceItems {
		featuredServices = append(featuredServices, *toServiceResponse(&serviceItems[i]))
	}

	categoriesCount, err := s.categories.CountActive(ctx)
	if err != nil {
		return nil, err
	}
	providersCount, err := s.providers.CountApprovedProviders(ctx)
	if err != nil {
		return nil, err
	}
	servicesCount, err := s.services.CountActivePublic(ctx)
	if err != nil {
		return nil, err
	}

	return &HomeResponse{
		FeaturedCategories: featuredCategories,
		FeaturedProviders:  featuredProviders,
		FeaturedServices:   featuredServices,
		Stats: HomeStats{
			CategoriesCount: categoriesCount,
			ProvidersCount:  providersCount,
			ServicesCount:   servicesCount,
		},
	}, nil
}

// ListCategories returns active categories for the public website.
func (s *Service) ListCategories(ctx context.Context) ([]categories.CategoryResponse, error) {
	items, err := s.categories.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]categories.CategoryResponse, 0, len(items))
	for i := range items {
		out = append(out, *toCategoryResponse(&items[i]))
	}
	return out, nil
}

// ListProviders returns approved providers for the public website.
func (s *Service) ListProviders(ctx context.Context, limit int) ([]ProviderResponse, error) {
	items, err := s.providers.ListApprovedProviders(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]ProviderResponse, 0, len(items))
	for i := range items {
		out = append(out, toProviderResponse(&items[i]))
	}
	return out, nil
}

// GetProvider returns an approved provider profile by id.
func (s *Service) GetProvider(ctx context.Context, providerID uuid.UUID) (*ProviderResponse, error) {
	profile, err := s.employees.GetApprovedByID(ctx, providerID)
	if err != nil {
		return nil, err
	}
	res := toProviderResponse(profile)
	return &res, nil
}

// ListServices returns active public services.
func (s *Service) ListServices(ctx context.Context, categoryID *uuid.UUID, limit int) ([]services.ServiceResponse, error) {
	items, err := s.services.ListPublicActive(ctx, categoryID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]services.ServiceResponse, 0, len(items))
	for i := range items {
		out = append(out, *toServiceResponse(&items[i]))
	}
	return out, nil
}

// GetService returns an active public service by id.
func (s *Service) GetService(ctx context.Context, serviceID uuid.UUID) (*services.ServiceResponse, error) {
	item, err := s.services.GetPublicActiveByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	return toServiceResponse(item), nil
}

func toCategoryResponse(category *categories.Category) *categories.CategoryResponse {
	return &categories.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
		CreatedAt:   category.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   category.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toProviderResponse(profile *employees.Profile) ProviderResponse {
	return ProviderResponse{
		ID:                  profile.ID,
		DisplayName:         profile.DisplayName,
		Bio:                 profile.Bio,
		ExperienceYears:     profile.ExperienceYears,
		ProfilePhotoURL:     profile.ProfilePhotoURL,
		LocationText:        profile.LocationText,
		Latitude:            profile.Latitude,
		Longitude:           profile.Longitude,
		ServiceAreaRadiusKm: profile.ServiceAreaRadiusKm,
		Languages:           profile.Languages,
		Skills:              profile.Skills,
		Rating:              profile.AverageRating,
		TotalReviews:        profile.TotalReviews,
	}
}

func toServiceResponse(service *services.EmployeeService) *services.ServiceResponse {
	return &services.ServiceResponse{
		ID:              service.ID,
		EmployeeID:      service.EmployeeID,
		CategoryID:      service.CategoryID,
		Title:           service.Title,
		Description:     service.Description,
		Price:           service.Price,
		DurationMinutes: service.DurationMinutes,
		IsActive:        service.IsActive,
		CreatedAt:       service.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       service.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

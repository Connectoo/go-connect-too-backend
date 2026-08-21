package search

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Store runs marketplace search queries.
type Store interface {
	SearchServices(ctx context.Context, params ServiceSearchParams) ([]serviceSearchRow, error)
	SearchEmployees(ctx context.Context, params EmployeeSearchParams) ([]employeeSearchRow, error)
}

// Service handles search business logic.
type Service struct {
	store Store
}

// NewService creates a search service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// SearchServices finds marketplace services.
func (s *Service) SearchServices(ctx context.Context, params ServiceSearchParams) ([]ServiceSearchItem, error) {
	if err := validateServiceSearch(params); err != nil {
		return nil, err
	}

	rows, err := s.store.SearchServices(ctx, params)
	if err != nil {
		return nil, err
	}

	out := make([]ServiceSearchItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, toServiceSearchItem(row))
	}
	return out, nil
}

// SearchEmployees finds marketplace employees.
func (s *Service) SearchEmployees(ctx context.Context, params EmployeeSearchParams) ([]EmployeeSearchItem, error) {
	if err := validateEmployeeSearch(params); err != nil {
		return nil, err
	}

	rows, err := s.store.SearchEmployees(ctx, params)
	if err != nil {
		return nil, err
	}

	out := make([]EmployeeSearchItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, toEmployeeSearchItem(row))
	}
	return out, nil
}

// ParseServiceSearchParams reads query string values for service search.
func ParseServiceSearchParams(values map[string][]string) (ServiceSearchParams, error) {
	params := ServiceSearchParams{
		Query:    firstValue(values, "q"),
		Location: firstValue(values, "location"),
		Sort:     firstValue(values, "sort"),
	}

	if raw := firstValue(values, "category_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return ServiceSearchParams{}, fmt.Errorf("%w: invalid category_id", ErrValidation)
		}
		params.CategoryID = &id
	}
	if raw := firstValue(values, "min_price"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || v < 0 {
			return ServiceSearchParams{}, fmt.Errorf("%w: invalid min_price", ErrValidation)
		}
		params.MinPrice = &v
	}
	if raw := firstValue(values, "max_price"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || v < 0 {
			return ServiceSearchParams{}, fmt.Errorf("%w: invalid max_price", ErrValidation)
		}
		params.MaxPrice = &v
	}
	if params.MinPrice != nil && params.MaxPrice != nil && *params.MinPrice > *params.MaxPrice {
		return ServiceSearchParams{}, fmt.Errorf("%w: min_price cannot exceed max_price", ErrValidation)
	}

	return params, nil
}

// ParseEmployeeSearchParams reads query string values for employee search.
func ParseEmployeeSearchParams(values map[string][]string) (EmployeeSearchParams, error) {
	params := EmployeeSearchParams{
		Query: firstValue(values, "q"),
		Sort:  firstValue(values, "sort"),
	}

	if raw := firstValue(values, "category_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return EmployeeSearchParams{}, fmt.Errorf("%w: invalid category_id", ErrValidation)
		}
		params.CategoryID = &id
	}
	if raw := firstValue(values, "latitude"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || v < -90 || v > 90 {
			return EmployeeSearchParams{}, fmt.Errorf("%w: invalid latitude", ErrValidation)
		}
		params.Latitude = &v
	}
	if raw := firstValue(values, "longitude"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || v < -180 || v > 180 {
			return EmployeeSearchParams{}, fmt.Errorf("%w: invalid longitude", ErrValidation)
		}
		params.Longitude = &v
	}
	if raw := firstValue(values, "radius_km"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || v <= 0 {
			return EmployeeSearchParams{}, fmt.Errorf("%w: invalid radius_km", ErrValidation)
		}
		params.RadiusKm = &v
	}
	if (params.Latitude == nil) != (params.Longitude == nil) {
		return EmployeeSearchParams{}, fmt.Errorf("%w: latitude and longitude must be provided together", ErrValidation)
	}
	if params.RadiusKm != nil && (params.Latitude == nil || params.Longitude == nil) {
		return EmployeeSearchParams{}, fmt.Errorf("%w: radius_km requires latitude and longitude", ErrValidation)
	}

	return params, nil
}

func validateServiceSearch(params ServiceSearchParams) error {
	if params.MinPrice != nil && params.MaxPrice != nil && *params.MinPrice > *params.MaxPrice {
		return fmt.Errorf("%w: min_price cannot exceed max_price", ErrValidation)
	}
	return nil
}

func validateEmployeeSearch(params EmployeeSearchParams) error {
	if (params.Latitude == nil) != (params.Longitude == nil) {
		return fmt.Errorf("%w: latitude and longitude must be provided together", ErrValidation)
	}
	if params.RadiusKm != nil && (params.Latitude == nil || params.Longitude == nil) {
		return fmt.Errorf("%w: radius_km requires latitude and longitude", ErrValidation)
	}
	return nil
}

func firstValue(values map[string][]string, key string) string {
	items, ok := values[key]
	if !ok || len(items) == 0 {
		return ""
	}
	return strings.TrimSpace(items[0])
}

func toServiceSearchItem(row serviceSearchRow) ServiceSearchItem {
	return ServiceSearchItem{
		ID:                  row.ID,
		EmployeeID:          row.EmployeeID,
		CategoryID:          row.CategoryID,
		Title:               row.Title,
		Description:         row.Description,
		Price:               row.Price,
		DurationMinutes:     row.DurationMinutes,
		IsActive:            row.IsActive,
		CreatedAt:           row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           row.UpdatedAt.UTC().Format(time.RFC3339),
		EmployeeDisplayName: row.EmployeeDisplayName,
		EmployeeLocation:    row.EmployeeLocation,
		Rating:              nil,
	}
}

func toEmployeeSearchItem(row employeeSearchRow) EmployeeSearchItem {
	return EmployeeSearchItem{
		ID:                  row.ID,
		DisplayName:         row.DisplayName,
		Bio:                 row.Bio,
		ExperienceYears:     row.ExperienceYears,
		ProfilePhotoURL:     row.ProfilePhotoURL,
		LocationText:        row.LocationText,
		Latitude:            row.Latitude,
		Longitude:           row.Longitude,
		ServiceAreaRadiusKm: row.ServiceAreaRadiusKm,
		Languages:           row.Languages,
		Skills:              row.Skills,
		DistanceKm:          row.DistanceKm,
		Rating:              nil,
	}
}

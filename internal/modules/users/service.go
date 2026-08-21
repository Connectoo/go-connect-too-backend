package users

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UserStore loads and updates users.
type UserStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Update(ctx context.Context, user *User, at time.Time) (*User, error)
}

// AddressStore manages saved addresses.
type AddressStore interface {
	ListAddressesByUserID(ctx context.Context, userID uuid.UUID) ([]Address, error)
	CreateAddress(ctx context.Context, addr *Address) (*Address, error)
	GetAddressByID(ctx context.Context, userID, addressID uuid.UUID) (*Address, error)
	UpdateAddress(ctx context.Context, addr *Address, at time.Time) (*Address, error)
	DeleteAddress(ctx context.Context, userID, addressID uuid.UUID) error
	ClearDefaultAddresses(ctx context.Context, userID uuid.UUID, exceptID *uuid.UUID, at time.Time) error
}

// Service handles user profile and address business logic.
type Service struct {
	users     UserStore
	addresses AddressStore
	now       func() time.Time
}

// NewService creates a users service.
func NewService(users UserStore, addresses AddressStore) *Service {
	return &Service{
		users:     users,
		addresses: addresses,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// GetProfile returns the authenticated user's profile.
func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*ProfileResponse, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toProfileResponse(user), nil
}

// UpdateProfile updates the authenticated user's profile.
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, role string, req UpdateProfileRequest) (*ProfileResponse, error) {
	if role != RoleCustomer {
		return nil, fmt.Errorf("%w: only customers can update profile here", ErrForbiddenProfile)
	}
	if err := validateProfileUpdate(req); err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Name = strings.TrimSpace(req.Name)
	user.Phone = normalizePhone(req.Phone)

	updated, err := s.users.Update(ctx, user, s.now())
	if err != nil {
		return nil, err
	}
	return toProfileResponse(updated), nil
}

// ListAddresses returns saved addresses for a customer.
func (s *Service) ListAddresses(ctx context.Context, userID uuid.UUID, role string) ([]AddressResponse, error) {
	if role != RoleCustomer {
		return nil, fmt.Errorf("%w: only customers can manage addresses", ErrForbiddenProfile)
	}

	items, err := s.addresses.ListAddressesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]AddressResponse, 0, len(items))
	for i := range items {
		out = append(out, *toAddressResponse(&items[i]))
	}
	return out, nil
}

// CreateAddress adds a saved address for a customer.
func (s *Service) CreateAddress(ctx context.Context, userID uuid.UUID, role string, req CreateAddressRequest) (*AddressResponse, error) {
	if role != RoleCustomer {
		return nil, fmt.Errorf("%w: only customers can manage addresses", ErrForbiddenProfile)
	}
	fields, err := validateAddressFields(req.Label, req.AddressLine, req.City, req.State, req.Country, req.Pincode, req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}

	at := s.now()
	if req.IsDefault {
		if err := s.addresses.ClearDefaultAddresses(ctx, userID, nil, at); err != nil {
			return nil, err
		}
	}

	created, err := s.addresses.CreateAddress(ctx, &Address{
		ID:          uuid.New(),
		UserID:      userID,
		Label:       fields.label,
		AddressLine: fields.addressLine,
		City:        fields.city,
		State:       fields.state,
		Country:     fields.country,
		Pincode:     fields.pincode,
		Latitude:    fields.latitude,
		Longitude:   fields.longitude,
		IsDefault:   req.IsDefault,
		CreatedAt:   at,
		UpdatedAt:   at,
	})
	if err != nil {
		return nil, err
	}
	return toAddressResponse(created), nil
}

// UpdateAddress replaces a saved address for a customer.
func (s *Service) UpdateAddress(ctx context.Context, userID uuid.UUID, role string, addressID uuid.UUID, req UpdateAddressRequest) (*AddressResponse, error) {
	if role != RoleCustomer {
		return nil, fmt.Errorf("%w: only customers can manage addresses", ErrForbiddenProfile)
	}
	fields, err := validateAddressFields(req.Label, req.AddressLine, req.City, req.State, req.Country, req.Pincode, req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}

	if _, err := s.addresses.GetAddressByID(ctx, userID, addressID); err != nil {
		return nil, err
	}

	at := s.now()
	if req.IsDefault {
		if err := s.addresses.ClearDefaultAddresses(ctx, userID, &addressID, at); err != nil {
			return nil, err
		}
	}

	updated, err := s.addresses.UpdateAddress(ctx, &Address{
		ID:          addressID,
		UserID:      userID,
		Label:       fields.label,
		AddressLine: fields.addressLine,
		City:        fields.city,
		State:       fields.state,
		Country:     fields.country,
		Pincode:     fields.pincode,
		Latitude:    fields.latitude,
		Longitude:   fields.longitude,
		IsDefault:   req.IsDefault,
	}, at)
	if err != nil {
		return nil, err
	}
	return toAddressResponse(updated), nil
}

// DeleteAddress removes a saved address for a customer.
func (s *Service) DeleteAddress(ctx context.Context, userID uuid.UUID, role string, addressID uuid.UUID) error {
	if role != RoleCustomer {
		return fmt.Errorf("%w: only customers can manage addresses", ErrForbiddenProfile)
	}
	return s.addresses.DeleteAddress(ctx, userID, addressID)
}

func validateProfileUpdate(req UpdateProfileRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	return nil
}

type addressFields struct {
	label, addressLine, city, state, country, pincode string
	latitude, longitude                               *float64
}

func validateAddressFields(label, addressLine, city, state, country, pincode string, latitude, longitude *float64) (addressFields, error) {
	fields := addressFields{
		label:       strings.TrimSpace(label),
		addressLine: strings.TrimSpace(addressLine),
		city:        strings.TrimSpace(city),
		state:       strings.TrimSpace(state),
		country:     strings.TrimSpace(country),
		pincode:     strings.TrimSpace(pincode),
		latitude:    latitude,
		longitude:   longitude,
	}
	if fields.label == "" {
		return addressFields{}, fmt.Errorf("%w: label is required", ErrValidation)
	}
	if fields.addressLine == "" {
		return addressFields{}, fmt.Errorf("%w: address_line is required", ErrValidation)
	}
	if fields.city == "" {
		return addressFields{}, fmt.Errorf("%w: city is required", ErrValidation)
	}
	if fields.state == "" {
		return addressFields{}, fmt.Errorf("%w: state is required", ErrValidation)
	}
	if fields.country == "" {
		return addressFields{}, fmt.Errorf("%w: country is required", ErrValidation)
	}
	if fields.pincode == "" {
		return addressFields{}, fmt.Errorf("%w: pincode is required", ErrValidation)
	}
	if fields.latitude != nil && (*fields.latitude < -90 || *fields.latitude > 90) {
		return addressFields{}, fmt.Errorf("%w: latitude must be between -90 and 90", ErrValidation)
	}
	if fields.longitude != nil && (*fields.longitude < -180 || *fields.longitude > 180) {
		return addressFields{}, fmt.Errorf("%w: longitude must be between -180 and 180", ErrValidation)
	}
	return fields, nil
}

func normalizePhone(phone *string) *string {
	if phone == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*phone)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// UserDeactivator deactivates user accounts.
type UserDeactivator interface {
	Deactivate(ctx context.Context, id uuid.UUID, at time.Time) (*User, error)
}

// Deactivate deactivates the authenticated user's account.
func (s *Service) Deactivate(ctx context.Context, userID uuid.UUID) (*ProfileResponse, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.DeactivatedAt != nil {
		return nil, ErrDeactivated
	}

	deactivator, ok := s.users.(UserDeactivator)
	if !ok {
		return nil, fmt.Errorf("%w: deactivate not configured", ErrValidation)
	}

	updated, err := deactivator.Deactivate(ctx, userID, s.now())
	if err != nil {
		return nil, err
	}
	return toProfileResponse(updated), nil
}

func toProfileResponse(user *User) *ProfileResponse {
	return &ProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toAddressResponse(addr *Address) *AddressResponse {
	return &AddressResponse{
		ID:          addr.ID,
		Label:       addr.Label,
		AddressLine: addr.AddressLine,
		City:        addr.City,
		State:       addr.State,
		Country:     addr.Country,
		Pincode:     addr.Pincode,
		Latitude:    addr.Latitude,
		Longitude:   addr.Longitude,
		IsDefault:   addr.IsDefault,
		CreatedAt:   addr.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   addr.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

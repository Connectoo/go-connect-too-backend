package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockUserStore struct {
	users map[uuid.UUID]*User
}

func (m *mockUserStore) GetByID(_ context.Context, id uuid.UUID) (*User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *user
	return &copy, nil
}

func (m *mockUserStore) Update(_ context.Context, user *User, at time.Time) (*User, error) {
	if _, ok := m.users[user.ID]; !ok {
		return nil, ErrNotFound
	}
	user.UpdatedAt = at
	m.users[user.ID] = user
	copy := *user
	return &copy, nil
}

type mockAddressStore struct {
	byUser map[uuid.UUID][]Address
}

func newMockAddressStore() *mockAddressStore {
	return &mockAddressStore{byUser: make(map[uuid.UUID][]Address)}
}

func (m *mockAddressStore) ListAddressesByUserID(_ context.Context, userID uuid.UUID) ([]Address, error) {
	items := m.byUser[userID]
	if items == nil {
		return []Address{}, nil
	}
	return append([]Address(nil), items...), nil
}

func (m *mockAddressStore) CreateAddress(_ context.Context, addr *Address) (*Address, error) {
	copy := *addr
	m.byUser[addr.UserID] = append(m.byUser[addr.UserID], copy)
	return &copy, nil
}

func (m *mockAddressStore) GetAddressByID(_ context.Context, userID, addressID uuid.UUID) (*Address, error) {
	for _, addr := range m.byUser[userID] {
		if addr.ID == addressID {
			copy := addr
			return &copy, nil
		}
	}
	return nil, ErrAddressNotFound
}

func (m *mockAddressStore) UpdateAddress(_ context.Context, addr *Address, at time.Time) (*Address, error) {
	items := m.byUser[addr.UserID]
	for i := range items {
		if items[i].ID == addr.ID {
			addr.UpdatedAt = at
			items[i] = *addr
			m.byUser[addr.UserID] = items
			copy := *addr
			return &copy, nil
		}
	}
	return nil, ErrAddressNotFound
}

func (m *mockAddressStore) DeleteAddress(_ context.Context, userID, addressID uuid.UUID) error {
	items := m.byUser[userID]
	for i, addr := range items {
		if addr.ID == addressID {
			m.byUser[userID] = append(items[:i], items[i+1:]...)
			return nil
		}
	}
	return ErrAddressNotFound
}

func (m *mockAddressStore) ClearDefaultAddresses(_ context.Context, userID uuid.UUID, exceptID *uuid.UUID, at time.Time) error {
	for i, addr := range m.byUser[userID] {
		if exceptID != nil && addr.ID == *exceptID {
			continue
		}
		addr.IsDefault = false
		addr.UpdatedAt = at
		m.byUser[userID][i] = addr
	}
	return nil
}

func TestService_UpdateProfile_customerOnly(t *testing.T) {
	userID := uuid.New()
	store := &mockUserStore{users: map[uuid.UUID]*User{
		userID: {ID: userID, Name: "A", Email: "a@example.com", Role: RoleCustomer, Status: StatusActive},
	}}
	svc := NewService(store, newMockAddressStore())

	_, err := svc.UpdateProfile(context.Background(), userID, RoleEmployee, UpdateProfileRequest{Name: "B"})
	if !errors.Is(err, ErrForbiddenProfile) {
		t.Fatalf("UpdateProfile() error = %v, want ErrForbiddenProfile", err)
	}
}

func TestService_CreateAddress_success(t *testing.T) {
	userID := uuid.New()
	svc := NewService(&mockUserStore{}, newMockAddressStore())

	res, err := svc.CreateAddress(context.Background(), userID, RoleCustomer, CreateAddressRequest{
		Label:       "Home",
		AddressLine: "123 Main St",
		City:        "Delhi",
		State:       "DL",
		Country:     "IN",
		Pincode:     "110001",
		IsDefault:   true,
	})
	if err != nil {
		t.Fatalf("CreateAddress() error = %v", err)
	}
	if res.Label != "Home" || !res.IsDefault {
		t.Fatalf("CreateAddress() = %+v", res)
	}
}

func TestService_CreateAddress_employeeForbidden(t *testing.T) {
	svc := NewService(&mockUserStore{}, newMockAddressStore())
	_, err := svc.CreateAddress(context.Background(), uuid.New(), RoleEmployee, CreateAddressRequest{
		Label: "Home", AddressLine: "1", City: "X", State: "Y", Country: "Z", Pincode: "1",
	})
	if !errors.Is(err, ErrForbiddenProfile) {
		t.Fatalf("CreateAddress() error = %v, want ErrForbiddenProfile", err)
	}
}

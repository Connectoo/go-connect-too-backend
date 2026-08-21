package search

import (
	"context"
	"errors"
	"testing"
)

type mockSearchStore struct {
	services  []serviceSearchRow
	employees []employeeSearchRow
}

func (m *mockSearchStore) SearchServices(_ context.Context, params ServiceSearchParams) ([]serviceSearchRow, error) {
	if params.MinPrice != nil && params.MaxPrice != nil && *params.MinPrice > *params.MaxPrice {
		return nil, errors.New("bad range")
	}
	return m.services, nil
}

func (m *mockSearchStore) SearchEmployees(_ context.Context, _ EmployeeSearchParams) ([]employeeSearchRow, error) {
	return m.employees, nil
}

func TestParseEmployeeSearchParams_requiresPairedCoordinates(t *testing.T) {
	_, err := ParseEmployeeSearchParams(map[string][]string{"latitude": {"12.9"}})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("ParseEmployeeSearchParams() error = %v, want ErrValidation", err)
	}
}

func TestParseServiceSearchParams_invalidCategory(t *testing.T) {
	_, err := ParseServiceSearchParams(map[string][]string{"category_id": {"not-a-uuid"}})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("ParseServiceSearchParams() error = %v, want ErrValidation", err)
	}
}

func TestService_SearchServices(t *testing.T) {
	svc := NewService(&mockSearchStore{services: []serviceSearchRow{}})
	items, err := svc.SearchServices(context.Background(), ServiceSearchParams{Query: "clean"})
	if err != nil {
		t.Fatalf("SearchServices() error = %v", err)
	}
	if items == nil {
		t.Fatal("SearchServices() returned nil slice")
	}
}

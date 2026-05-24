package categories

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockCategoryStore struct {
	active []Category
	byName map[string]*Category
}

func newMockCategoryStore() *mockCategoryStore {
	return &mockCategoryStore{
		active: []Category{},
		byName: make(map[string]*Category),
	}
}

func (m *mockCategoryStore) ListActive(_ context.Context) ([]Category, error) {
	out := make([]Category, 0, len(m.active))
	for _, category := range m.active {
		if category.IsActive {
			out = append(out, category)
		}
	}
	return out, nil
}

func (m *mockCategoryStore) Create(_ context.Context, category *Category) (*Category, error) {
	key := strings.ToLower(category.Name)
	if _, exists := m.byName[key]; exists {
		return nil, ErrDuplicateName
	}

	copy := *category
	m.byName[key] = &copy
	if copy.IsActive {
		m.active = append(m.active, copy)
	}
	return &copy, nil
}

func TestService_ListActive(t *testing.T) {
	store := newMockCategoryStore()
	store.active = []Category{
		{ID: uuid.New(), Name: "Plumbing", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Inactive", IsActive: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	svc := NewService(store)
	items, err := svc.ListActive(context.Background())
	if err != nil {
		t.Fatalf("ListActive() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListActive() len = %d, want 1 active category", len(items))
	}
	if items[0].Name != "Plumbing" {
		t.Fatalf("Name = %q, want Plumbing", items[0].Name)
	}
}

func TestService_CreateCategory_validation(t *testing.T) {
	svc := NewService(newMockCategoryStore())

	_, err := svc.CreateCategory(context.Background(), CreateCategoryRequest{Name: "   "})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("CreateCategory() error = %v, want ErrValidation", err)
	}
}

func TestService_CreateCategory_success(t *testing.T) {
	store := newMockCategoryStore()
	svc := NewService(store)
	svc.now = func() time.Time { return time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC) }

	desc := "Home plumbing services"
	active := true
	res, err := svc.CreateCategory(context.Background(), CreateCategoryRequest{
		Name:        "Plumbing",
		Description: &desc,
		IsActive:    &active,
	})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}
	if res.Name != "Plumbing" {
		t.Fatalf("Name = %q, want Plumbing", res.Name)
	}
	if res.Description == nil || *res.Description != desc {
		t.Fatalf("Description = %v, want %q", res.Description, desc)
	}
	if !res.IsActive {
		t.Fatal("IsActive = false, want true")
	}
}

func TestService_CreateCategory_duplicateName(t *testing.T) {
	store := newMockCategoryStore()
	svc := NewService(store)

	_, err := svc.CreateCategory(context.Background(), CreateCategoryRequest{Name: "Cleaning"})
	if err != nil {
		t.Fatalf("first CreateCategory() error = %v", err)
	}

	_, err = svc.CreateCategory(context.Background(), CreateCategoryRequest{Name: "cleaning"})
	if !errors.Is(err, ErrDuplicateName) {
		t.Fatalf("second CreateCategory() error = %v, want ErrDuplicateName", err)
	}
}

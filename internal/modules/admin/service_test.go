package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

type mockStatsStore struct {
	summary *DashboardSummary
	err     error
}

func (m mockStatsStore) DashboardSummary(_ context.Context) (*DashboardSummary, error) {
	return m.summary, m.err
}

type mockUserAdminStore struct {
	user  *users.User
	users []users.User
	total int
	err   error
}

func (m mockUserAdminStore) GetByID(_ context.Context, _ uuid.UUID) (*users.User, error) {
	return m.user, m.err
}

func (m mockUserAdminStore) List(_ context.Context, _ users.ListFilter) ([]users.User, int, error) {
	return m.users, m.total, m.err
}

func (m mockUserAdminStore) UpdateStatus(_ context.Context, _ uuid.UUID, status string, _ time.Time) (*users.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	updated := *m.user
	updated.Status = status
	return &updated, nil
}

func TestService_SuspendUser_rejectsAdmin(t *testing.T) {
	adminUser := &users.User{
		ID:     uuid.New(),
		Role:   users.RoleAdmin,
		Status: users.StatusActive,
	}
	svc := NewService(mockStatsStore{}, mockUserAdminStore{user: adminUser})

	_, err := svc.SuspendUser(context.Background(), adminUser.ID)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("SuspendUser() error = %v, want ErrValidation", err)
	}
}

func TestService_ListUsers_returnsPaginatedResult(t *testing.T) {
	userID := uuid.New()
	svc := NewService(mockStatsStore{}, mockUserAdminStore{
		users: []users.User{{
			ID:        userID,
			Name:      "Jane",
			Email:     "jane@example.com",
			Role:      users.RoleCustomer,
			Status:    users.StatusActive,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}},
		total: 1,
	})

	res, err := svc.ListUsers(context.Background(), UserListFilter{Page: pagination.Params{Page: 1, Limit: 20}})
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(res.Items) != 1 || res.Total != 1 {
		t.Fatalf("ListUsers() = %+v, want one item", res)
	}
}

package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

// StatsStore loads dashboard metrics.
type StatsStore interface {
	DashboardSummary(ctx context.Context) (*DashboardSummary, error)
}

// UserAdminStore loads and updates users for admin operations.
type UserAdminStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*users.User, error)
	List(ctx context.Context, filter users.ListFilter) ([]users.User, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, at time.Time) (*users.User, error)
}

// Service handles admin dashboard and user management.
type Service struct {
	stats StatsStore
	users UserAdminStore
	now   func() time.Time
}

// NewService creates an admin service.
func NewService(stats StatsStore, userStore UserAdminStore) *Service {
	return &Service{
		stats: stats,
		users: userStore,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

// DashboardSummary returns basic platform metrics.
func (s *Service) DashboardSummary(ctx context.Context) (*DashboardSummaryResponse, error) {
	summary, err := s.stats.DashboardSummary(ctx)
	if err != nil {
		return nil, err
	}
	return &DashboardSummaryResponse{
		TotalUsers:          summary.TotalUsers,
		TotalCustomers:      summary.TotalCustomers,
		TotalEmployees:      summary.TotalEmployees,
		PendingEmployees:    summary.PendingEmployees,
		TotalBookings:       summary.TotalBookings,
		ActiveBookings:      summary.ActiveBookings,
		TotalServices:       summary.TotalServices,
		ActiveServices:      summary.ActiveServices,
		TotalPayments:       summary.TotalPayments,
		CompletedPayments:   summary.CompletedPayments,
		TotalRevenue:        summary.TotalRevenue,
		ActiveSubscriptions: summary.ActiveSubscriptions,
	}, nil
}

// ListUsers returns paginated users for admin views.
func (s *Service) ListUsers(ctx context.Context, filter UserListFilter) (UserListResponse, error) {
	items, total, err := s.users.List(ctx, users.ListFilter{
		Role:   filter.Role,
		Status: filter.Status,
		Query:  filter.Query,
		Offset: filter.Page.Offset(),
		Limit:  filter.Page.Limit,
	})
	if err != nil {
		return pagination.Result[UserResponse]{}, err
	}

	out := make([]UserResponse, 0, len(items))
	for i := range items {
		out = append(out, toUserResponse(&items[i]))
	}
	return pagination.NewResult(out, filter.Page, total), nil
}

// GetUser returns a user by id for admin views.
func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := toUserResponse(user)
	return &res, nil
}

// SuspendUser marks a user account as suspended.
func (s *Service) SuspendUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	return s.setUserStatus(ctx, userID, users.StatusSuspended)
}

// ActivateUser marks a user account as active.
func (s *Service) ActivateUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	return s.setUserStatus(ctx, userID, users.StatusActive)
}

func (s *Service) setUserStatus(ctx context.Context, userID uuid.UUID, status string) (*UserResponse, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Status == status {
		res := toUserResponse(user)
		return &res, nil
	}
	if status == users.StatusSuspended && user.Role == users.RoleAdmin {
		return nil, fmt.Errorf("%w: admin accounts cannot be suspended", ErrValidation)
	}

	updated, err := s.users.UpdateStatus(ctx, userID, status, s.now())
	if err != nil {
		return nil, err
	}
	res := toUserResponse(updated)
	return &res, nil
}

func toUserResponse(user *users.User) UserResponse {
	return UserResponse{
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

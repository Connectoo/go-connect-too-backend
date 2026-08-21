package admin

import (
	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
)

// UserListFilter filters admin user listings.
type UserListFilter struct {
	Role   string
	Status string
	Query  string
	Page   pagination.Params
}

// UserListResponse is a paginated admin user list.
type UserListResponse = pagination.Result[UserResponse]

// UserResponse is the admin user payload.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// DashboardSummaryResponse is the admin dashboard payload.
type DashboardSummaryResponse struct {
	TotalUsers          int     `json:"total_users"`
	TotalCustomers      int     `json:"total_customers"`
	TotalEmployees      int     `json:"total_employees"`
	PendingEmployees    int     `json:"pending_employees"`
	TotalBookings       int     `json:"total_bookings"`
	ActiveBookings      int     `json:"active_bookings"`
	TotalServices       int     `json:"total_services"`
	ActiveServices      int     `json:"active_services"`
	TotalPayments       int     `json:"total_payments"`
	CompletedPayments   int     `json:"completed_payments"`
	TotalRevenue        float64 `json:"total_revenue"`
	ActiveSubscriptions int     `json:"active_subscriptions"`
}

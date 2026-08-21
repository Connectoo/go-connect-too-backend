package admin

// DashboardSummary holds basic platform metrics.
type DashboardSummary struct {
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

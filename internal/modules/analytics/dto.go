package analytics

// EmployeeSummaryResponse is the employee analytics overview.
type EmployeeSummaryResponse struct {
	Period                PeriodResponse `json:"period"`
	ProfileViews          int            `json:"profile_views"`
	TotalBookings         int            `json:"total_bookings"`
	CompletedBookings     int            `json:"completed_bookings"`
	CancelledBookings     int            `json:"cancelled_bookings"`
	AverageResponseTimeMs *int64         `json:"average_response_time_ms"`
	EstimatedRevenue      string         `json:"estimated_revenue"`
	RatingTrend           []RatingPoint  `json:"rating_trend"`
}

// RatingPoint is one bucket in a rating trend series.
type RatingPoint struct {
	Period        string  `json:"period"`
	AverageRating float64 `json:"average_rating"`
	ReviewCount   int     `json:"review_count"`
}

// EmployeeBookingsResponse is booking analytics for an employee.
type EmployeeBookingsResponse struct {
	Period      PeriodResponse    `json:"period"`
	Total       int               `json:"total"`
	ByStatus    []StatusCountItem `json:"by_status"`
	DailyVolume []DailyCountItem  `json:"daily_volume"`
}

// StatusCountItem is a status aggregate in API responses.
type StatusCountItem struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// DailyCountItem is a daily booking count.
type DailyCountItem struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// EmployeeReviewsResponse is review analytics for an employee.
type EmployeeReviewsResponse struct {
	Period             PeriodResponse `json:"period"`
	AverageRating      *float64       `json:"average_rating"`
	TotalReviews       int            `json:"total_reviews"`
	RatingTrend        []RatingPoint  `json:"rating_trend"`
	RatingDistribution map[int]int    `json:"rating_distribution"`
}

// AdminSummaryResponse is the admin analytics overview.
type AdminSummaryResponse struct {
	Period                  PeriodResponse `json:"period"`
	TotalUsers              int            `json:"total_users"`
	TotalEmployees          int            `json:"total_employees"`
	ApprovedEmployees       int            `json:"approved_employees"`
	ActiveSubscriptions     int            `json:"active_subscriptions"`
	MonthlyRecurringRevenue int64          `json:"monthly_recurring_revenue"`
	BookingVolume           int            `json:"booking_volume"`
	FailedPayments          int            `json:"failed_payments"`
	ChurnRate               *float64       `json:"churn_rate"`
}

// AdminRevenueResponse is subscription and booking revenue analytics.
type AdminRevenueResponse struct {
	Period              PeriodResponse     `json:"period"`
	SubscriptionRevenue int64              `json:"subscription_revenue"`
	BookingRevenue      string             `json:"booking_revenue"`
	DailySubscription   []DailyRevenueItem `json:"daily_subscription_revenue"`
}

// DailyRevenueItem is daily revenue in minor currency units (paise).
type DailyRevenueItem struct {
	Date   string `json:"date"`
	Amount int64  `json:"amount"`
}

// AdminBookingsResponse is platform booking volume analytics.
type AdminBookingsResponse struct {
	Period      PeriodResponse    `json:"period"`
	Total       int               `json:"total"`
	ByStatus    []StatusCountItem `json:"by_status"`
	DailyVolume []DailyCountItem  `json:"daily_volume"`
}

// AdminCategoriesResponse is popular category analytics.
type AdminCategoriesResponse struct {
	Period     PeriodResponse      `json:"period"`
	Categories []CategoryCountItem `json:"categories"`
}

// CategoryCountItem is a category ranked by bookings.
type CategoryCountItem struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	BookingCount int    `json:"booking_count"`
}

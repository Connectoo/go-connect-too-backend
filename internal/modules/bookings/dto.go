package bookings

import (
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/availability"
)

// CreateBookingRequest creates a booking for an active service.
type CreateBookingRequest struct {
	ServiceID     uuid.UUID              `json:"service_id"`
	BookingDate   string                 `json:"booking_date"`
	StartTime     availability.TimeOfDay `json:"start_time"`
	EndTime       availability.TimeOfDay `json:"end_time"`
	CustomerNotes *string                `json:"customer_notes,omitempty"`
}

// CancelBookingRequest optionally carries a cancellation reason.
type CancelBookingRequest struct {
	Reason *string `json:"reason,omitempty"`
}

// EmployeeActionRequest optionally carries employee notes or a rejection reason.
type EmployeeActionRequest struct {
	EmployeeNotes *string `json:"employee_notes,omitempty"`
	Reason        *string `json:"reason,omitempty"`
}

// BookingResponse is the API booking payload.
type BookingResponse struct {
	ID            uuid.UUID `json:"id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	ServiceID     uuid.UUID `json:"service_id"`
	BookingDate   string    `json:"booking_date"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	Status        string    `json:"status"`
	CustomerNotes *string   `json:"customer_notes,omitempty"`
	EmployeeNotes *string   `json:"employee_notes,omitempty"`
	TotalAmount   float64   `json:"total_amount"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
}

func toResponse(b *Booking) *BookingResponse {
	return &BookingResponse{
		ID:            b.ID,
		CustomerID:    b.CustomerID,
		EmployeeID:    b.EmployeeID,
		ServiceID:     b.ServiceID,
		BookingDate:   b.BookingDate.Format("2006-01-02"),
		StartTime:     b.StartTime.String(),
		EndTime:       b.EndTime.String(),
		Status:        b.Status,
		CustomerNotes: b.CustomerNotes,
		EmployeeNotes: b.EmployeeNotes,
		TotalAmount:   b.TotalAmount,
		CreatedAt:     b.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     b.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

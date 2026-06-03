package reviews

import "github.com/google/uuid"

// CreateReviewRequest creates a review for a completed booking.
type CreateReviewRequest struct {
	Rating  int     `json:"rating"`
	Comment *string `json:"comment,omitempty"`
}

// ReplyRequest is an employee reply to a review.
type ReplyRequest struct {
	Reply string `json:"reply"`
}

// ReviewResponse is the API payload for a review.
type ReviewResponse struct {
	ID         uuid.UUID      `json:"id"`
	BookingID  uuid.UUID      `json:"booking_id"`
	CustomerID uuid.UUID      `json:"customer_id"`
	EmployeeID uuid.UUID      `json:"employee_id"`
	Rating     int            `json:"rating"`
	Comment    *string        `json:"comment,omitempty"`
	Status     string         `json:"status"`
	Reply      *ReplyResponse `json:"reply,omitempty"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

// ReplyResponse is the API payload for a review reply.
type ReplyResponse struct {
	ID         uuid.UUID `json:"id"`
	ReviewID   uuid.UUID `json:"review_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	Reply      string    `json:"reply"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

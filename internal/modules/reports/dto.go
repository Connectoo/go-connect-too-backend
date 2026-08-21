package reports

import "github.com/google/uuid"

// CreateReportRequest submits a user report.
type CreateReportRequest struct {
	ReportedUserID uuid.UUID  `json:"reported_user_id"`
	BookingID      *uuid.UUID `json:"booking_id,omitempty"`
	Reason         string     `json:"reason"`
	Description    *string    `json:"description,omitempty"`
}

// ReportResponse is the API payload for a report.
type ReportResponse struct {
	ID             uuid.UUID  `json:"id"`
	ReporterID     uuid.UUID  `json:"reporter_id"`
	ReportedUserID uuid.UUID  `json:"reported_user_id"`
	BookingID      *uuid.UUID `json:"booking_id,omitempty"`
	Reason         string     `json:"reason"`
	Description    *string    `json:"description,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      string     `json:"created_at"`
	UpdatedAt      string     `json:"updated_at"`
}

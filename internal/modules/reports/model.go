package reports

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusOpen     = "open"
	StatusResolved = "resolved"
)

// Report is a user-submitted trust or safety report.
type Report struct {
	ID             uuid.UUID
	ReporterID     uuid.UUID
	ReportedUserID uuid.UUID
	BookingID      *uuid.UUID
	Reason         string
	Description    *string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

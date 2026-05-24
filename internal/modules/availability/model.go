package availability

import (
	"time"

	"github.com/google/uuid"
)

// Days of the week. Sunday is 0 (matches time.Weekday).
const (
	DaySunday    = 0
	DaySaturday  = 6
	MinDayOfWeek = DaySunday
	MaxDayOfWeek = DaySaturday
)

// Availability is a weekly availability slot for an employee.
//
// StartTime and EndTime are clock times of day (no date component).
type Availability struct {
	ID          uuid.UUID
	EmployeeID  uuid.UUID
	DayOfWeek   int
	StartTime   TimeOfDay
	EndTime     TimeOfDay
	IsAvailable bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

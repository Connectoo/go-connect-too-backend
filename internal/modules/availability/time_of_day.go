package availability

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeOfDay represents a clock time of day with minute precision.
//
// It serializes as "HH:MM" in JSON and as TIME in PostgreSQL.
type TimeOfDay struct {
	Hour   int
	Minute int
}

// ParseTimeOfDay parses an "HH:MM" string into a TimeOfDay value.
func ParseTimeOfDay(value string) (TimeOfDay, error) {
	trimmed := strings.TrimSpace(value)
	parts := strings.Split(trimmed, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return TimeOfDay{}, fmt.Errorf("invalid time format")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return TimeOfDay{}, fmt.Errorf("invalid hour")
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return TimeOfDay{}, fmt.Errorf("invalid minute")
	}

	t := TimeOfDay{Hour: hour, Minute: minute}
	if !t.valid() {
		return TimeOfDay{}, fmt.Errorf("time out of range")
	}
	return t, nil
}

func (t TimeOfDay) valid() bool {
	return t.Hour >= 0 && t.Hour < 24 && t.Minute >= 0 && t.Minute < 60
}

// String returns the "HH:MM" representation.
func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

// Minutes returns the time of day as total minutes from midnight.
func (t TimeOfDay) Minutes() int {
	return t.Hour*60 + t.Minute
}

// Before reports whether t is strictly before other.
func (t TimeOfDay) Before(other TimeOfDay) bool {
	return t.Minutes() < other.Minutes()
}

// MarshalJSON implements json.Marshaler.
func (t TimeOfDay) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *TimeOfDay) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	parsed, err := ParseTimeOfDay(raw)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

// Value implements driver.Valuer for SQL TIME columns.
func (t TimeOfDay) Value() (driver.Value, error) {
	return t.String() + ":00", nil
}

// Scan implements sql.Scanner for SQL TIME columns.
func (t *TimeOfDay) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*t = TimeOfDay{}
		return nil
	case string:
		return t.scanString(v)
	case []byte:
		return t.scanString(string(v))
	case time.Time:
		*t = TimeOfDay{Hour: v.Hour(), Minute: v.Minute()}
		return nil
	default:
		return fmt.Errorf("unsupported scan type %T for TimeOfDay", src)
	}
}

func (t *TimeOfDay) scanString(value string) error {
	parsed, err := ParseTimeOfDay(value)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

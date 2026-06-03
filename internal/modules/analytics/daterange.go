package analytics

import (
	"fmt"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// DateRange is a half-open interval [From, To) in UTC.
type DateRange struct {
	From time.Time
	To   time.Time
}

// PeriodResponse exposes the applied filter in API payloads.
type PeriodResponse struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (r DateRange) ToResponse() PeriodResponse {
	endInclusive := r.To.Add(-time.Nanosecond)
	return PeriodResponse{
		From: r.From.UTC().Format(dateLayout),
		To:   endInclusive.UTC().Format(dateLayout),
	}
}

// ParseDateRange parses optional from/to query params (YYYY-MM-DD).
// When both are empty, returns the last defaultDays ending today (UTC).
func ParseDateRange(fromParam, toParam string, defaultDays int) (DateRange, error) {
	fromParam = strings.TrimSpace(fromParam)
	toParam = strings.TrimSpace(toParam)

	if fromParam == "" && toParam == "" {
		if defaultDays <= 0 {
			defaultDays = 30
		}
		now := time.Now().UTC()
		to := startOfDay(now).Add(24 * time.Hour)
		from := to.Add(-time.Duration(defaultDays) * 24 * time.Hour)
		return DateRange{From: from, To: to}, nil
	}

	if fromParam == "" || toParam == "" {
		return DateRange{}, fmt.Errorf("%w: from and to are required together", ErrInvalidDateRange)
	}

	from, err := parseDateParam(fromParam)
	if err != nil {
		return DateRange{}, fmt.Errorf("%w: invalid from date", ErrInvalidDateRange)
	}
	toInclusive, err := parseDateParam(toParam)
	if err != nil {
		return DateRange{}, fmt.Errorf("%w: invalid to date", ErrInvalidDateRange)
	}

	fromStart := startOfDay(from)
	toEnd := startOfDay(toInclusive).Add(24 * time.Hour)
	if !fromStart.Before(toEnd) {
		return DateRange{}, fmt.Errorf("%w: from must be before or equal to to", ErrInvalidDateRange)
	}

	return DateRange{From: fromStart, To: toEnd}, nil
}

func parseDateParam(value string) (time.Time, error) {
	t, err := time.ParseInLocation(dateLayout, value, time.UTC)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func startOfDay(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

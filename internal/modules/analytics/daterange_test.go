package analytics

import (
	"testing"
	"time"
)

func TestParseDateRangeDefaults(t *testing.T) {
	dr, err := ParseDateRange("", "", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dr.From.Before(dr.To) {
		t.Fatalf("expected from before to")
	}
	if dr.To.Sub(dr.From) != 30*24*time.Hour {
		t.Fatalf("expected 30 day range, got %s", dr.To.Sub(dr.From))
	}
}

func TestParseDateRangeExplicit(t *testing.T) {
	dr, err := ParseDateRange("2026-01-01", "2026-01-31", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dr.From.Format("2006-01-02") != "2026-01-01" {
		t.Fatalf("unexpected from: %s", dr.From)
	}
	resp := dr.ToResponse()
	if resp.To != "2026-01-31" {
		t.Fatalf("unexpected to label: %s", resp.To)
	}
}

func TestParseDateRangePartialParams(t *testing.T) {
	_, err := ParseDateRange("2026-01-01", "", 30)
	if err == nil {
		t.Fatal("expected error for partial params")
	}
}

func TestParseDateRangeInvalidOrder(t *testing.T) {
	_, err := ParseDateRange("2026-02-01", "2026-01-01", 30)
	if err == nil {
		t.Fatal("expected error when from is after to")
	}
}

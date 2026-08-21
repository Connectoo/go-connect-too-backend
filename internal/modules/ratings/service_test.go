package ratings

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type mockRatingStore struct {
	summaries map[uuid.UUID]*Summary
}

func (m *mockRatingStore) RecalculateAndUpdate(_ context.Context, employeeID uuid.UUID) (*Summary, error) {
	summary, ok := m.summaries[employeeID]
	if !ok {
		return &Summary{TotalReviews: 0}, nil
	}
	return summary, nil
}

func TestRefreshEmployeeRating(t *testing.T) {
	employeeID := uuid.New()
	avg := 4.5
	store := &mockRatingStore{
		summaries: map[uuid.UUID]*Summary{
			employeeID: {AverageRating: &avg, TotalReviews: 2},
		},
	}
	svc := NewService(store)

	summary, err := svc.RefreshEmployeeRating(context.Background(), employeeID)
	if err != nil {
		t.Fatalf("RefreshEmployeeRating() error = %v", err)
	}
	if summary.TotalReviews != 2 || summary.AverageRating == nil || *summary.AverageRating != 4.5 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

package ratings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Summary holds computed employee rating stats.
type Summary struct {
	AverageRating *float64
	TotalReviews  int
}

// Repository updates denormalized employee rating fields.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a ratings repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ComputeSummary calculates average rating and count from approved reviews.
func (r *Repository) ComputeSummary(ctx context.Context, employeeID uuid.UUID) (*Summary, error) {
	query := `
		SELECT AVG(rating)::float8, COUNT(*)
		FROM reviews
		WHERE employee_id = $1 AND status = 'approved'`

	var avg sql.NullFloat64
	var count int
	if err := r.db.QueryRowContext(ctx, query, employeeID).Scan(&avg, &count); err != nil {
		return nil, fmt.Errorf("compute rating summary: %w", err)
	}

	summary := &Summary{TotalReviews: count}
	if avg.Valid {
		value := avg.Float64
		summary.AverageRating = &value
	}
	return summary, nil
}

// UpdateEmployeeProfile writes computed rating stats to employee_profiles.
func (r *Repository) UpdateEmployeeProfile(ctx context.Context, employeeID uuid.UUID, summary *Summary) error {
	query := `
		UPDATE employee_profiles
		SET average_rating = $2,
		    total_reviews = $3,
		    updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, employeeID, summary.AverageRating, summary.TotalReviews)
	if err != nil {
		return fmt.Errorf("update employee rating: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("employee rating rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("employee profile not found")
	}
	return nil
}

// RecalculateAndUpdate recomputes and persists employee rating stats.
func (r *Repository) RecalculateAndUpdate(ctx context.Context, employeeID uuid.UUID) (*Summary, error) {
	summary, err := r.ComputeSummary(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if err := r.UpdateEmployeeProfile(ctx, employeeID, summary); err != nil {
		return nil, err
	}
	return summary, nil
}

// Service handles rating recalculation business logic.
type Service struct {
	store RatingStore
}

// RatingStore recalculates employee ratings.
type RatingStore interface {
	RecalculateAndUpdate(ctx context.Context, employeeID uuid.UUID) (*Summary, error)
}

// NewService creates a ratings service.
func NewService(store RatingStore) *Service {
	return &Service{store: store}
}

// RefreshEmployeeRating recomputes rating stats after review moderation changes.
func (s *Service) RefreshEmployeeRating(ctx context.Context, employeeID uuid.UUID) (*Summary, error) {
	return s.store.RecalculateAndUpdate(ctx, employeeID)
}

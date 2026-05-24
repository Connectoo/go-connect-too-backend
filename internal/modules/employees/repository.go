package employees

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository persists employee profiles.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an employee repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const profileColumns = `
	id, user_id, display_name, phone, bio, experience_years, profile_photo_url,
	location_text, latitude, longitude, service_area_radius_km, languages, skills,
	verification_status, created_at, updated_at`

// CreateForUserInTx inserts a profile row for a newly registered employee user.
func (r *Repository) CreateForUserInTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, at time.Time) error {
	query := `
		INSERT INTO employee_profiles (id, user_id, verification_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := tx.ExecContext(ctx, query, uuid.New(), userID, VerificationPending, at, at)
	if err != nil {
		return fmt.Errorf("insert employee profile: %w", err)
	}
	return nil
}

// GetByUserID loads a profile for the given user.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	query := `SELECT` + profileColumns + ` FROM employee_profiles WHERE user_id = $1`

	row := r.db.QueryRowContext(ctx, query, userID)
	profile, err := scanProfile(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get employee profile by user id: %w", err)
	}
	return profile, nil
}

// GetByID loads a profile by primary key.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
	query := `SELECT` + profileColumns + ` FROM employee_profiles WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	profile, err := scanProfile(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get employee profile by id: %w", err)
	}
	return profile, nil
}

// UpdateByUserID replaces editable profile fields for the given user.
func (r *Repository) UpdateByUserID(ctx context.Context, userID uuid.UUID, profile *Profile, at time.Time) (*Profile, error) {
	query := `
		UPDATE employee_profiles
		SET display_name = $2,
		    phone = $3,
		    bio = $4,
		    experience_years = $5,
		    profile_photo_url = $6,
		    location_text = $7,
		    latitude = $8,
		    longitude = $9,
		    service_area_radius_km = $10,
		    languages = $11,
		    skills = $12,
		    verification_status = $13,
		    updated_at = $14
		WHERE user_id = $1
		RETURNING` + profileColumns

	languages := pgtype.FlatArray[string](profile.Languages)
	skills := pgtype.FlatArray[string](profile.Skills)

	row := r.db.QueryRowContext(ctx, query,
		userID,
		profile.DisplayName,
		profile.Phone,
		profile.Bio,
		profile.ExperienceYears,
		profile.ProfilePhotoURL,
		profile.LocationText,
		profile.Latitude,
		profile.Longitude,
		profile.ServiceAreaRadiusKm,
		languages,
		skills,
		profile.VerificationStatus,
		at,
	)

	updated, err := scanProfile(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update employee profile: %w", err)
	}
	return updated, nil
}

// UpdateVerificationStatus sets approval state for a profile.
func (r *Repository) UpdateVerificationStatus(ctx context.Context, id uuid.UUID, status string, at time.Time) (*Profile, error) {
	query := `
		UPDATE employee_profiles
		SET verification_status = $2,
		    updated_at = $3
		WHERE id = $1
		RETURNING` + profileColumns

	row := r.db.QueryRowContext(ctx, query, id, status, at)
	profile, err := scanProfile(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update employee verification status: %w", err)
	}
	return profile, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProfile(row rowScanner) (*Profile, error) {
	var profile Profile
	var languages pgtype.FlatArray[string]
	var skills pgtype.FlatArray[string]

	err := row.Scan(
		&profile.ID,
		&profile.UserID,
		&profile.DisplayName,
		&profile.Phone,
		&profile.Bio,
		&profile.ExperienceYears,
		&profile.ProfilePhotoURL,
		&profile.LocationText,
		&profile.Latitude,
		&profile.Longitude,
		&profile.ServiceAreaRadiusKm,
		&languages,
		&skills,
		&profile.VerificationStatus,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	profile.Languages = []string(languages)
	profile.Skills = []string(skills)
	if profile.Languages == nil {
		profile.Languages = []string{}
	}
	if profile.Skills == nil {
		profile.Skills = []string{}
	}

	return &profile, nil
}

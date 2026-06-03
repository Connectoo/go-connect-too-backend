package public

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

// Repository loads public website data.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a public repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const publicProviderColumns = `
	ep.id, ep.display_name, ep.bio, ep.experience_years, ep.profile_photo_url,
	ep.location_text, ep.latitude, ep.longitude, ep.service_area_radius_km,
	ep.languages, ep.skills`

// ListApprovedProviders returns approved employee profiles for the public website.
func (r *Repository) ListApprovedProviders(ctx context.Context, limit int) ([]employees.Profile, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT` + publicProviderColumns + `
		FROM employee_profiles ep
		INNER JOIN users u ON u.id = ep.user_id
		WHERE ep.verification_status = $1
		  AND u.status = 'active'
		ORDER BY ep.created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, employees.VerificationApproved, limit)
	if err != nil {
		return nil, fmt.Errorf("list approved providers: %w", err)
	}
	defer rows.Close()

	var out []employees.Profile
	for rows.Next() {
		profile, err := scanPublicProvider(rows)
		if err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		out = append(out, *profile)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate providers: %w", err)
	}
	if out == nil {
		out = []employees.Profile{}
	}
	return out, nil
}

// CountApprovedProviders returns the number of active approved providers.
func (r *Repository) CountApprovedProviders(ctx context.Context) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM employee_profiles ep
		INNER JOIN users u ON u.id = ep.user_id
		WHERE ep.verification_status = $1 AND u.status = 'active'`
	if err := r.db.QueryRowContext(ctx, query, employees.VerificationApproved).Scan(&count); err != nil {
		return 0, fmt.Errorf("count approved providers: %w", err)
	}
	return count, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPublicProvider(row rowScanner) (*employees.Profile, error) {
	var profile employees.Profile
	var languages pgtype.FlatArray[string]
	var skills pgtype.FlatArray[string]

	err := row.Scan(
		&profile.ID,
		&profile.DisplayName,
		&profile.Bio,
		&profile.ExperienceYears,
		&profile.ProfilePhotoURL,
		&profile.LocationText,
		&profile.Latitude,
		&profile.Longitude,
		&profile.ServiceAreaRadiusKm,
		&languages,
		&skills,
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

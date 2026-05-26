package search

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
)

type serviceSearchRow struct {
	ID                  uuid.UUID
	EmployeeID          uuid.UUID
	CategoryID          uuid.UUID
	Title               string
	Description         *string
	Price               float64
	DurationMinutes     int
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	EmployeeDisplayName *string
	EmployeeLocation    *string
}

type employeeSearchRow struct {
	ID                  uuid.UUID
	DisplayName         *string
	Bio                 *string
	ExperienceYears     int
	ProfilePhotoURL     *string
	LocationText        *string
	Latitude            *float64
	Longitude           *float64
	ServiceAreaRadiusKm *float64
	Languages           []string
	Skills              []string
	DistanceKm          *float64
}

// Repository runs PostgreSQL search queries.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a search repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SearchServices finds active services from approved employees.
func (r *Repository) SearchServices(ctx context.Context, params ServiceSearchParams) ([]serviceSearchRow, error) {
	limit := clampLimit(params.Limit)

	qb := newQueryBuilder()
	qb.add(employees.VerificationApproved)

	query := `
		SELECT
			es.id, es.employee_id, es.category_id, es.title, es.description, es.price,
			es.duration_minutes, es.is_active, es.created_at, es.updated_at,
			ep.display_name, ep.location_text
		FROM employee_services es
		INNER JOIN employee_profiles ep ON ep.id = es.employee_id
		WHERE es.is_active = true
		  AND ep.verification_status = $1`

	if params.Query != "" {
		p := qb.next()
		query += fmt.Sprintf(" AND (es.title ILIKE $%d OR COALESCE(es.description, '') ILIKE $%d)", p, p)
		qb.add("%" + params.Query + "%")
	}
	if params.CategoryID != nil {
		query += fmt.Sprintf(" AND es.category_id = $%d", qb.next())
		qb.add(*params.CategoryID)
	}
	if params.Location != "" {
		query += fmt.Sprintf(" AND COALESCE(ep.location_text, '') ILIKE $%d", qb.next())
		qb.add("%" + params.Location + "%")
	}
	if params.MinPrice != nil {
		query += fmt.Sprintf(" AND es.price >= $%d", qb.next())
		qb.add(*params.MinPrice)
	}
	if params.MaxPrice != nil {
		query += fmt.Sprintf(" AND es.price <= $%d", qb.next())
		qb.add(*params.MaxPrice)
	}

	query += " " + serviceOrderBy(params.Sort)
	query += fmt.Sprintf(" LIMIT $%d", qb.next())
	qb.add(limit)

	rows, err := r.db.QueryContext(ctx, query, qb.args...)
	if err != nil {
		return nil, fmt.Errorf("search services: %w", err)
	}
	defer rows.Close()

	var out []serviceSearchRow
	for rows.Next() {
		row, err := scanServiceSearchRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate service search rows: %w", err)
	}
	if out == nil {
		out = []serviceSearchRow{}
	}
	return out, nil
}

// SearchEmployees finds approved employee profiles.
func (r *Repository) SearchEmployees(ctx context.Context, params EmployeeSearchParams) ([]employeeSearchRow, error) {
	limit := clampLimit(params.Limit)
	hasCoords := params.Latitude != nil && params.Longitude != nil

	qb := newQueryBuilder()
	qb.add(employees.VerificationApproved)

	distanceSelect := ""
	var latPos, lonPos int
	if hasCoords {
		latPos = qb.next()
		lonPos = qb.next()
		qb.add(*params.Latitude)
		qb.add(*params.Longitude)
		distanceSelect = fmt.Sprintf(`, (
			6371 * acos(
				LEAST(1, GREATEST(-1,
					cos(radians($%d)) * cos(radians(ep.latitude)) *
					cos(radians(ep.longitude) - radians($%d)) +
					sin(radians($%d)) * sin(radians(ep.latitude))
				))
			)
		) AS distance_km`, latPos, lonPos, latPos)
	}

	query := `
		SELECT DISTINCT
			ep.id, ep.display_name, ep.bio, ep.experience_years, ep.profile_photo_url,
			ep.location_text, ep.latitude, ep.longitude, ep.service_area_radius_km,
			ep.languages, ep.skills` + distanceSelect + `
		FROM employee_profiles ep`

	if params.CategoryID != nil {
		query += `
		INNER JOIN employee_services es ON es.employee_id = ep.id AND es.is_active = true`
	}

	query += ` WHERE ep.verification_status = $1`

	if params.CategoryID != nil {
		query += fmt.Sprintf(" AND es.category_id = $%d", qb.next())
		qb.add(*params.CategoryID)
	}
	if params.Query != "" {
		p := qb.next()
		query += fmt.Sprintf(` AND (
			COALESCE(ep.display_name, '') ILIKE $%[1]d OR
			COALESCE(ep.bio, '') ILIKE $%[1]d OR
			EXISTS (SELECT 1 FROM unnest(ep.skills) skill WHERE skill ILIKE $%[1]d)
		)`, p)
		qb.add("%" + params.Query + "%")
	}
	if hasCoords && params.RadiusKm != nil {
		query += fmt.Sprintf(` AND ep.latitude IS NOT NULL AND ep.longitude IS NOT NULL
			AND (
				6371 * acos(
					LEAST(1, GREATEST(-1,
						cos(radians($%d)) * cos(radians(ep.latitude)) *
						cos(radians(ep.longitude) - radians($%d)) +
						sin(radians($%d)) * sin(radians(ep.latitude))
					))
				)
			) <= $%d`, latPos, lonPos, latPos, qb.next())
		qb.add(*params.RadiusKm)
	}

	query += " " + employeeOrderBy(params.Sort, hasCoords)
	query += fmt.Sprintf(" LIMIT $%d", qb.next())
	qb.add(limit)

	rows, err := r.db.QueryContext(ctx, query, qb.args...)
	if err != nil {
		return nil, fmt.Errorf("search employees: %w", err)
	}
	defer rows.Close()

	var out []employeeSearchRow
	for rows.Next() {
		row, err := scanEmployeeSearchRow(rows, hasCoords)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate employee search rows: %w", err)
	}
	if out == nil {
		out = []employeeSearchRow{}
	}
	return out, nil
}

type queryBuilder struct {
	pos  int
	args []any
}

func newQueryBuilder() *queryBuilder {
	return &queryBuilder{pos: 1, args: []any{}}
}

func (q *queryBuilder) next() int {
	q.pos++
	return q.pos
}

func (q *queryBuilder) add(arg any) {
	q.args = append(q.args, arg)
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func serviceOrderBy(sort string) string {
	switch sort {
	case "price_asc":
		return "ORDER BY es.price ASC, es.created_at DESC"
	case "price_desc":
		return "ORDER BY es.price DESC, es.created_at DESC"
	case "name_asc":
		return "ORDER BY es.title ASC"
	default:
		return "ORDER BY es.created_at DESC"
	}
}

func employeeOrderBy(sort string, hasCoords bool) string {
	switch sort {
	case "distance":
		if hasCoords {
			return "ORDER BY distance_km ASC NULLS LAST"
		}
		return "ORDER BY ep.display_name ASC NULLS LAST"
	default:
		return "ORDER BY ep.display_name ASC NULLS LAST"
	}
}

func scanServiceSearchRow(row *sql.Rows) (serviceSearchRow, error) {
	var item serviceSearchRow
	var description sql.NullString
	var displayName, location sql.NullString

	err := row.Scan(
		&item.ID,
		&item.EmployeeID,
		&item.CategoryID,
		&item.Title,
		&description,
		&item.Price,
		&item.DurationMinutes,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
		&displayName,
		&location,
	)
	if err != nil {
		return serviceSearchRow{}, fmt.Errorf("scan service search row: %w", err)
	}
	if description.Valid {
		item.Description = &description.String
	}
	if displayName.Valid {
		item.EmployeeDisplayName = &displayName.String
	}
	if location.Valid {
		item.EmployeeLocation = &location.String
	}
	return item, nil
}

func scanEmployeeSearchRow(row *sql.Rows, withDistance bool) (employeeSearchRow, error) {
	var item employeeSearchRow
	var displayName, bio, photo, location sql.NullString
	var lat, lon, radius sql.NullFloat64
	var languages pgtype.FlatArray[string]
	var skills pgtype.FlatArray[string]
	var distance sql.NullFloat64

	dest := []any{
		&item.ID,
		&displayName,
		&bio,
		&item.ExperienceYears,
		&photo,
		&location,
		&lat,
		&lon,
		&radius,
		&languages,
		&skills,
	}
	if withDistance {
		dest = append(dest, &distance)
	}

	if err := row.Scan(dest...); err != nil {
		return employeeSearchRow{}, fmt.Errorf("scan employee search row: %w", err)
	}

	if displayName.Valid {
		item.DisplayName = &displayName.String
	}
	if bio.Valid {
		item.Bio = &bio.String
	}
	if photo.Valid {
		item.ProfilePhotoURL = &photo.String
	}
	if location.Valid {
		item.LocationText = &location.String
	}
	if lat.Valid {
		item.Latitude = &lat.Float64
	}
	if lon.Valid {
		item.Longitude = &lon.Float64
	}
	if radius.Valid {
		item.ServiceAreaRadiusKm = &radius.Float64
	}
	item.Languages = []string(languages)
	if item.Languages == nil {
		item.Languages = []string{}
	}
	item.Skills = []string(skills)
	if item.Skills == nil {
		item.Skills = []string{}
	}
	if distance.Valid {
		item.DistanceKm = &distance.Float64
	}
	return item, nil
}

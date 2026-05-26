package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const addressColumns = `
	id, user_id, label, address_line, city, state, country, pincode,
	latitude, longitude, is_default, created_at, updated_at`

// ListAddressesByUserID returns addresses for a user.
func (r *Repository) ListAddressesByUserID(ctx context.Context, userID uuid.UUID) ([]Address, error) {
	query := `SELECT` + addressColumns + ` FROM user_addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user addresses: %w", err)
	}
	defer rows.Close()

	var out []Address
	for rows.Next() {
		addr, err := scanAddress(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user address: %w", err)
		}
		out = append(out, *addr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user addresses: %w", err)
	}
	if out == nil {
		out = []Address{}
	}
	return out, nil
}

// CreateAddress inserts an address for a user.
func (r *Repository) CreateAddress(ctx context.Context, addr *Address) (*Address, error) {
	query := `
		INSERT INTO user_addresses (
			id, user_id, label, address_line, city, state, country, pincode,
			latitude, longitude, is_default, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING` + addressColumns

	row := r.db.QueryRowContext(ctx, query,
		addr.ID,
		addr.UserID,
		addr.Label,
		addr.AddressLine,
		addr.City,
		addr.State,
		addr.Country,
		addr.Pincode,
		addr.Latitude,
		addr.Longitude,
		addr.IsDefault,
		addr.CreatedAt,
		addr.UpdatedAt,
	)

	created, err := scanAddress(row)
	if err != nil {
		return nil, fmt.Errorf("insert user address: %w", err)
	}
	return created, nil
}

// GetAddressByID returns an address when it belongs to the user.
func (r *Repository) GetAddressByID(ctx context.Context, userID, addressID uuid.UUID) (*Address, error) {
	query := `SELECT` + addressColumns + ` FROM user_addresses WHERE id = $1 AND user_id = $2`

	row := r.db.QueryRowContext(ctx, query, addressID, userID)
	addr, err := scanAddress(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAddressNotFound
		}
		return nil, fmt.Errorf("get user address: %w", err)
	}
	return addr, nil
}

// UpdateAddress replaces address fields for the owner.
func (r *Repository) UpdateAddress(ctx context.Context, addr *Address, at time.Time) (*Address, error) {
	query := `
		UPDATE user_addresses
		SET label = $3,
		    address_line = $4,
		    city = $5,
		    state = $6,
		    country = $7,
		    pincode = $8,
		    latitude = $9,
		    longitude = $10,
		    is_default = $11,
		    updated_at = $12
		WHERE id = $1 AND user_id = $2
		RETURNING` + addressColumns

	row := r.db.QueryRowContext(ctx, query,
		addr.ID,
		addr.UserID,
		addr.Label,
		addr.AddressLine,
		addr.City,
		addr.State,
		addr.Country,
		addr.Pincode,
		addr.Latitude,
		addr.Longitude,
		addr.IsDefault,
		at,
	)

	updated, err := scanAddress(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAddressNotFound
		}
		return nil, fmt.Errorf("update user address: %w", err)
	}
	return updated, nil
}

// DeleteAddress removes an address owned by the user.
func (r *Repository) DeleteAddress(ctx context.Context, userID, addressID uuid.UUID) error {
	query := `DELETE FROM user_addresses WHERE id = $1 AND user_id = $2`

	res, err := r.db.ExecContext(ctx, query, addressID, userID)
	if err != nil {
		return fmt.Errorf("delete user address: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted address rows: %w", err)
	}
	if rows == 0 {
		return ErrAddressNotFound
	}
	return nil
}

// ClearDefaultAddresses unsets default flag for all user addresses except optional skip id.
func (r *Repository) ClearDefaultAddresses(ctx context.Context, userID uuid.UUID, exceptID *uuid.UUID, at time.Time) error {
	query := `
		UPDATE user_addresses
		SET is_default = false,
		    updated_at = $2
		WHERE user_id = $1 AND is_default = true`
	args := []any{userID, at}

	if exceptID != nil {
		query += ` AND id <> $3`
		args = append(args, *exceptID)
	}

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("clear default addresses: %w", err)
	}
	return nil
}

func scanAddress(row rowScanner) (*Address, error) {
	var addr Address
	err := row.Scan(
		&addr.ID,
		&addr.UserID,
		&addr.Label,
		&addr.AddressLine,
		&addr.City,
		&addr.State,
		&addr.Country,
		&addr.Pincode,
		&addr.Latitude,
		&addr.Longitude,
		&addr.IsDefault,
		&addr.CreatedAt,
		&addr.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

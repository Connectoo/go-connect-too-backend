package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository persists notifications and device tokens.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a notifications repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const notificationColumns = `id, user_id, type, title, body, data, read_at, created_at`

// Create inserts a notification row.
func (r *Repository) Create(ctx context.Context, notification *Notification) (*Notification, error) {
	data, err := json.Marshal(notification.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal notification data: %w", err)
	}

	query := `
		INSERT INTO notifications (id, user_id, type, title, body, data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + notificationColumns

	row := r.db.QueryRowContext(
		ctx,
		query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Body,
		data,
		notification.CreatedAt,
	)

	created, err := scanNotification(row)
	if err != nil {
		return nil, fmt.Errorf("insert notification: %w", err)
	}
	return created, nil
}

// ListByUserID returns paginated notifications for a user.
func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]Notification, int, error) {
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	query := `
		SELECT ` + notificationColumns + `
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, userID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	items, err := scanNotifications(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetByID loads a notification by id.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Notification, error) {
	query := `SELECT ` + notificationColumns + ` FROM notifications WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanNotification(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get notification: %w", err)
	}
	return item, nil
}

// MarkRead sets read_at for a notification owned by the user.
func (r *Repository) MarkRead(ctx context.Context, id, userID uuid.UUID, at time.Time) (*Notification, error) {
	query := `
		UPDATE notifications
		SET read_at = $1
		WHERE id = $2 AND user_id = $3
		RETURNING ` + notificationColumns

	row := r.db.QueryRowContext(ctx, query, at, id, userID)
	item, err := scanNotification(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("mark notification read: %w", err)
	}
	return item, nil
}

// MarkAllRead sets read_at for all unread notifications owned by the user.
func (r *Repository) MarkAllRead(ctx context.Context, userID uuid.UUID, at time.Time) (int64, error) {
	query := `
		UPDATE notifications
		SET read_at = $1
		WHERE user_id = $2 AND read_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, at, userID)
	if err != nil {
		return 0, fmt.Errorf("mark all notifications read: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("mark all notifications read rows affected: %w", err)
	}
	return affected, nil
}

// UpsertDeviceToken registers or reactivates a device token for a user.
func (r *Repository) UpsertDeviceToken(ctx context.Context, token *DeviceToken) (*DeviceToken, error) {
	query := `
		INSERT INTO device_tokens (id, user_id, platform, token, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, TRUE, $5, $6)
		ON CONFLICT (user_id, token) DO UPDATE
		SET platform = EXCLUDED.platform,
		    is_active = TRUE,
		    updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, platform, token, is_active, created_at, updated_at`

	row := r.db.QueryRowContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.Platform,
		token.Token,
		token.CreatedAt,
		token.UpdatedAt,
	)

	var created DeviceToken
	if err := row.Scan(
		&created.ID,
		&created.UserID,
		&created.Platform,
		&created.Token,
		&created.IsActive,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("upsert device token: %w", err)
	}
	return &created, nil
}

// ListActiveDeviceTokens returns active tokens for a user.
func (r *Repository) ListActiveDeviceTokens(ctx context.Context, userID uuid.UUID) ([]DeviceToken, error) {
	query := `
		SELECT id, user_id, platform, token, is_active, created_at, updated_at
		FROM device_tokens
		WHERE user_id = $1 AND is_active = TRUE
		ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list device tokens: %w", err)
	}
	defer rows.Close()

	items := make([]DeviceToken, 0)
	for rows.Next() {
		var item DeviceToken
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Platform,
			&item.Token,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan device token: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device tokens: %w", err)
	}
	return items, nil
}

type notificationScanner interface {
	Scan(dest ...any) error
}

func scanNotification(row notificationScanner) (*Notification, error) {
	var item Notification
	var data []byte
	if err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Type,
		&item.Title,
		&item.Body,
		&data,
		&item.ReadAt,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &item.Data); err != nil {
			return nil, fmt.Errorf("unmarshal notification data: %w", err)
		}
	}
	if item.Data == nil {
		item.Data = map[string]any{}
	}
	return &item, nil
}

func scanNotifications(rows *sql.Rows) ([]Notification, error) {
	items := make([]Notification, 0)
	for rows.Next() {
		item, err := scanNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return items, nil
}

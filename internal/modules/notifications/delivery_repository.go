package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LogDelivery records a notification delivery attempt.
func (r *Repository) LogDelivery(ctx context.Context, delivery *Delivery) error {
	query := `
		INSERT INTO notification_deliveries (
			id, notification_id, user_id, channel, provider, status, error_message, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		delivery.ID,
		delivery.NotificationID,
		delivery.UserID,
		delivery.Channel,
		delivery.Provider,
		delivery.Status,
		delivery.ErrorMessage,
		delivery.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert notification delivery: %w", err)
	}
	return nil
}

// LogDeliveryNow records a delivery attempt with the current timestamp.
func (r *Repository) LogDeliveryNow(ctx context.Context, userID uuid.UUID, provider, status string, errMsg *string) error {
	return r.LogDelivery(ctx, &Delivery{
		ID:           uuid.New(),
		UserID:       userID,
		Channel:      DeliveryChannelPush,
		Provider:     provider,
		Status:       status,
		ErrorMessage: errMsg,
		CreatedAt:    time.Now().UTC(),
	})
}

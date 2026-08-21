package notifications

import (
	"time"

	"github.com/google/uuid"
)

// NotificationResponse is the API notification payload.
type NotificationResponse struct {
	ID        uuid.UUID      `json:"id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Data      map[string]any `json:"data"`
	ReadAt    *time.Time     `json:"read_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// RegisterDeviceTokenRequest registers or refreshes a device token.
type RegisterDeviceTokenRequest struct {
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

// DeviceTokenResponse is the API device token payload.
type DeviceTokenResponse struct {
	ID        uuid.UUID `json:"id"`
	Platform  string    `json:"platform"`
	Token     string    `json:"token"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toNotificationResponse(item *Notification) NotificationResponse {
	data := item.Data
	if data == nil {
		data = map[string]any{}
	}
	return NotificationResponse{
		ID:        item.ID,
		Type:      item.Type,
		Title:     item.Title,
		Body:      item.Body,
		Data:      data,
		ReadAt:    item.ReadAt,
		CreatedAt: item.CreatedAt,
	}
}

func toNotificationResponses(items []Notification) []NotificationResponse {
	out := make([]NotificationResponse, 0, len(items))
	for i := range items {
		out = append(out, toNotificationResponse(&items[i]))
	}
	return out
}

func toDeviceTokenResponse(item *DeviceToken) DeviceTokenResponse {
	return DeviceTokenResponse{
		ID:        item.ID,
		Platform:  item.Platform,
		Token:     item.Token,
		IsActive:  item.IsActive,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

package notifications

import (
	"context"
)

// PushMessage is sent to a device through a push provider.
type PushMessage struct {
	Title string
	Body  string
	Data  map[string]string
}

// PushProvider delivers mobile push notifications.
type PushProvider interface {
	SendToUser(ctx context.Context, userID string, message PushMessage) error
}

// NoopPushProvider discards push notifications until FCM is configured.
type NoopPushProvider struct{}

// SendToUser implements PushProvider.
func (NoopPushProvider) SendToUser(context.Context, string, PushMessage) error {
	return nil
}

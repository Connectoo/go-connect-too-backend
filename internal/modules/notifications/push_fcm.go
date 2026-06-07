package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2/google"
)

const fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

// DeliveryLogger persists push delivery attempts.
type DeliveryLogger interface {
	LogDeliveryNow(ctx context.Context, userID uuid.UUID, provider, status string, errMsg *string) error
	ListActiveDeviceTokens(ctx context.Context, userID uuid.UUID) ([]DeviceToken, error)
}

// FCMPushProvider sends push notifications through Firebase Cloud Messaging HTTP v1.
type FCMPushProvider struct {
	projectID   string
	credentials []byte
	client      *http.Client
	tokens      DeliveryLogger
	log         *slog.Logger
}

// NewFCMPushProvider creates an FCM push provider.
func NewFCMPushProvider(projectID string, credentialsJSON []byte, tokens DeliveryLogger, log *slog.Logger) *FCMPushProvider {
	return &FCMPushProvider{
		projectID:   projectID,
		credentials: credentialsJSON,
		client:      &http.Client{Timeout: 10 * time.Second},
		tokens:      tokens,
		log:         log,
	}
}

// SendToUser delivers a push notification to all active device tokens for a user.
func (p *FCMPushProvider) SendToUser(ctx context.Context, userID string, message PushMessage) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}
	if p.tokens == nil {
		return p.logDelivery(ctx, uid, DeliveryStatusSkipped, strPtr("device token store not configured"))
	}

	deviceTokens, err := p.tokens.ListActiveDeviceTokens(ctx, uid)
	if err != nil {
		_ = p.logDelivery(ctx, uid, DeliveryStatusFailed, strPtr(err.Error()))
		return err
	}
	if len(deviceTokens) == 0 {
		return p.logDelivery(ctx, uid, DeliveryStatusSkipped, strPtr("no active device tokens"))
	}

	accessToken, err := p.accessToken(ctx)
	if err != nil {
		_ = p.logDelivery(ctx, uid, DeliveryStatusFailed, strPtr(err.Error()))
		return err
	}

	var lastErr error
	sent := false
	for _, token := range deviceTokens {
		if err := p.send(ctx, accessToken, token.Token, message); err != nil {
			lastErr = err
			if p.log != nil {
				p.log.Error("fcm send failed", slog.String("error", err.Error()), slog.String("user_id", userID))
			}
			continue
		}
		sent = true
	}

	if sent {
		return p.logDelivery(ctx, uid, DeliveryStatusSent, nil)
	}
	if lastErr != nil {
		msg := lastErr.Error()
		_ = p.logDelivery(ctx, uid, DeliveryStatusFailed, &msg)
		return lastErr
	}
	return nil
}

func (p *FCMPushProvider) accessToken(ctx context.Context) (string, error) {
	creds, err := google.CredentialsFromJSON(ctx, p.credentials, fcmScope)
	if err != nil {
		return "", fmt.Errorf("parse fcm credentials: %w", err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("fetch fcm access token: %w", err)
	}
	return token.AccessToken, nil
}

func (p *FCMPushProvider) send(ctx context.Context, accessToken, deviceToken string, message PushMessage) error {
	payload := map[string]any{
		"message": map[string]any{
			"token": deviceToken,
			"notification": map[string]string{
				"title": message.Title,
				"body":  message.Body,
			},
			"data": message.Data,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", p.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("fcm send status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func (p *FCMPushProvider) logDelivery(ctx context.Context, userID uuid.UUID, status string, errMsg *string) error {
	if p.tokens == nil {
		return nil
	}
	return p.tokens.LogDeliveryNow(ctx, userID, "fcm", status, errMsg)
}

func strPtr(value string) *string {
	return &value
}

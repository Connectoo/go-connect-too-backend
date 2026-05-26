package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const ProviderRazorpay = "razorpay"

var ErrGatewayNotConfigured = errors.New("payment gateway not configured")

type Gateway interface {
	CreateOrder(ctx context.Context, req CreateOrderInput) (*GatewayOrder, error)
	VerifyPayment(orderID, paymentID, signature string) bool
	VerifyWebhook(payload []byte, signature string) bool
}

type CreateOrderInput struct {
	Amount   int64
	Currency string
	Receipt  string
	Notes    map[string]string
}

type GatewayOrder struct {
	ProviderOrderID string
	Amount          int64
	Currency        string
	RawResponse     []byte
}

type RazorpayGateway struct {
	keyID         string
	keySecret     string
	webhookSecret string
	baseURL       string
	httpClient    *http.Client
}

func NewRazorpayGateway(keyID, keySecret, webhookSecret string) *RazorpayGateway {
	return &RazorpayGateway{
		keyID:         strings.TrimSpace(keyID),
		keySecret:     strings.TrimSpace(keySecret),
		webhookSecret: strings.TrimSpace(webhookSecret),
		baseURL:       "https://api.razorpay.com/v1",
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *RazorpayGateway) KeyID() string { return g.keyID }

func (g *RazorpayGateway) CreateOrder(ctx context.Context, input CreateOrderInput) (*GatewayOrder, error) {
	if g.keyID == "" || g.keySecret == "" {
		return nil, ErrGatewayNotConfigured
	}
	if input.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be greater than zero", ErrValidation)
	}

	body := map[string]any{
		"amount":          input.Amount,
		"currency":        input.Currency,
		"receipt":         input.Receipt,
		"payment_capture": 1,
		"notes":           input.Notes,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal razorpay order: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/orders", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build razorpay order request: %w", err)
	}
	req.SetBasicAuth(g.keyID, g.keySecret)
	req.Header.Set("Content-Type", "application/json")

	res, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create razorpay order: %w", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read razorpay order response: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("create razorpay order: provider status %d", res.StatusCode)
	}

	var out struct {
		ID       string `json:"id"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode razorpay order response: %w", err)
	}
	if out.ID == "" {
		return nil, fmt.Errorf("create razorpay order: missing provider order id")
	}
	return &GatewayOrder{ProviderOrderID: out.ID, Amount: out.Amount, Currency: out.Currency, RawResponse: raw}, nil
}

func (g *RazorpayGateway) VerifyPayment(orderID, paymentID, signature string) bool {
	if g.keySecret == "" || orderID == "" || paymentID == "" || signature == "" {
		return false
	}
	return verifyHMACSHA256([]byte(orderID+"|"+paymentID), signature, g.keySecret)
}

func (g *RazorpayGateway) VerifyWebhook(payload []byte, signature string) bool {
	if g.webhookSecret == "" || len(payload) == 0 || signature == "" {
		return false
	}
	return verifyHMACSHA256(payload, signature, g.webhookSecret)
}

func verifyHMACSHA256(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := mac.Sum(nil)
	actual, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	return hmac.Equal(actual, expected)
}

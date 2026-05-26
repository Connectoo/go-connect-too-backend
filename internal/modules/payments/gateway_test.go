package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestRazorpayGateway_VerifyPayment(t *testing.T) {
	gateway := NewRazorpayGateway("key", "secret", "webhook-secret")
	sig := sign([]byte("order_123|pay_123"), "secret")

	if !gateway.VerifyPayment("order_123", "pay_123", sig) {
		t.Fatal("VerifyPayment() = false, want true")
	}
	if gateway.VerifyPayment("order_123", "pay_456", sig) {
		t.Fatal("VerifyPayment() = true for tampered payment id")
	}
}

func TestRazorpayGateway_VerifyWebhook(t *testing.T) {
	gateway := NewRazorpayGateway("key", "secret", "webhook-secret")
	payload := []byte(`{"event":"payment.captured"}`)
	sig := sign(payload, "webhook-secret")

	if !gateway.VerifyWebhook(payload, sig) {
		t.Fatal("VerifyWebhook() = false, want true")
	}
	if gateway.VerifyWebhook([]byte(`{"event":"payment.failed"}`), sig) {
		t.Fatal("VerifyWebhook() = true for tampered payload")
	}
}

func sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

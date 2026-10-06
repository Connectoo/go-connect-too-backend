package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const twilioDefaultTemplate = "Your Go Connect verification code is {code}. It expires in a few minutes."

// TwilioOTPSender delivers OTP codes as SMS via the Twilio Messages REST API.
// We keep our own OTP generation/verification; Twilio is only the transport.
// The Twilio Go SDK is intentionally NOT used — plain net/http keeps the
// dependency surface small.
type TwilioOTPSender struct {
	accountSID          string
	authToken           string
	fromNumber          string
	messagingServiceSID string
	template            string
	client              *http.Client
	log                 *slog.Logger
}

// NewTwilioOTPSender builds a Twilio SMS OTP sender. When client is nil a
// timeout-bounded default client is used. When template is empty the default
// SMS template is applied.
func NewTwilioOTPSender(accountSID, authToken, fromNumber, messagingServiceSID, template string, client *http.Client, log *slog.Logger) *TwilioOTPSender {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if strings.TrimSpace(template) == "" {
		template = twilioDefaultTemplate
	}
	return &TwilioOTPSender{
		accountSID:          accountSID,
		authToken:           authToken,
		fromNumber:          fromNumber,
		messagingServiceSID: messagingServiceSID,
		template:            template,
		client:              client,
		log:                 log,
	}
}

// Enabled implements OTPSender.
func (s *TwilioOTPSender) Enabled() bool { return true }

// Send implements OTPSender by delivering the code via Twilio SMS. It never logs
// the OTP code, the auth token, or any other secret.
func (s *TwilioOTPSender) Send(phone, code string) error {
	to := toE164(phone)
	body := strings.ReplaceAll(s.template, "{code}", code)

	form := url.Values{}
	form.Set("To", to)
	form.Set("Body", body)
	// Prefer a Messaging Service SID when configured, otherwise a from number.
	if s.messagingServiceSID != "" {
		form.Set("MessagingServiceSid", s.messagingServiceSID)
	} else {
		form.Set("From", s.fromNumber)
	}

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.accountSID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("twilio build request: %w", err)
	}
	req.SetBasicAuth(s.accountSID, s.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("twilio send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		var twErr struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(respBody, &twErr)
		if s.log != nil {
			// Log provider failure only — never the OTP code or any secret.
			s.log.Warn("twilio otp send failed",
				slog.Int("status", resp.StatusCode),
				slog.Int("twilio_code", twErr.Code),
			)
		}
		return fmt.Errorf("twilio send failed: status %d: %s (code %d)", resp.StatusCode, twErr.Message, twErr.Code)
	}

	if s.log != nil {
		// Debug-level success breadcrumb — phone only, never the OTP code.
		s.log.Debug("otp sms sent", slog.String("phone", to))
	}
	return nil
}

// toE164 normalizes a phone number for Twilio, which REQUIRES a single leading
// '+' (the opposite of MSG91). It strips whitespace and any leading '+' or '00'
// international prefix, then prepends exactly one '+'. Remaining digits are kept
// as-is (stored phones already include the country code).
func toE164(phone string) string {
	p := strings.ReplaceAll(phone, " ", "")
	p = strings.TrimSpace(p)
	// Drop any leading '+' then any leading '00' international prefix.
	p = strings.TrimPrefix(p, "+")
	p = strings.TrimPrefix(p, "00")
	return "+" + p
}

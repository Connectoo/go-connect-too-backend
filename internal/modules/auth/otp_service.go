package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
)

// OTPStore manages one-time phone login codes.
type OTPStore interface {
	CreateOTPCode(ctx context.Context, code *OTPLoginCode) error
	GetActiveOTPByPhoneRole(ctx context.Context, phone, role string) (*OTPLoginCode, error)
	IncrementOTPAttempt(ctx context.Context, id uuid.UUID) (int, error)
	MarkOTPConsumed(ctx context.Context, id uuid.UUID, at time.Time) error
	CountOTPRequestsSince(ctx context.Context, phone, role string, since time.Time) (int, error)
	LatestOTPCreatedAt(ctx context.Context, phone, role string) (*time.Time, error)
}

// OTPSender delivers one-time login codes (e.g. via SMS).
type OTPSender interface {
	Enabled() bool
	Send(phone, code string) error
}

// NoopOTPSender discards OTP codes until a real SMS provider is configured.
type NoopOTPSender struct{}

// Enabled implements OTPSender.
func (NoopOTPSender) Enabled() bool { return false }

// Send implements OTPSender.
func (NoopOTPSender) Send(string, string) error { return nil }

func (s *Service) otpCodeLength() int {
	if s.cfg != nil && s.cfg.OTPCodeLength > 0 {
		return s.cfg.OTPCodeLength
	}
	return 6
}

func (s *Service) otpCodeTTL() time.Duration {
	if s.cfg != nil && s.cfg.OTPCodeTTL > 0 {
		return s.cfg.OTPCodeTTL
	}
	return 5 * time.Minute
}

func (s *Service) otpResendCooldown() time.Duration {
	if s.cfg != nil && s.cfg.OTPResendCooldown > 0 {
		return s.cfg.OTPResendCooldown
	}
	return 60 * time.Second
}

func (s *Service) otpRequestWindow() time.Duration {
	if s.cfg != nil && s.cfg.OTPRequestWindow > 0 {
		return s.cfg.OTPRequestWindow
	}
	return 15 * time.Minute
}

func (s *Service) otpMaxPerWindow() int {
	if s.cfg != nil && s.cfg.OTPMaxPerWindow > 0 {
		return s.cfg.OTPMaxPerWindow
	}
	return 5
}

func (s *Service) otpMaxAttempts() int {
	if s.cfg != nil && s.cfg.OTPMaxAttempts > 0 {
		return s.cfg.OTPMaxAttempts
	}
	return 5
}

// RequestOTP generates and sends a one-time login code for a phone+role.
// It always returns a neutral result: callers must not learn whether the phone
// is registered. A code is only generated/stored/sent when a matching active
// user exists.
func (s *Service) RequestOTP(ctx context.Context, req OTPRequestRequest) error {
	if s.otp == nil {
		return ErrOTPNotConfigured
	}
	phone := strings.TrimSpace(req.Phone)
	role := strings.TrimSpace(req.Role)
	if phone == "" || role == "" {
		return ErrValidation
	}

	now := s.now()

	// Rate limiting: cooldown + max-per-window (keyed on phone+role).
	since := now.Add(-s.otpRequestWindow())
	count, err := s.otp.CountOTPRequestsSince(ctx, phone, role, since)
	if err != nil {
		return err
	}
	if count >= s.otpMaxPerWindow() {
		return ErrOTPRateLimited
	}
	latest, err := s.otp.LatestOTPCreatedAt(ctx, phone, role)
	if err != nil {
		return err
	}
	if latest != nil && now.Sub(*latest) < s.otpResendCooldown() {
		return ErrOTPRateLimited
	}

	user, err := s.users.GetByPhoneAndRole(ctx, phone, role)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			// Neutral response: do not reveal whether the phone is registered.
			return nil
		}
		return err
	}

	// Only allow OTP login for accounts whose status permits login.
	if user.Status != users.StatusActive || user.DeactivatedAt != nil {
		return nil
	}

	code, err := generateNumericCode(s.otpCodeLength())
	if err != nil {
		return err
	}
	hash := hashLifecycleToken(code, s.lifecycleSecret)

	if err := s.otp.CreateOTPCode(ctx, &OTPLoginCode{
		ID:        uuid.New(),
		Phone:     phone,
		Role:      role,
		CodeHash:  hash,
		ExpiresAt: now.Add(s.otpCodeTTL()),
		CreatedAt: now,
	}); err != nil {
		return err
	}

	if s.otpSender != nil && s.otpSender.Enabled() {
		if err := s.otpSender.Send(phone, code); err != nil {
			return err
		}
	}
	return nil
}

// VerifyOTP validates a one-time code and issues the standard token pair.
func (s *Service) VerifyOTP(ctx context.Context, req OTPVerifyRequest) (*AuthResponse, error) {
	if s.otp == nil {
		return nil, ErrOTPNotConfigured
	}
	phone := strings.TrimSpace(req.Phone)
	role := strings.TrimSpace(req.Role)
	code := strings.TrimSpace(req.Code)
	if phone == "" || role == "" || code == "" {
		return nil, ErrValidation
	}

	stored, err := s.otp.GetActiveOTPByPhoneRole(ctx, phone, role)
	if err != nil {
		return nil, err
	}

	now := s.now()
	if now.After(stored.ExpiresAt) {
		return nil, ErrOTPExpired
	}
	if stored.AttemptCount >= s.otpMaxAttempts() {
		return nil, ErrOTPTooManyAttempts
	}

	if hashLifecycleToken(code, s.lifecycleSecret) != stored.CodeHash {
		attempts, incErr := s.otp.IncrementOTPAttempt(ctx, stored.ID)
		if incErr != nil {
			return nil, incErr
		}
		if attempts >= s.otpMaxAttempts() {
			return nil, ErrOTPTooManyAttempts
		}
		return nil, ErrOTPInvalid
	}

	user, err := s.users.GetByPhoneAndRole(ctx, phone, role)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return nil, ErrOTPInvalid
		}
		return nil, err
	}
	if user.Status != users.StatusActive || user.DeactivatedAt != nil {
		return nil, ErrUserInactive
	}

	if err := s.otp.MarkOTPConsumed(ctx, stored.ID, now); err != nil {
		return nil, err
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:   toUserResponse(user),
		Tokens: *tokens,
	}, nil
}

// generateNumericCode returns a zero-padded random numeric code of the given length.
func generateNumericCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	digits := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0' + n.Int64())
	}
	return string(digits), nil
}

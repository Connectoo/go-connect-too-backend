package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

type mockOTPStore struct {
	codes []*OTPLoginCode
}

func newMockOTPStore() *mockOTPStore {
	return &mockOTPStore{}
}

func (m *mockOTPStore) CreateOTPCode(_ context.Context, code *OTPLoginCode) error {
	copy := *code
	m.codes = append(m.codes, &copy)
	return nil
}

func (m *mockOTPStore) GetActiveOTPByPhoneRole(_ context.Context, phone, role string) (*OTPLoginCode, error) {
	var latest *OTPLoginCode
	for _, c := range m.codes {
		if c.Phone != phone || c.Role != role || c.ConsumedAt != nil {
			continue
		}
		if latest == nil || c.CreatedAt.After(latest.CreatedAt) {
			latest = c
		}
	}
	if latest == nil {
		return nil, ErrOTPInvalid
	}
	copy := *latest
	return &copy, nil
}

func (m *mockOTPStore) IncrementOTPAttempt(_ context.Context, id uuid.UUID) (int, error) {
	for _, c := range m.codes {
		if c.ID == id {
			c.AttemptCount++
			return c.AttemptCount, nil
		}
	}
	return 0, ErrOTPInvalid
}

func (m *mockOTPStore) MarkOTPConsumed(_ context.Context, id uuid.UUID, at time.Time) error {
	for _, c := range m.codes {
		if c.ID == id {
			if c.ConsumedAt != nil {
				return ErrOTPInvalid
			}
			c.ConsumedAt = &at
			return nil
		}
	}
	return ErrOTPInvalid
}

func (m *mockOTPStore) CountOTPRequestsSince(_ context.Context, phone, role string, since time.Time) (int, error) {
	count := 0
	for _, c := range m.codes {
		if c.Phone == phone && c.Role == role && !c.CreatedAt.Before(since) {
			count++
		}
	}
	return count, nil
}

func (m *mockOTPStore) LatestOTPCreatedAt(_ context.Context, phone, role string) (*time.Time, error) {
	var latest *time.Time
	for _, c := range m.codes {
		if c.Phone != phone || c.Role != role {
			continue
		}
		if latest == nil || c.CreatedAt.After(*latest) {
			t := c.CreatedAt
			latest = &t
		}
	}
	return latest, nil
}

type mockOTPSender struct {
	enabled  bool
	lastCode string
	calls    int
}

func (m *mockOTPSender) Enabled() bool { return m.enabled }

func (m *mockOTPSender) Send(_ string, code string) error {
	m.calls++
	m.lastCode = code
	return nil
}

const (
	testOTPPhone = "+15551230000"
	testOTPRole  = users.RoleCustomer
)

func newOTPTestService(t *testing.T, store *mockUserStore, otpStore *mockOTPStore, sender *mockOTPSender) *Service {
	t.Helper()
	cfg := testConfig()
	svc := NewService(cfg, store, &mockRegistrar{users: store}, newMockRefreshStore(),
		security.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTAccessTTL),
		WithOTPStore(otpStore),
		WithOTPSender(sender),
	)
	svc.now = func() time.Time {
		return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	}
	return svc
}

func seedOTPUser(t *testing.T, store *mockUserStore, status string) *users.User {
	t.Helper()
	phone := testOTPPhone
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	user := &users.User{
		ID:        uuid.New(),
		Name:      "OTP User",
		Email:     "otp@example.com",
		Phone:     &phone,
		Role:      testOTPRole,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.users[userStoreKey(user.Email, user.Role)] = user
	store.byID[user.ID] = user
	store.byPhone[phoneStoreKey(phone, user.Role)] = user
	return user
}

func requestAndGetCode(t *testing.T, svc *Service, sender *mockOTPSender) string {
	t.Helper()
	if err := svc.RequestOTP(context.Background(), OTPRequestRequest{Phone: testOTPPhone, Role: testOTPRole}); err != nil {
		t.Fatalf("RequestOTP() error = %v", err)
	}
	if sender.lastCode == "" {
		t.Fatal("expected OTP sender to receive a code")
	}
	return sender.lastCode
}

func TestVerifyOTPSuccess(t *testing.T) {
	store := newMockUserStore()
	seedOTPUser(t, store, users.StatusActive)
	otpStore := newMockOTPStore()
	sender := &mockOTPSender{enabled: true}
	svc := newOTPTestService(t, store, otpStore, sender)

	code := requestAndGetCode(t, svc, sender)

	res, err := svc.VerifyOTP(context.Background(), OTPVerifyRequest{Phone: testOTPPhone, Role: testOTPRole, Code: code})
	if err != nil {
		t.Fatalf("VerifyOTP() error = %v", err)
	}
	if res.Tokens.AccessToken == "" || res.Tokens.RefreshToken == "" {
		t.Fatal("expected JWT token pair")
	}
	if res.User == nil || res.User.Role != testOTPRole {
		t.Fatalf("unexpected user: %+v", res.User)
	}
	if otpStore.codes[0].ConsumedAt == nil {
		t.Fatal("expected OTP code to be marked consumed")
	}
}

func TestVerifyOTPExpired(t *testing.T) {
	store := newMockUserStore()
	seedOTPUser(t, store, users.StatusActive)
	otpStore := newMockOTPStore()
	sender := &mockOTPSender{enabled: true}
	svc := newOTPTestService(t, store, otpStore, sender)

	code := requestAndGetCode(t, svc, sender)
	// advance time past the TTL
	svc.now = func() time.Time {
		return time.Date(2026, 5, 24, 12, 10, 0, 0, time.UTC)
	}

	_, err := svc.VerifyOTP(context.Background(), OTPVerifyRequest{Phone: testOTPPhone, Role: testOTPRole, Code: code})
	if err != ErrOTPExpired {
		t.Fatalf("error = %v, want %v", err, ErrOTPExpired)
	}
}

func TestVerifyOTPWrongCode(t *testing.T) {
	store := newMockUserStore()
	seedOTPUser(t, store, users.StatusActive)
	otpStore := newMockOTPStore()
	sender := &mockOTPSender{enabled: true}
	svc := newOTPTestService(t, store, otpStore, sender)

	requestAndGetCode(t, svc, sender)

	_, err := svc.VerifyOTP(context.Background(), OTPVerifyRequest{Phone: testOTPPhone, Role: testOTPRole, Code: "000000X"})
	if err != ErrOTPInvalid {
		t.Fatalf("error = %v, want %v", err, ErrOTPInvalid)
	}
	if otpStore.codes[0].AttemptCount != 1 {
		t.Fatalf("attempt_count = %d, want 1", otpStore.codes[0].AttemptCount)
	}
}

func TestVerifyOTPTooManyAttempts(t *testing.T) {
	store := newMockUserStore()
	seedOTPUser(t, store, users.StatusActive)
	otpStore := newMockOTPStore()
	sender := &mockOTPSender{enabled: true}
	svc := newOTPTestService(t, store, otpStore, sender)

	requestAndGetCode(t, svc, sender)

	var lastErr error
	for i := 0; i < svc.otpMaxAttempts(); i++ {
		_, lastErr = svc.VerifyOTP(context.Background(), OTPVerifyRequest{Phone: testOTPPhone, Role: testOTPRole, Code: "999999"})
	}
	if lastErr != ErrOTPTooManyAttempts {
		t.Fatalf("error = %v, want %v", lastErr, ErrOTPTooManyAttempts)
	}

	// subsequent verify (even with correct code) is rejected for too many attempts
	_, err := svc.VerifyOTP(context.Background(), OTPVerifyRequest{Phone: testOTPPhone, Role: testOTPRole, Code: sender.lastCode})
	if err != ErrOTPTooManyAttempts {
		t.Fatalf("error = %v, want %v", err, ErrOTPTooManyAttempts)
	}
}

func TestVerifyOTPReuseConsumed(t *testing.T) {
	store := newMockUserStore()
	seedOTPUser(t, store, users.StatusActive)
	otpStore := newMockOTPStore()
	sender := &mockOTPSender{enabled: true}
	svc := newOTPTestService(t, store, otpStore, sender)

	code := requestAndGetCode(t, svc, sender)

	if _, err := svc.VerifyOTP(context.Background(), OTPVerifyRequest{Phone: testOTPPhone, Role: testOTPRole, Code: code}); err != nil {
		t.Fatalf("first VerifyOTP() error = %v", err)
	}

	_, err := svc.VerifyOTP(context.Background(), OTPVerifyRequest{Phone: testOTPPhone, Role: testOTPRole, Code: code})
	if err != ErrOTPInvalid {
		t.Fatalf("error = %v, want %v (reuse of consumed code)", err, ErrOTPInvalid)
	}
}

func TestRequestOTPRateLimited(t *testing.T) {
	store := newMockUserStore()
	seedOTPUser(t, store, users.StatusActive)
	otpStore := newMockOTPStore()
	sender := &mockOTPSender{enabled: true}
	svc := newOTPTestService(t, store, otpStore, sender)

	// first request succeeds
	if err := svc.RequestOTP(context.Background(), OTPRequestRequest{Phone: testOTPPhone, Role: testOTPRole}); err != nil {
		t.Fatalf("first RequestOTP() error = %v", err)
	}
	// immediate second request hits the resend cooldown
	err := svc.RequestOTP(context.Background(), OTPRequestRequest{Phone: testOTPPhone, Role: testOTPRole})
	if err != ErrOTPRateLimited {
		t.Fatalf("error = %v, want %v", err, ErrOTPRateLimited)
	}
}

func TestRequestOTPMaxPerWindow(t *testing.T) {
	store := newMockUserStore()
	seedOTPUser(t, store, users.StatusActive)
	otpStore := newMockOTPStore()
	sender := &mockOTPSender{enabled: true}
	svc := newOTPTestService(t, store, otpStore, sender)

	base := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	// fill up the window, spacing past the cooldown each time
	for i := 0; i < svc.otpMaxPerWindow(); i++ {
		at := base.Add(time.Duration(i) * 2 * time.Minute)
		svc.now = func() time.Time { return at }
		if err := svc.RequestOTP(context.Background(), OTPRequestRequest{Phone: testOTPPhone, Role: testOTPRole}); err != nil {
			t.Fatalf("RequestOTP() #%d error = %v", i, err)
		}
	}
	// one more within the window exceeds the max-per-window limit
	svc.now = func() time.Time { return base.Add(time.Duration(svc.otpMaxPerWindow()) * 2 * time.Minute) }
	err := svc.RequestOTP(context.Background(), OTPRequestRequest{Phone: testOTPPhone, Role: testOTPRole})
	if err != ErrOTPRateLimited {
		t.Fatalf("error = %v, want %v", err, ErrOTPRateLimited)
	}
}

func TestRequestOTPUnknownPhoneNeutral(t *testing.T) {
	store := newMockUserStore()
	otpStore := newMockOTPStore()
	sender := &mockOTPSender{enabled: true}
	svc := newOTPTestService(t, store, otpStore, sender)

	// no user seeded: must return nil and store/send nothing
	if err := svc.RequestOTP(context.Background(), OTPRequestRequest{Phone: testOTPPhone, Role: testOTPRole}); err != nil {
		t.Fatalf("RequestOTP() error = %v, want nil (neutral)", err)
	}
	if len(otpStore.codes) != 0 {
		t.Fatalf("expected no stored codes, got %d", len(otpStore.codes))
	}
	if sender.calls != 0 {
		t.Fatalf("expected no codes sent, got %d", sender.calls)
	}
}

func TestVerifyOTPInactiveUser(t *testing.T) {
	store := newMockUserStore()
	otpStore := newMockOTPStore()
	sender := &mockOTPSender{enabled: true}
	svc := newOTPTestService(t, store, otpStore, sender)

	// user exists but is suspended; request stays neutral (no code generated)
	user := seedOTPUser(t, store, users.StatusSuspended)
	_ = user

	if err := svc.RequestOTP(context.Background(), OTPRequestRequest{Phone: testOTPPhone, Role: testOTPRole}); err != nil {
		t.Fatalf("RequestOTP() error = %v", err)
	}
	if len(otpStore.codes) != 0 {
		t.Fatalf("expected no code for suspended user, got %d", len(otpStore.codes))
	}

	// seed a code directly and flip the user active->suspended at verify time to
	// exercise the status check in VerifyOTP.
	now := svc.now()
	code := "123456"
	otpStore.codes = append(otpStore.codes, &OTPLoginCode{
		ID:        uuid.New(),
		Phone:     testOTPPhone,
		Role:      testOTPRole,
		CodeHash:  hashLifecycleToken(code, svc.lifecycleSecret),
		ExpiresAt: now.Add(5 * time.Minute),
		CreatedAt: now,
	})

	_, err := svc.VerifyOTP(context.Background(), OTPVerifyRequest{Phone: testOTPPhone, Role: testOTPRole, Code: code})
	if err != ErrUserInactive {
		t.Fatalf("error = %v, want %v", err, ErrUserInactive)
	}
}

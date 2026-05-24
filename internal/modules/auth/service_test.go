package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/config"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

type mockUserStore struct {
	users map[string]*users.User
	byID  map[uuid.UUID]*users.User
}

func userStoreKey(email, role string) string {
	return email + "\x00" + role
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{
		users: make(map[string]*users.User),
		byID:  make(map[uuid.UUID]*users.User),
	}
}

func (m *mockUserStore) Create(_ context.Context, user *users.User) error {
	key := userStoreKey(user.Email, user.Role)
	if _, ok := m.users[key]; ok {
		return users.ErrDuplicateEmail
	}
	copy := *user
	m.users[key] = &copy
	m.byID[user.ID] = &copy
	return nil
}

func (m *mockUserStore) GetByEmailAndRole(_ context.Context, email, role string) (*users.User, error) {
	user, ok := m.users[userStoreKey(email, role)]
	if !ok {
		return nil, users.ErrNotFound
	}
	copy := *user
	return &copy, nil
}

func (m *mockUserStore) GetByID(_ context.Context, id uuid.UUID) (*users.User, error) {
	user, ok := m.byID[id]
	if !ok {
		return nil, users.ErrNotFound
	}
	copy := *user
	return &copy, nil
}

type mockRefreshStore struct {
	tokens map[string]*RefreshToken
}

func newMockRefreshStore() *mockRefreshStore {
	return &mockRefreshStore{tokens: make(map[string]*RefreshToken)}
}

func (m *mockRefreshStore) CreateRefreshToken(_ context.Context, token *RefreshToken) error {
	copy := *token
	m.tokens[token.TokenHash] = &copy
	return nil
}

func (m *mockRefreshStore) GetActiveByHash(_ context.Context, tokenHash string) (*RefreshToken, error) {
	token, ok := m.tokens[tokenHash]
	if !ok || token.RevokedAt != nil {
		return nil, ErrInvalidToken
	}
	copy := *token
	return &copy, nil
}

func (m *mockRefreshStore) RevokeByHash(_ context.Context, tokenHash string, revokedAt time.Time) error {
	token, ok := m.tokens[tokenHash]
	if !ok || token.RevokedAt != nil {
		return ErrInvalidToken
	}
	token.RevokedAt = &revokedAt
	return nil
}

func testConfig() *config.Config {
	return &config.Config{
		JWTAccessSecret:  "test-access-secret-min-32-characters",
		JWTRefreshSecret: "test-refresh-secret-min-32-characters",
		JWTAccessTTL:     15 * time.Minute,
		JWTRefreshTTL:    7 * 24 * time.Hour,
	}
}

type mockRegistrar struct {
	users *mockUserStore
}

func (m *mockRegistrar) RegisterCustomer(ctx context.Context, user *users.User) error {
	return m.users.Create(ctx, user)
}

func (m *mockRegistrar) RegisterEmployee(ctx context.Context, user *users.User) error {
	return m.users.Create(ctx, user)
}

func newTestService(t *testing.T, userStore UserStore, refreshStore RefreshStore) *Service {
	t.Helper()
	store, ok := userStore.(*mockUserStore)
	if !ok {
		t.Fatal("tests require *mockUserStore")
	}
	cfg := testConfig()
	svc := NewService(cfg, userStore, &mockRegistrar{users: store}, refreshStore, security.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTAccessTTL))
	svc.now = func() time.Time {
		return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	}
	return svc
}

func TestRegisterCustomerSuccess(t *testing.T) {
	svc := newTestService(t, newMockUserStore(), newMockRefreshStore())

	res, err := svc.RegisterCustomer(context.Background(), RegisterCustomerRequest{
		Name:     "Jane Customer",
		Email:    "jane@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("RegisterCustomer() error = %v", err)
	}
	if res.User == nil || res.User.Role != users.RoleCustomer {
		t.Fatalf("unexpected user: %+v", res.User)
	}
	if res.Tokens.AccessToken == "" || res.Tokens.RefreshToken == "" {
		t.Fatal("expected tokens")
	}
}

func TestRegisterSameEmailDifferentRoles(t *testing.T) {
	svc := newTestService(t, newMockUserStore(), newMockRefreshStore())

	email := "shared@example.com"
	if _, err := svc.RegisterEmployee(context.Background(), RegisterEmployeeRequest{
		Name:     "Employee",
		Email:    email,
		Password: "password123",
	}); err != nil {
		t.Fatalf("RegisterEmployee() error = %v", err)
	}

	res, err := svc.RegisterCustomer(context.Background(), RegisterCustomerRequest{
		Name:     "Customer",
		Email:    email,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("RegisterCustomer() error = %v", err)
	}
	if res.User == nil || res.User.Role != users.RoleCustomer {
		t.Fatalf("unexpected user: %+v", res.User)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	store := newMockUserStore()
	svc := newTestService(t, store, newMockRefreshStore())

	req := RegisterCustomerRequest{
		Name:     "Jane",
		Email:    "dup@example.com",
		Password: "password123",
	}
	if _, err := svc.RegisterCustomer(context.Background(), req); err != nil {
		t.Fatalf("first register error = %v", err)
	}

	_, err := svc.RegisterCustomer(context.Background(), req)
	if err != users.ErrDuplicateEmail {
		t.Fatalf("error = %v, want %v", err, users.ErrDuplicateEmail)
	}
}

func TestLoginSuccess(t *testing.T) {
	store := newMockUserStore()
	svc := newTestService(t, store, newMockRefreshStore())

	_, err := svc.RegisterCustomer(context.Background(), RegisterCustomerRequest{
		Name:     "John",
		Email:    "john@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error = %v", err)
	}

	res, err := svc.LoginCustomer(context.Background(), LoginRequest{
		Email:    "john@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if res.Tokens.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	store := newMockUserStore()
	svc := newTestService(t, store, newMockRefreshStore())

	_, err := svc.RegisterCustomer(context.Background(), RegisterCustomerRequest{
		Name:     "John",
		Email:    "john2@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error = %v", err)
	}

	_, err = svc.LoginCustomer(context.Background(), LoginRequest{
		Email:    "john2@example.com",
		Password: "wrong-password",
	})
	if err != ErrInvalidCredentials {
		t.Fatalf("error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestRefreshInvalidToken(t *testing.T) {
	svc := newTestService(t, newMockUserStore(), newMockRefreshStore())

	_, err := svc.Refresh(context.Background(), "not-a-real-token")
	if err != ErrInvalidToken {
		t.Fatalf("error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestRefreshSuccess(t *testing.T) {
	store := newMockUserStore()
	refreshStore := newMockRefreshStore()
	svc := newTestService(t, store, refreshStore)

	registered, err := svc.RegisterCustomer(context.Background(), RegisterCustomerRequest{
		Name:     "Refresh User",
		Email:    "refresh@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error = %v", err)
	}

	refreshed, err := svc.Refresh(context.Background(), registered.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed.Tokens.AccessToken == "" {
		t.Fatal("expected new access token")
	}
}

func TestParseAccessTokenInvalid(t *testing.T) {
	cfg := testConfig()
	tm := security.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTAccessTTL)

	_, err := tm.ParseAccessToken("invalid.token.value")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

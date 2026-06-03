package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/config"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// UserStore loads and creates users.
type UserStore interface {
	Create(ctx context.Context, user *users.User) error
	GetByEmailAndRole(ctx context.Context, email, role string) (*users.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*users.User, error)
}

// RefreshStore manages refresh tokens.
type RefreshStore interface {
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetActiveByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error
}

// Service handles authentication business logic.
type Service struct {
	cfg           *config.Config
	users         UserStore
	registrar     AccountRegistrar
	refresh       RefreshStore
	tokens        *security.TokenManager
	refreshTTL    time.Duration
	refreshSecret []byte
	now           func() time.Time
}

// NewService creates an auth service.
func NewService(cfg *config.Config, userStore UserStore, registrar AccountRegistrar, refreshStore RefreshStore, tokens *security.TokenManager) *Service {
	return &Service{
		cfg:           cfg,
		users:         userStore,
		registrar:     registrar,
		refresh:       refreshStore,
		tokens:        tokens,
		refreshTTL:    cfg.JWTRefreshTTL,
		refreshSecret: []byte(cfg.JWTRefreshSecret),
		now:           func() time.Time { return time.Now().UTC() },
	}
}

// RegisterCustomer creates a customer account and issues tokens.
func (s *Service) RegisterCustomer(ctx context.Context, req RegisterCustomerRequest) (*AuthResponse, error) {
	return s.register(ctx, req.Name, req.Email, req.Phone, req.Password, users.RoleCustomer)
}

// RegisterEmployee creates an employee account and issues tokens.
func (s *Service) RegisterEmployee(ctx context.Context, req RegisterEmployeeRequest) (*AuthResponse, error) {
	return s.register(ctx, req.Name, req.Email, req.Phone, req.Password, users.RoleEmployee)
}

func (s *Service) register(ctx context.Context, name, email string, phone *string, password, role string) (*AuthResponse, error) {
	if err := validateRegister(name, email, password); err != nil {
		return nil, err
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := s.now()
	user := &users.User{
		ID:           uuid.New(),
		Name:         strings.TrimSpace(name),
		Email:        strings.ToLower(strings.TrimSpace(email)),
		Phone:        normalizePhone(phone),
		PasswordHash: hash,
		Role:         role,
		Status:       users.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	switch role {
	case users.RoleCustomer:
		err = s.registrar.RegisterCustomer(ctx, user)
	case users.RoleEmployee:
		err = s.registrar.RegisterEmployee(ctx, user)
	default:
		return nil, ErrValidation
	}
	if err != nil {
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

// LoginCustomer authenticates a customer account.
func (s *Service) LoginCustomer(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	return s.login(ctx, req, users.RoleCustomer)
}

// LoginEmployee authenticates an employee account.
func (s *Service) LoginEmployee(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	return s.login(ctx, req, users.RoleEmployee)
}

// LoginAdmin authenticates an admin account.
func (s *Service) LoginAdmin(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	return s.login(ctx, req, users.RoleAdmin)
}

func (s *Service) login(ctx context.Context, req LoginRequest, role string) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return nil, ErrValidation
	}

	user, err := s.users.GetByEmailAndRole(ctx, email, role)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := security.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status != users.StatusActive {
		return nil, ErrUserInactive
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

// Refresh exchanges a refresh token for a new token pair.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidToken
	}

	hash := hashRefreshToken(refreshToken, s.refreshSecret)
	stored, err := s.refresh.GetActiveByHash(ctx, hash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if s.now().After(stored.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	user, err := s.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	if user.Status != users.StatusActive {
		return nil, ErrUserInactive
	}

	if err := s.refresh.RevokeByHash(ctx, hash, s.now()); err != nil {
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

// Logout revokes the refresh token.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return ErrInvalidToken
	}

	hash := hashRefreshToken(refreshToken, s.refreshSecret)
	if err := s.refresh.RevokeByHash(ctx, hash, s.now()); err != nil {
		return ErrInvalidToken
	}
	return nil
}

// Me returns the authenticated user's profile.
func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (s *Service) issueTokens(ctx context.Context, user *users.User) (*TokenPair, error) {
	accessToken, accessExp, err := s.tokens.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	plainRefresh, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	now := s.now()
	refreshExp := now.Add(s.refreshTTL)
	hash := hashRefreshToken(plainRefresh, s.refreshSecret)

	rt := &RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: refreshExp,
		CreatedAt: now,
	}

	if err := s.refresh.CreateRefreshToken(ctx, rt); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     plainRefresh,
		AccessExpiresAt:  accessExp.Format(time.RFC3339),
		RefreshExpiresAt: refreshExp.Format(time.RFC3339),
	}, nil
}

func validateRegister(name, email, password string) error {
	if strings.TrimSpace(name) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(email) == "" || !strings.Contains(email, "@") {
		return ErrValidation
	}
	if len(password) < 8 {
		return ErrValidation
	}
	return nil
}

func normalizePhone(phone *string) *string {
	if phone == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*phone)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toUserResponse(user *users.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func hashRefreshToken(plain string, secret []byte) string {
	sum := sha256.Sum256(append([]byte(plain), secret...))
	return hex.EncodeToString(sum[:])
}

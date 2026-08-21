package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// EmailSender delivers transactional email messages.
type EmailSender interface {
	Enabled() bool
	Send(to, subject, body string) error
}

// UserLifecycleStore extends user loading with lifecycle updates.
type UserLifecycleStore interface {
	UserStore
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string, at time.Time) error
	MarkEmailVerified(ctx context.Context, id uuid.UUID, at time.Time) (*users.User, error)
}

// LifecycleStore manages password reset and email verification tokens.
type LifecycleStore interface {
	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetActivePasswordResetByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetUsed(ctx context.Context, tokenHash string, usedAt time.Time) error
	CreateEmailVerificationToken(ctx context.Context, token *EmailVerificationToken) error
	GetActiveEmailVerificationByHash(ctx context.Context, tokenHash string) (*EmailVerificationToken, error)
	MarkEmailVerificationUsed(ctx context.Context, tokenHash string, usedAt time.Time) error
}

const (
	resetTokenTTL        = 30 * time.Minute
	verificationTokenTTL = 24 * time.Hour
)

// ForgotPassword creates a reset token and sends email when SMTP is configured.
func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	if s.lifecycle == nil {
		return fmt.Errorf("lifecycle store not configured")
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	role := strings.TrimSpace(req.Role)
	if email == "" || role == "" {
		return ErrValidation
	}

	user, err := s.users.GetByEmailAndRole(ctx, email, role)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return nil
		}
		return err
	}

	plain, hash, err := s.createLifecycleToken()
	if err != nil {
		return err
	}

	now := s.now()
	if err := s.lifecycle.CreatePasswordResetToken(ctx, &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: now.Add(resetTokenTTL),
		CreatedAt: now,
	}); err != nil {
		return err
	}

	if s.mailer != nil && s.mailer.Enabled() {
		body := fmt.Sprintf("Use this token to reset your password: %s\nIt expires in 30 minutes.", plain)
		if err := s.mailer.Send(user.Email, "Password reset", body); err != nil {
			return fmt.Errorf("send password reset email: %w", err)
		}
	}
	return nil
}

// ResetPassword validates a reset token and updates the password.
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if s.lifecycle == nil {
		return fmt.Errorf("lifecycle store not configured")
	}
	token := strings.TrimSpace(req.Token)
	password := req.NewPassword
	if token == "" || len(password) < 8 {
		return ErrValidation
	}

	hash := hashLifecycleToken(token, s.lifecycleSecret)
	stored, err := s.lifecycle.GetActivePasswordResetByHash(ctx, hash)
	if err != nil {
		return ErrInvalidToken
	}
	if s.now().After(stored.ExpiresAt) {
		return ErrInvalidToken
	}

	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	now := s.now()
	lifecycleUser, ok := s.lifecycleUsers.(UserLifecycleStore)
	if !ok {
		return fmt.Errorf("user lifecycle store not configured")
	}
	if err := lifecycleUser.UpdatePassword(ctx, stored.UserID, passwordHash, now); err != nil {
		return err
	}
	return s.lifecycle.MarkPasswordResetUsed(ctx, hash, now)
}

// VerifyEmail marks a user's email as verified using a one-time token.
func (s *Service) VerifyEmail(ctx context.Context, req VerifyEmailRequest) error {
	if s.lifecycle == nil {
		return fmt.Errorf("lifecycle store not configured")
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		return ErrValidation
	}

	hash := hashLifecycleToken(token, s.lifecycleSecret)
	stored, err := s.lifecycle.GetActiveEmailVerificationByHash(ctx, hash)
	if err != nil {
		return ErrInvalidToken
	}
	if s.now().After(stored.ExpiresAt) {
		return ErrInvalidToken
	}

	now := s.now()
	lifecycleUser, ok := s.lifecycleUsers.(UserLifecycleStore)
	if !ok {
		return fmt.Errorf("user lifecycle store not configured")
	}
	if _, err := lifecycleUser.MarkEmailVerified(ctx, stored.UserID, now); err != nil {
		return err
	}
	return s.lifecycle.MarkEmailVerificationUsed(ctx, hash, now)
}

// ResendVerification creates a new email verification token for the authenticated user.
func (s *Service) ResendVerification(ctx context.Context, userID uuid.UUID) error {
	if s.lifecycle == nil {
		return fmt.Errorf("lifecycle store not configured")
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.EmailVerifiedAt != nil {
		return ErrValidation
	}

	plain, hash, err := s.createLifecycleToken()
	if err != nil {
		return err
	}

	now := s.now()
	if err := s.lifecycle.CreateEmailVerificationToken(ctx, &EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: now.Add(verificationTokenTTL),
		CreatedAt: now,
	}); err != nil {
		return err
	}

	if s.mailer != nil && s.mailer.Enabled() {
		body := fmt.Sprintf("Verify your email with this token: %s\nIt expires in 24 hours.", plain)
		if err := s.mailer.Send(user.Email, "Verify your email", body); err != nil {
			return fmt.Errorf("send verification email: %w", err)
		}
	}
	return nil
}

// ChangePassword updates the authenticated user's password.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	if req.CurrentPassword == "" || len(req.NewPassword) < 8 {
		return ErrValidation
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := security.CheckPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return ErrInvalidCredentials
	}

	passwordHash, err := security.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	lifecycleUser, ok := s.lifecycleUsers.(UserLifecycleStore)
	if !ok {
		return fmt.Errorf("user lifecycle store not configured")
	}
	return lifecycleUser.UpdatePassword(ctx, userID, passwordHash, s.now())
}

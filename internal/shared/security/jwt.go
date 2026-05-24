package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AccessClaims are stored in the JWT access token.
type AccessClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager signs and validates JWT access tokens.
type TokenManager struct {
	accessSecret []byte
	accessTTL    time.Duration
}

// NewTokenManager creates a JWT access token manager.
func NewTokenManager(accessSecret string, accessTTL time.Duration) *TokenManager {
	return &TokenManager{
		accessSecret: []byte(accessSecret),
		accessTTL:    accessTTL,
	}
}

// GenerateAccessToken issues a signed access token for the user.
func (m *TokenManager) GenerateAccessToken(userID uuid.UUID, role string) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(m.accessTTL)

	claims := AccessClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signed, expiresAt, nil
}

// ParseAccessToken validates the token and returns claims.
func (m *TokenManager) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.accessSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid access token claims")
	}

	return claims, nil
}

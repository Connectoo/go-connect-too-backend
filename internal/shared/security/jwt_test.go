package security

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateAndParseAccessToken(t *testing.T) {
	tm := NewTokenManager("test-access-secret-min-32-characters", 15*time.Minute)
	userID := uuid.New()

	token, _, err := tm.GenerateAccessToken(userID, "customer")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := tm.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}

	if claims.UserID != userID || claims.Role != "customer" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

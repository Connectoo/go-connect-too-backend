package security

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if err := CheckPassword(hash, "password123"); err != nil {
		t.Fatalf("CheckPassword() error = %v", err)
	}

	if err := CheckPassword(hash, "wrong"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}

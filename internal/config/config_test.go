package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/app?sslmode=disable")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("JWT_ACCESS_SECRET", "test-access-secret-min-32-characters")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-min-32-characters")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPPort != 9090 {
		t.Errorf("HTTPPort = %d, want 9090", cfg.HTTPPort)
	}
	if cfg.DatabaseURL == "" {
		t.Error("DatabaseURL should not be empty")
	}
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when DATABASE_URL is missing")
	}
}

func TestLoadMissingJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/app?sslmode=disable")
	os.Unsetenv("JWT_ACCESS_SECRET")
	os.Unsetenv("JWT_REFRESH_SECRET")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when JWT secrets are missing")
	}
}

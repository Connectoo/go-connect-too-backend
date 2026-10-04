package config

import (
	"os"
	"strings"
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

func TestLoadOTPSMSTemplateDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/app?sslmode=disable")
	t.Setenv("JWT_ACCESS_SECRET", "test-access-secret-min-32-characters")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-min-32-characters")
	os.Unsetenv("OTP_SMS_TEMPLATE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.OTPSMSTemplate == "" {
		t.Fatal("OTPSMSTemplate should default to a non-empty value")
	}
	if !strings.Contains(cfg.OTPSMSTemplate, "{code}") {
		t.Errorf("OTPSMSTemplate = %q, want it to contain the {code} token", cfg.OTPSMSTemplate)
	}
}

func TestTwilioEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{name: "all empty", cfg: Config{}, want: false},
		{
			name: "sid+token+from",
			cfg:  Config{TwilioAccountSID: "AC123", TwilioAuthToken: "tok", TwilioFromNumber: "+1555"},
			want: true,
		},
		{
			name: "sid+token+messaging service",
			cfg:  Config{TwilioAccountSID: "AC123", TwilioAuthToken: "tok", TwilioMessagingServiceSID: "MG123"},
			want: true,
		},
		{
			name: "sid+token but no from or messaging service",
			cfg:  Config{TwilioAccountSID: "AC123", TwilioAuthToken: "tok"},
			want: false,
		},
		{
			name: "missing token",
			cfg:  Config{TwilioAccountSID: "AC123", TwilioFromNumber: "+1555"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.TwilioEnabled(); got != tt.want {
				t.Errorf("TwilioEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailOTPEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{name: "smtp host set", cfg: Config{SMTPHost: "smtp.example.com"}, want: true},
		{name: "smtp host empty", cfg: Config{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.EmailOTPEnabled(); got != tt.want {
				t.Errorf("EmailOTPEnabled() = %v, want %v", got, tt.want)
			}
		})
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

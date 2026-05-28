package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	AppEnv      string
	HTTPPort    int
	DatabaseURL string
	LogLevel    string

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	RazorpayKeyID         string
	RazorpayKeySecret     string
	RazorpayWebhookSecret string
}

// Load reads configuration from the environment.
// A .env file in the working directory is loaded when present (local dev);
// variables already set in the process environment take precedence.
func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", getEnv("HTTP_PORT", "8080")))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	maxOpen, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdle, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}

	connLifetimeSec, err := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME_SEC", "300"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME_SEC: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		return nil, fmt.Errorf("JWT_REFRESH_SECRET is required")
	}

	accessTTLMin, err := strconv.Atoi(getEnv("JWT_ACCESS_TTL_MINUTES", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL_MINUTES: %w", err)
	}

	refreshTTLDays, err := strconv.Atoi(getEnv("JWT_REFRESH_TTL_DAYS", "7"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL_DAYS: %w", err)
	}

	return &Config{
		AppEnv:                getEnv("APP_ENV", "development"),
		HTTPPort:              port,
		DatabaseURL:           databaseURL,
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		DBMaxOpenConns:        maxOpen,
		DBMaxIdleConns:        maxIdle,
		DBConnMaxLifetime:     time.Duration(connLifetimeSec) * time.Second,
		JWTAccessSecret:       accessSecret,
		JWTRefreshSecret:      refreshSecret,
		JWTAccessTTL:          time.Duration(accessTTLMin) * time.Minute,
		JWTRefreshTTL:         time.Duration(refreshTTLDays) * 24 * time.Hour,
		RazorpayKeyID:         os.Getenv("RAZORPAY_KEY_ID"),
		RazorpayKeySecret:     os.Getenv("RAZORPAY_KEY_SECRET"),
		RazorpayWebhookSecret: os.Getenv("RAZORPAY_WEBHOOK_SECRET"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

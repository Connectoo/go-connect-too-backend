package app

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MustafaKheda/go-connect-too-backend/internal/config"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

type fakeDB struct {
	pingErr error
}

func (f *fakeDB) Ping(context.Context) error {
	return f.pingErr
}

func TestHealthCheckSuccess(t *testing.T) {
	srv := newTestServer(t, &fakeDB{pingErr: nil})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body response.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !body.Success {
		t.Fatalf("expected success response, got %+v", body)
	}
}

func TestSwaggerUIDocs(t *testing.T) {
	srv := newTestServer(t, &fakeDB{pingErr: nil})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/docs/", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
}

func TestOpenAPISpec(t *testing.T) {
	srv := newTestServer(t, &fakeDB{pingErr: nil})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/docs/openapi.yaml", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Header().Get("Content-Type") != "application/yaml" {
		t.Fatalf("unexpected Content-Type: %s", rec.Header().Get("Content-Type"))
	}
}

func TestHealthCheckDatabaseDown(t *testing.T) {
	srv := newTestServer(t, &fakeDB{pingErr: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body response.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Success || body.Error != "HEALTH_CHECK_FAILED" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func newTestServer(t *testing.T, db Pinger) *Server {
	t.Helper()

	cfg := &config.Config{
		AppEnv:   "test",
		HTTPPort: 8080,
		LogLevel: "error",
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewServer(cfg, log, db, nil)
}

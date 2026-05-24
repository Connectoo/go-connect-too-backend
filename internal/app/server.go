package app
import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/MustafaKheda/go-connect-too-backend/internal/config"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/auth"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

// Pinger checks database connectivity for health checks.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Server is the HTTP application server.
type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	db     Pinger
	router chi.Router
	http   *http.Server
}

// NewServer builds the HTTP server and route tree.
func NewServer(cfg *config.Config, log *slog.Logger, db Pinger, sqlDB *sql.DB) *Server {
	s := &Server{
		cfg: cfg,
		log: log,
		db:  db,
	}

	s.router = chi.NewRouter()
	s.router.Use(chimiddleware.RequestID)
	s.router.Use(chimiddleware.RealIP)
	s.router.Use(chimiddleware.Recoverer)
	s.router.Use(middleware.RequestLogger(log))

	tokenManager := security.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTAccessTTL)
	userRepo := users.NewRepository(sqlDB)
	authRepo := auth.NewRepository(sqlDB)
	authSvc := auth.NewService(cfg, userRepo, authRepo, tokenManager)
	authHandler := auth.NewHandler(authSvc)

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", s.healthCheck)
		auth.RegisterRoutes(r, authHandler, tokenManager)
		if cfg.AppEnv != "production" {
			registerDocsRoutes(r)
		}
	})

	s.http = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Run starts the HTTP server and blocks until the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("http server starting", slog.String("addr", s.http.Addr))
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		s.log.Info("http server shutting down")
		if err := s.http.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		s.log.Error("health check failed", slog.String("error", err.Error()))
		response.Error(w, http.StatusServiceUnavailable, "Service is unhealthy", sharederrors.CodeHealthCheckFailed)
		return
	}

	response.JSON(w, http.StatusOK, "Service is healthy", map[string]string{
		"status":   "ok",
		"database": "up",
	})
}

// Handler exposes the root HTTP handler for tests.
func (s *Server) Handler() http.Handler {
	return s.router
}

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
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/admin"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/analytics"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/auth"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/availability"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/badges"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/bookings"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/categories"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/chat"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/events"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/kyc"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/moderation"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/notifications"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/payments"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/public"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/ratings"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/reports"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/reviews"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/search"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/services"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/settings"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/storage"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/subscriptions"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/support"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/webhooks"
	ws "github.com/MustafaKheda/go-connect-too-backend/internal/modules/websocket"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/workers"
	platformemail "github.com/MustafaKheda/go-connect-too-backend/internal/platform/email"
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
	s.router.Use(middleware.CORS)
	s.router.Use(middleware.RequestLogger(log))

	tokenManager := security.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTAccessTTL)
	userRepo := users.NewRepository(sqlDB)
	customerRepo := customers.NewRepository(sqlDB)
	employeeRepo := employees.NewRepository(sqlDB)
	registrar := auth.NewRegistrar(sqlDB, userRepo, customerRepo, employeeRepo)
	authRepo := auth.NewRepository(sqlDB)
	emailSender := platformemail.NewSender(platformemail.Config{
		Host: cfg.SMTPHost,
		Port: cfg.SMTPPort,
		User: cfg.SMTPUser,
		Pass: cfg.SMTPPass,
		From: cfg.SMTPFrom,
	})
	authSvc := auth.NewService(cfg, userRepo, registrar, authRepo, tokenManager,
		auth.WithLifecycleStore(authRepo),
		auth.WithEmailSender(emailSender),
	)
	authHandler := auth.NewHandler(authSvc, log)
	userStatusUpdater := users.NewStatusUpdater(userRepo)
	badgeRepo := badges.NewRepository(sqlDB)
	badgeSvc := badges.NewService(badgeRepo)
	kycRepo := kyc.NewRepository(sqlDB)
	kycStatusAdapter := kyc.NewStatusAdapter(kycRepo)
	employeeSvc := employees.NewService(employeeRepo, userStatusUpdater).
		WithKYCChecker(kycStatusAdapter)
	analyticsRepo := analytics.NewRepository(sqlDB)
	analyticsSvc := analytics.NewService(analyticsRepo, employeeRepo)
	employeeSvc = employeeSvc.WithBadges(badgeSvc).WithProfileViews(analyticsSvc)
	employeeHandler := employees.NewHandler(employeeSvc, log)
	analyticsHandler := analytics.NewHandler(analyticsSvc, log)
	eventDispatcher := events.NewDispatcher()
	kycSvc := kyc.NewService(
		kyc.NewEmployeeRepositoryAdapter(employeeRepo),
		kycRepo,
		kyc.WithVerificationSync(kyc.NewEmployeeVerificationAdapter(employeeRepo)),
		kyc.WithEmployeeUserLookup(kyc.NewEmployeeUserAdapter(employeeRepo)),
		kyc.WithEventPublisher(eventDispatcher),
	)
	kycHandler := kyc.NewHandler(kycSvc, log)

	var storageHandler *storage.Handler
	if cfg.StorageEnabled() {
		s3Store, err := storage.NewS3Storage(context.Background(), storage.S3Config{
			Bucket:    cfg.S3Bucket,
			Region:    cfg.S3Region,
			AccessKey: cfg.S3AccessKey,
			SecretKey: cfg.S3SecretKey,
			Endpoint:  cfg.S3Endpoint,
			BaseURL:   cfg.S3BaseURL,
		})
		if err != nil {
			log.Error("s3 storage init failed", slog.String("error", err.Error()))
		} else {
			storageRepo := storage.NewRepository(sqlDB)
			storageSvc := storage.NewService(storageRepo, s3Store)
			storageHandler = storage.NewHandler(storageSvc, log)
			kycSvc = kyc.NewService(
				kyc.NewEmployeeRepositoryAdapter(employeeRepo),
				kycRepo,
				kyc.WithVerificationSync(kyc.NewEmployeeVerificationAdapter(employeeRepo)),
				kyc.WithEmployeeUserLookup(kyc.NewEmployeeUserAdapter(employeeRepo)),
				kyc.WithEventPublisher(eventDispatcher),
				kyc.WithFileResolver(storageSvc),
			)
			kycHandler = kyc.NewHandler(kycSvc, log)
			employeeSvc = employeeSvc.WithProfileFiles(storageSvc)
			employeeHandler = employees.NewHandler(employeeSvc, log)
		}
	}

	categoryRepo := categories.NewRepository(sqlDB)
	categorySvc := categories.NewService(categoryRepo)
	categoryHandler := categories.NewHandler(categorySvc, log)
	subscriptionRepo := subscriptions.NewRepository(sqlDB)
	razorpayGateway := payments.NewRazorpayGateway(cfg.RazorpayKeyID, cfg.RazorpayKeySecret, cfg.RazorpayWebhookSecret)
	paymentRepo := payments.NewRepository(sqlDB)
	paymentSvc := payments.NewService(employeeRepo, paymentRepo, razorpayGateway, razorpayGateway.KeyID())
	paymentHandler := payments.NewHandler(paymentSvc, log)
	subscriptionSvc := subscriptions.NewService(employeeRepo, subscriptionRepo, paymentSvc).
		WithPaymentVerifier(paymentSvc)
	subscriptionHandler := subscriptions.NewHandler(subscriptionSvc, log)
	webhookHandler := webhooks.NewHandler(paymentSvc, log)
	serviceRepo := services.NewRepository(sqlDB)
	serviceSvc := services.NewService(employeeRepo, serviceRepo, subscriptionRepo)
	serviceHandler := services.NewHandler(serviceSvc, log)
	availabilityRepo := availability.NewRepository(sqlDB)
	availabilitySvc := availability.NewService(employeeRepo, availabilityRepo)
	availabilityHandler := availability.NewHandler(availabilitySvc, log)
	userSvc := users.NewService(userRepo, userRepo)
	userHandler := users.NewHandler(userSvc, log)
	searchRepo := search.NewRepository(sqlDB)
	searchSvc := search.NewService(searchRepo)
	searchHandler := search.NewHandler(searchSvc, log)
	wsHub := ws.NewHub()
	bookingEmailProvider := notifications.NewSMTPProvider(notifications.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Password: cfg.SMTPPass,
		From:     cfg.SMTPFrom,
	})
	notificationRepo := notifications.NewRepository(sqlDB)
	notificationSvc := notifications.NewService(notificationRepo)
	notificationHandler := notifications.NewHandler(notificationSvc, log)
	pushProvider := notifications.PushProvider(notifications.NoopPushProvider{})
	if cfg.FCMEnabled() {
		pushProvider = notifications.NewFCMPushProvider(cfg.FCMProjectID, []byte(cfg.FCMCredentialsJSON), notificationRepo, log)
	}
	chatRepo := chat.NewRepository(sqlDB)
	chatSvc := chat.NewService(customerRepo, employeeRepo, chatRepo, eventDispatcher)
	chatHandler := chat.NewHandler(chatSvc, log)
	wsHandler := ws.NewHandler(wsHub, tokenManager, log)
	notificationWorker := workers.NewNotificationWorkerWithEmail(
		customerRepo,
		employeeRepo,
		workers.NewUserEmailAdapter(userRepo),
		notificationSvc,
		pushProvider,
		bookingEmailProvider,
		chatSvc,
		wsHub,
		log,
	)
	notificationWorker.Register(eventDispatcher)
	bookingPublisher := workers.NewBookingPublisher(eventDispatcher)
	bookingRepo := bookings.NewRepository(sqlDB)
	bookingSvc := bookings.NewService(customerRepo, employeeRepo, serviceRepo, bookingRepo, bookingPublisher)
	bookingHandler := bookings.NewHandler(bookingSvc, log)
	ratingRepo := ratings.NewRepository(sqlDB)
	ratingSvc := ratings.NewService(ratingRepo)
	reviewRepo := reviews.NewRepository(sqlDB)
	reviewSvc := reviews.NewService(customerRepo, employeeRepo, bookingRepo, reviewRepo, badgeSvc, ratings.NewRefresher(ratingSvc))
	reviewHandler := reviews.NewHandler(reviewSvc, log)
	reportRepo := reports.NewRepository(sqlDB)
	reportSvc := reports.NewService(bookingRepo, reportRepo)
	reportHandler := reports.NewHandler(reportSvc, log)
	moderationSvc := moderation.NewService(reviewSvc, reportSvc)
	moderationHandler := moderation.NewHandler(moderationSvc, log)
	adminRepo := admin.NewRepository(sqlDB)
	adminAuditRepo := admin.NewAuditRepository(sqlDB)
	adminSvc := admin.NewService(adminRepo, userRepo)
	adminHandler := admin.NewHandler(adminSvc, log)
	supportRepo := support.NewRepository(sqlDB)
	supportSvc := support.NewService(customerRepo, supportRepo)
	supportHandler := support.NewHandler(supportSvc, log)
	publicRepo := public.NewRepository(sqlDB)
	publicSvc := public.NewService(categoryRepo, publicRepo, serviceRepo, employeeRepo).WithProfileViews(analyticsSvc)
	publicHandler := public.NewHandler(publicSvc, log)
	settingsRepo := settings.NewRepository(sqlDB)
	settingsSvc := settings.NewService(settingsRepo)
	settingsHandler := settings.NewHandler(settingsSvc, log)

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", s.healthCheck)
		auth.RegisterRoutes(r, authHandler, tokenManager)
		public.RegisterRoutes(r, publicHandler)
		users.RegisterRoutes(r, userHandler, tokenManager)
		search.RegisterRoutes(r, searchHandler)
		employees.RegisterRoutes(r, employeeHandler, tokenManager)
		kyc.RegisterRoutes(r, kycHandler, tokenManager)
		if storageHandler != nil {
			storage.RegisterRoutes(r, storageHandler, tokenManager)
		}
		categories.RegisterRoutes(r, categoryHandler, tokenManager)
		subscriptions.RegisterRoutes(r, subscriptionHandler, tokenManager)
		payments.RegisterRoutes(r, paymentHandler, tokenManager, adminAuditRepo)
		webhooks.RegisterRoutes(r, webhookHandler)
		services.RegisterRoutes(r, serviceHandler, tokenManager)
		availability.RegisterRoutes(r, availabilityHandler, tokenManager)
		bookings.RegisterRoutes(r, bookingHandler, tokenManager, adminAuditRepo)
		reviews.RegisterRoutes(r, reviewHandler, tokenManager)
		notifications.RegisterRoutes(r, notificationHandler, tokenManager)
		chat.RegisterRoutes(r, chatHandler, tokenManager)
		ws.RegisterRoutes(r, wsHandler)
		admin.RegisterRoutes(r, adminHandler, tokenManager, adminAuditRepo)
		support.RegisterRoutes(r, supportHandler, tokenManager, adminAuditRepo)
		analytics.RegisterRoutes(r, analyticsHandler, tokenManager)
		settings.RegisterRoutes(r, settingsHandler, tokenManager, adminAuditRepo)
		moderation.RegisterRoutes(r, moderationHandler, tokenManager)
		reports.RegisterRoutes(r, reportHandler, tokenManager, adminAuditRepo)
		if cfg.AppEnv != "production" {
			registerDocsRoutes(r)
		}
	})

	s.http = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0,
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

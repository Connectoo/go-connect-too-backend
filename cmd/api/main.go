package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/MustafaKheda/go-connect-too-backend/internal/app"
	"github.com/MustafaKheda/go-connect-too-backend/internal/config"
	"github.com/MustafaKheda/go-connect-too-backend/internal/platform/database"
	"github.com/MustafaKheda/go-connect-too-backend/internal/platform/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)
	log.Info("starting api server", slog.String("env", cfg.AppEnv))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgres(ctx, cfg)
	if err != nil {
		log.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("failed to close database", slog.String("error", err.Error()))
		}
	}()

	server := app.NewServer(cfg, log, db, db.DB)
	if err := server.Run(ctx); err != nil {
		log.Error("server stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("server stopped")
}

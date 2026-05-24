package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MustafaKheda/go-connect-too-backend/internal/config"
)

// Postgres wraps a PostgreSQL connection pool.
type Postgres struct {
	DB *sql.DB
}

// NewPostgres opens a PostgreSQL connection pool using pgx via database/sql.
func NewPostgres(ctx context.Context, cfg *config.Config) (*Postgres, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Postgres{DB: db}, nil
}

// Ping checks database connectivity.
func (p *Postgres) Ping(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}

// Close closes the database connection pool.
func (p *Postgres) Close() error {
	return p.DB.Close()
}

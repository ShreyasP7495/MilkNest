package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/milknest/backend/pkg/config"
)

// New opens a *sql.DB pool against Postgres using the pgx stdlib driver.
// It pings once to fail fast on bad credentials.
func New(cfg config.DBConfig) (*sql.DB, error) {
	conn, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(30 * time.Minute)
	conn.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return conn, nil
}

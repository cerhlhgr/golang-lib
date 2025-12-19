package application

import (
	"context"
	"fmt"
	"golang-lib/application"
	"golang-lib/config"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(ctx context.Context, a *application.Application) error {
	pool, err := newPostgresPool(ctx)
	if err != nil {
		log.Fatal(err)
	}

	a.AddCloser(
		func() error {
			fmt.Println("closing database connection")
			pool.Close()
			return nil
		},
	)

	a.PGConn = pool

	return nil
}

func newPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := config.MustString("POSTGRES_DSN")
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("DATABASE_URL is empty")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_DSN: %w", err)
	}

	cfg.MaxConns = int32(config.GetInt("PG_MAX_CONNS", 10))
	cfg.MinConns = int32(config.GetInt("PG_MIN_CONNS", 0))
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}

	// проверка, что реально подключились
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg ping: %w", err)
	}

	return pool, nil
}

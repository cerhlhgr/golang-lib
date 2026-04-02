package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cerhlhgr/golang-lib/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(ctx context.Context) (*Instance, error) {
	dsn := config.MustString("POSTGRES_DSN")
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("POSTGRES_DSN is empty")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse POSTGRES_DSN: %w", err)
	}

	cfg.MaxConns = int32(config.GetInt("PG_MAX_CONNS", 10))
	cfg.MinConns = int32(config.GetInt("PG_MIN_CONNS", 0))
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg ping: %w", err)
	}

	return NewInstance(pool), nil
}

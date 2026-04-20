package application

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func initDB(ctx context.Context, a *Application) error {
	if a.PGConn != nil {
		return nil
	}

	cfg := a.postgresConfig
	if cfg == nil {
		parsed := PostgresConfigFromEnv()
		cfg = &parsed
	}

	pool, err := newPostgresPool(ctx, *cfg)
	if err != nil {
		return err
	}

	a.AddCloserContext("postgres", func(context.Context) error {
		pool.Close()
		return nil
	})

	a.PGConn = pool
	a.RegisterReadinessCheck("postgres", func(ctx context.Context) error {
		return pool.Ping(ctx)
	})
	a.RegisterLivenessCheck("postgres", func(ctx context.Context) error {
		return pool.Ping(ctx)
	})

	return nil
}

func newPostgresPool(ctx context.Context, pgCfg PostgresConfig) (*pgxpool.Pool, error) {
	pgCfg.setDefaults()
	if err := pgCfg.Validate(); err != nil {
		return nil, err
	}

	cfg, err := pgxpool.ParseConfig(pgCfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}

	cfg.MaxConns = pgCfg.MaxConns
	cfg.MinConns = pgCfg.MinConns
	cfg.MaxConnLifetime = pgCfg.MaxConnLifetime
	cfg.MaxConnIdleTime = pgCfg.MaxConnIdleTime
	cfg.HealthCheckPeriod = pgCfg.HealthCheck

	connectCtx, cancel := context.WithTimeout(ctx, pgCfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}

	if pgCfg.PingOnStart {
		pingCtx, pingCancel := context.WithTimeout(ctx, pgCfg.PingTimeout)
		defer pingCancel()

		if err := pool.Ping(pingCtx); err != nil {
			pool.Close()
			return nil, fmt.Errorf("pg ping: %w", err)
		}
	}

	return pool, nil
}

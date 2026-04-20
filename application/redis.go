package application

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func initRedis(ctx context.Context, a *Application) error {
	if a.RedisClient != nil {
		return nil
	}

	cfg := a.redisConfig
	if cfg == nil {
		parsed := RedisConfigFromEnv()
		cfg = &parsed
	}

	client, err := newRedisClient(ctx, *cfg)
	if err != nil {
		return err
	}

	a.RedisClient = client
	a.AddCloserContext("redis", func(context.Context) error {
		return client.Close()
	})
	a.RegisterReadinessCheck("redis", func(ctx context.Context) error {
		return client.Ping(ctx).Err()
	})
	a.RegisterLivenessCheck("redis", func(ctx context.Context) error {
		return client.Ping(ctx).Err()
	})

	return nil
}

func newRedisClient(ctx context.Context, cfg RedisConfig) (*redis.Client, error) {
	cfg.setDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		PoolTimeout:  cfg.PoolTimeout,
	})

	if cfg.PingOnStart {
		pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
		defer cancel()

		if err := client.Ping(pingCtx).Err(); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("redis ping: %w", err)
		}
	}

	return client, nil
}

package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/cerhlhgr/golang-lib/config"
	redisClient "github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context) (*Instance, error) {
	addr := config.MustString("REDIS_ADDR")
	if strings.TrimSpace(addr) == "" {
		return nil, fmt.Errorf("REDIS_ADDR is empty")
	}

	client := redisClient.NewClient(&redisClient.Options{
		Addr:     addr,
		Username: config.GetString("REDIS_USERNAME", ""),
		Password: config.GetString("REDIS_PASSWORD", ""),
		DB:       config.GetInt("REDIS_DB", 0),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return NewInstance(client), nil
}

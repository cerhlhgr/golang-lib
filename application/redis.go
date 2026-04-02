package application

import (
	"context"

	"github.com/cerhlhgr/golang-lib/redis"
)

func initRedis(ctx context.Context, a *Application) error {
	instance, err := redis.NewRedis(ctx)
	if err != nil {
		return err
	}

	a.Redis = instance
	a.AddCloser(func() error {
		return instance.Close()
	})

	return nil
}

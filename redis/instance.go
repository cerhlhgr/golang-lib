package redis

import "github.com/redis/go-redis/v9"

type Instance struct {
	Client *redis.Client
}

func NewInstance(client *redis.Client) *Instance {
	return &Instance{Client: client}
}

func (i *Instance) Close() error {
	if i == nil || i.Client == nil {
		return nil
	}

	return i.Client.Close()
}

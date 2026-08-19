package health

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisChecker struct {
	client *redis.Client
}

func NewRedisChecker(client *redis.Client) *RedisChecker {
	if client == nil {
		panic("redis client is nil")
	}

	return &RedisChecker{
		client: client,
	}
}

func (c *RedisChecker) Check(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check: %w", err)
	}

	return nil
}

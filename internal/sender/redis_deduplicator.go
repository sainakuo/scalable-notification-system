package sender

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDeduplicator struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisDeduplicator(
	client *redis.Client,
	ttl time.Duration,
) *RedisDeduplicator {
	if client == nil {
		panic("redis client is nil")
	}

	if ttl <= 0 {
		panic("deduplication ttl must be greater than zero")
	}

	return &RedisDeduplicator{
		client: client,
		ttl:    ttl,
	}
}

func (d *RedisDeduplicator) IsProcessed(
	ctx context.Context,
	taskID int,
) (bool, error) {
	key := processedTaskKey(taskID)

	exists, err := d.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check processed task: %w", err)
	}

	return exists > 0, nil
}

func (d *RedisDeduplicator) MarkProcessed(
	ctx context.Context,
	taskID int,
) error {
	key := processedTaskKey(taskID)

	if err := d.client.Set(
		ctx,
		key,
		"1",
		d.ttl,
	).Err(); err != nil {
		return fmt.Errorf("mark task processed: %w", err)
	}

	return nil
}

func processedTaskKey(taskID int) string {
	return "notification:processed:" + strconv.Itoa(taskID)
}

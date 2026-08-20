package queue

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	Client *redis.Client
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{Client: client}
}

func (q *RedisQueue) PushTask(ctx context.Context, taskID int) error {
	return q.Client.LPush(ctx, "tasks_queue", strconv.Itoa(taskID)).Err()
}

func (q *RedisQueue) PopTask(ctx context.Context) (int, error) {
	for {
		result, err := q.Client.BRPop(
			ctx,
			5*time.Second,
			"tasks_queue",
		).Result()

		if err == redis.Nil {
			continue
		}

		if err != nil {
			return 0, fmt.Errorf("pop task from redis: %w", err)
		}

		taskID, err := strconv.Atoi(result[1])
		if err != nil {
			return 0, fmt.Errorf("parse task id: %w", err)
		}

		return taskID, nil
	}
}

func (q *RedisQueue) Size(ctx context.Context) (int, error) {
	size, err := q.Client.LLen(ctx, "tasks_queue").Result()
	if err != nil {
		return 0, fmt.Errorf("get queue size: %w", err)
	}

	return int(size), nil
}

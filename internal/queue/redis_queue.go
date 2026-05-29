package queue

import (
	"context"
	"strconv"

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

package queue

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sainakuo/scalable-notification-system/internal/config"
)

func TestRedisQueue_PushAndPop_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	cfg := config.LoadConfig()

	redisClient := config.ConnectRedis(cfg.RedisAddr)
	defer redisClient.Close()

	queue := NewRedisQueue(redisClient)

	// Убираем возможные данные от предыдущих запусков теста.
	if err := redisClient.Del(ctx, "tasks_queue").Err(); err != nil {
		t.Fatalf("clear queue: %v", err)
	}

	t.Cleanup(func() {
		_ = redisClient.Del(
			context.Background(),
			"tasks_queue",
		).Err()
	})

	expectedTaskID := 42

	if err := queue.PushTask(ctx, expectedTaskID); err != nil {
		t.Fatalf("push task: %v", err)
	}

	size, err := queue.Size(ctx)
	if err != nil {
		t.Fatalf("get queue size: %v", err)
	}

	if size != 1 {
		t.Fatalf(
			"expected queue size 1, got %d",
			size,
		)
	}

	taskID, err := queue.PopTask(ctx)
	if err != nil {
		t.Fatalf("pop task: %v", err)
	}

	if taskID != expectedTaskID {
		t.Fatalf(
			"expected task ID %d, got %d",
			expectedTaskID,
			taskID,
		)
	}
}
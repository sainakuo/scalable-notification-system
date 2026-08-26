package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/model"
	"github.com/sainakuo/scalable-notification-system/internal/queue"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
)

func TestTaskService_CreateTask_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	cfg := config.LoadConfig()

	db, err := config.ConnectPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	redisClient := config.ConnectRedis(cfg.RedisAddr)
	defer redisClient.Close()

	taskRepo := repository.NewTaskRepository(db)
	taskQueue := queue.NewRedisQueue(redisClient)
	taskService := NewTaskService(taskRepo, taskQueue)

	if err := redisClient.Del(ctx, "tasks_queue").Err(); err != nil {
		t.Fatalf("clear queue: %v", err)
	}

	var createdTaskID int

	t.Cleanup(func() {
		_ = redisClient.Del(
			context.Background(),
			"tasks_queue",
		).Err()

		if createdTaskID != 0 {
			_ = taskRepo.DeleteTask(
				context.Background(),
				createdTaskID,
			)
		}
	})

	createdTask, err := taskService.CreateTask(
		ctx,
		model.Task{
			UserID:  999,
			Type:    "email",
			Payload: "service integration test",
		},
	)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	createdTaskID = createdTask.ID

	// Проверяем PostgreSQL
	storedTask, err := taskRepo.GetTaskByID(ctx, createdTask.ID)
	if err != nil {
		t.Fatalf("get created task: %v", err)
	}

	if storedTask.ID != createdTask.ID {
		t.Fatalf(
			"expected stored task ID %d, got %d",
			createdTask.ID,
			storedTask.ID,
		)
	}

	if storedTask.Status != "pending" {
		t.Fatalf(
			"expected status pending, got %s",
			storedTask.Status,
		)
	}

	// Проверяем Redis
	queuedTaskID, err := taskQueue.PopTask(ctx)
	if err != nil {
		t.Fatalf("pop task: %v", err)
	}

	if queuedTaskID != createdTask.ID {
		t.Fatalf(
			"expected queued task ID %d, got %d",
			createdTask.ID,
			queuedTaskID,
		)
	}
}

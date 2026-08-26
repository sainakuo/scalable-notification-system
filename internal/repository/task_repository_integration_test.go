package repository

import (
	"context"
	"os"
	"testing"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/model"
)

func TestTaskRepository_CreateAndGetTask_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	cfg := config.LoadConfig()

	db, err := config.ConnectPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	repo := NewTaskRepository(db)

	created, err := repo.CreateTask(
		ctx,
		model.Task{
			UserID:  999,
			Type:    "email",
			Payload: "integration test",
			Status:  "pending",
		},
	)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	t.Cleanup(func() {
		_ = repo.DeleteTask(context.Background(), created.ID)
	})

	got, err := repo.GetTaskByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf(
			"expected task ID %d, got %d",
			created.ID,
			got.ID,
		)
	}

	if got.Payload != "integration test" {
		t.Fatalf(
			"expected payload integration test, got %s",
			got.Payload,
		)
	}
}
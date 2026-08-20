package worker

import (
	"context"

	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type TaskRepository interface {
	GetTaskByID(ctx context.Context, id int) (model.Task, error)
	UpdateStatus(ctx context.Context, taskID int, status string) error
	IncrementRetryCount(ctx context.Context, taskID int) error
}

type TaskQueue interface {
	PopTask(ctx context.Context) (int, error)
	PushTask(ctx context.Context, taskID int) error
	Size(ctx context.Context) (int, error)
}

type NotificationSender interface {
	Send(ctx context.Context, task model.Task) error
}

type Metrics interface {
	TaskProcessed()
	TaskFailed()
	TaskRetried()
	ObserveTaskDuration(seconds float64)
	SetQueueSize(size int)
}

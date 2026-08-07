package service

import (
	"context"

	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task model.Task) (model.Task, error)
	GetTaskByID(ctx context.Context, id int) (model.Task, error)
	GetAllTasks(ctx context.Context) ([]model.Task, error)
	DeleteTask(ctx context.Context, id int) error
}

type TaskQueue interface {
	PushTask(ctx context.Context, taskID int) error
}

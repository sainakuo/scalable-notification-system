package service

import (
	"context"

	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type TaskRepository interface {
	CreateTask(task model.Task) (model.Task, error)
	GetTaskByID(id int) (model.Task, error)
	GetAllTasks() ([]model.Task, error)
}

type TaskQueue interface {
	PushTask(ctx context.Context, taskID int) error
}

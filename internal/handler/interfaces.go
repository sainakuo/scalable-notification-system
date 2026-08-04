package handler

import (
	"context"

	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type TaskService interface {
	CreateTask(
		ctx context.Context,
		task model.Task,
	) (model.Task, error)

	GetTaskByID(
		ctx context.Context,
		id int,
	) (model.Task, error)

	GetAllTasks(ctx context.Context) ([]model.Task, error)
}

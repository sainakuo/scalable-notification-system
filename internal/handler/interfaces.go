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
		id int,
	) (model.Task, error)

	GetAllTasks() ([]model.Task, error)
}

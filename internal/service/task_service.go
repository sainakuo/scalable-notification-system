package service

import (
	"context"
	"fmt"

	"github.com/sainakuo/scalable-notification-system/internal/model"
	"github.com/sainakuo/scalable-notification-system/internal/queue"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
)

type TaskService struct {
	Repo  *repository.TaskRepository
	Queue *queue.RedisQueue
}

func NewTaskService(repo *repository.TaskRepository, taskQueue *queue.RedisQueue) *TaskService {
	return &TaskService{
		Repo:  repo,
		Queue: taskQueue,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, task model.Task) (model.Task, error) {

	task.Status = "pending"

	createdTask, err := s.Repo.CreateTask(task)

	if err != nil {
		return model.Task{}, fmt.Errorf("create task in repository: %w", err)
	}

	if err := s.Queue.PushTask(ctx, createdTask.ID); err != nil {
		return model.Task{}, fmt.Errorf("push task to queue: %w", err)
	}

	return createdTask, nil
}

func (s *TaskService) GetTaskByID(id int) (model.Task, error) {
	task, err := s.Repo.GetTaskByID(id)
	if err != nil {
		return model.Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return task, nil
}

func (s *TaskService) GetAllTasks() ([]model.Task, error) {
	tasks, err := s.Repo.GetAllTasks()
	if err != nil {
		return nil, fmt.Errorf("get all tasks: %w", err)
	}

	return tasks, nil
}

package service

import (
	"context"
	"fmt"

	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type TaskService struct {
	repo  TaskRepository
	queue TaskQueue
}

func NewTaskService(repo TaskRepository, taskQueue TaskQueue) *TaskService {

	if repo == nil {
		panic("task repository is nil")
	}

	if taskQueue == nil {
		panic("task queue is nil")
	}
	return &TaskService{
		repo:  repo,
		queue: taskQueue,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, task model.Task) (model.Task, error) {
	task.Status = "pending"

	createdTask, err := s.repo.CreateTask(task)

	if err != nil {
		return model.Task{}, fmt.Errorf("create task in repository: %w", err)
	}

	if err := s.queue.PushTask(ctx, createdTask.ID); err != nil {
		return model.Task{}, fmt.Errorf("push task to queue: %w", err)
	}

	return createdTask, nil
}

func (s *TaskService) GetTaskByID(id int) (model.Task, error) {
	task, err := s.repo.GetTaskByID(id)
	if err != nil {
		return model.Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return task, nil
}

func (s *TaskService) GetAllTasks() ([]model.Task, error) {
	tasks, err := s.repo.GetAllTasks()
	if err != nil {
		return nil, fmt.Errorf("get all tasks: %w", err)
	}

	return tasks, nil
}

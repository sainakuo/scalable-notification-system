package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type mockTaskRepository struct {
	createTaskFunc func(
		ctx context.Context,
		task model.Task,
	) (model.Task, error)

	getTaskByIDFunc func(
		ctx context.Context,
		id int,
	) (model.Task, error)

	getAllTasksFunc func(
		ctx context.Context,
	) ([]model.Task, error)

	deleteTaskFunc func(
		ctx context.Context,
		id int,
	) error
}

func (m *mockTaskRepository) CreateTask(
	ctx context.Context,
	task model.Task,
) (model.Task, error) {
	return m.createTaskFunc(ctx, task)
}

func (m *mockTaskRepository) GetTaskByID(
	ctx context.Context,
	id int,
) (model.Task, error) {

	if m.getTaskByIDFunc != nil {
		return m.getTaskByIDFunc(ctx, id)
	}

	return model.Task{}, nil
}

func (m *mockTaskRepository) GetAllTasks(
	ctx context.Context,
) ([]model.Task, error) {
	if m.getAllTasksFunc != nil {
		return m.getAllTasksFunc(ctx)
	}

	return nil, nil
}

func (m *mockTaskRepository) DeleteTask(
	ctx context.Context,
	id int,
) error {
	if m.deleteTaskFunc != nil {
		return m.deleteTaskFunc(ctx, id)
	}

	return nil
}

type mockTaskQueue struct {
	pushTaskFunc func(
		ctx context.Context,
		taskID int,
	) error
}

func (m *mockTaskQueue) PushTask(
	ctx context.Context,
	taskID int,
) error {
	return m.pushTaskFunc(ctx, taskID)
}

func TestTaskService_CreateTask_Success(t *testing.T) {
	ctx := context.Background()

	input := model.Task{
		UserID:  10,
		Type:    "email",
		Payload: "hello",
	}

	created := model.Task{
		ID:      42,
		UserID:  10,
		Type:    "email",
		Payload: "hello",
		Status:  "pending",
	}

	repo := &mockTaskRepository{
		createTaskFunc: func(
			ctx context.Context,
			task model.Task,
		) (model.Task, error) {

			if task.Status != "pending" {
				t.Fatalf(
					"expected status pending, got %s",
					task.Status,
				)
			}

			return created, nil
		},
	}

	var pushedTaskID int

	queue := &mockTaskQueue{
		pushTaskFunc: func(
			ctx context.Context,
			taskID int,
		) error {
			pushedTaskID = taskID
			return nil
		},
	}

	service := NewTaskService(repo, queue)

	result, err := service.CreateTask(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != created.ID {
		t.Errorf(
			"expected task ID %d, got %d",
			created.ID,
			result.ID,
		)
	}

	if pushedTaskID != created.ID {
		t.Errorf(
			"expected pushed task ID %d, got %d",
			created.ID,
			pushedTaskID,
		)
	}
}

func TestTaskService_CreateTask_RepositoryError(t *testing.T) {
	ctx := context.Background()

	expectedErr := errors.New("database error")

	repo := &mockTaskRepository{
		createTaskFunc: func(
			ctx context.Context,
			task model.Task,
		) (model.Task, error) {
			return model.Task{}, expectedErr
		},
	}

	queueCalled := false

	queue := &mockTaskQueue{
		pushTaskFunc: func(
			ctx context.Context,
			taskID int,
		) error {
			queueCalled = true
			return nil
		},
	}

	service := NewTaskService(repo, queue)

	_, err := service.CreateTask(
		ctx,
		model.Task{
			UserID:  10,
			Type:    "email",
			Payload: "hello",
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}

	if queueCalled {
		t.Fatal("queue should not be called when repository fails")
	}
}

func TestTaskService_CreateTask_QueueError_RollsBackTask(t *testing.T) {
	ctx := context.Background()

	queueErr := errors.New("redis error")

	createdTask := model.Task{
		ID:      42,
		UserID:  10,
		Type:    "email",
		Payload: "hello",
		Status:  "pending",
	}

	deletedTaskID := 0

	repo := &mockTaskRepository{
		createTaskFunc: func(
			ctx context.Context,
			task model.Task,
		) (model.Task, error) {
			return createdTask, nil
		},

		deleteTaskFunc: func(
			ctx context.Context,
			id int,
		) error {
			deletedTaskID = id
			return nil
		},
	}

	queue := &mockTaskQueue{
		pushTaskFunc: func(
			ctx context.Context,
			taskID int,
		) error {
			return queueErr
		},
	}

	service := NewTaskService(repo, queue)

	_, err := service.CreateTask(
		ctx,
		model.Task{
			UserID:  10,
			Type:    "email",
			Payload: "hello",
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, queueErr) {
		t.Fatalf(
			"expected queue error, got %v",
			err,
		)
	}

	if deletedTaskID != createdTask.ID {
		t.Fatalf(
			"expected task %d to be deleted, got %d",
			createdTask.ID,
			deletedTaskID,
		)
	}
}

func TestTaskService_CreateTask_QueueError_RollbackError(t *testing.T) {
	ctx := context.Background()

	queueErr := errors.New("redis error")
	rollbackErr := errors.New("delete task error")

	createdTask := model.Task{
		ID:      42,
		UserID:  10,
		Type:    "email",
		Payload: "hello",
		Status:  "pending",
	}

	repo := &mockTaskRepository{
		createTaskFunc: func(
			ctx context.Context,
			task model.Task,
		) (model.Task, error) {
			return createdTask, nil
		},

		deleteTaskFunc: func(
			ctx context.Context,
			id int,
		) error {
			return rollbackErr
		},
	}

	queue := &mockTaskQueue{
		pushTaskFunc: func(
			ctx context.Context,
			taskID int,
		) error {
			return queueErr
		},
	}

	service := NewTaskService(repo, queue)

	_, err := service.CreateTask(
		ctx,
		model.Task{
			UserID:  10,
			Type:    "email",
			Payload: "hello",
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, queueErr) {
		t.Fatalf("expected queue error, got %v", err)
	}

	if !errors.Is(err, rollbackErr) {
		t.Fatalf("expected rollback error, got %v", err)
	}
}

func TestTaskService_GetTaskByID_Success(t *testing.T) {
	ctx := context.Background()

	expectedTask := model.Task{
		ID:      42,
		UserID:  10,
		Type:    "email",
		Payload: "hello",
		Status:  "done",
	}

	repo := &mockTaskRepository{
		getTaskByIDFunc: func(
			ctx context.Context,
			id int,
		) (model.Task, error) {
			if id != 42 {
				t.Fatalf("expected id 42, got %d", id)
			}

			return expectedTask, nil
		},
	}

	queue := &mockTaskQueue{
		pushTaskFunc: func(
			ctx context.Context,
			taskID int,
		) error {
			return nil
		},
	}

	service := NewTaskService(repo, queue)

	result, err := service.GetTaskByID(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != expectedTask.ID {
		t.Fatalf(
			"expected task ID %d, got %d",
			expectedTask.ID,
			result.ID,
		)
	}
}

func TestTaskService_GetTaskByID_Error(t *testing.T) {
	ctx := context.Background()

	repo := &mockTaskRepository{
		getTaskByIDFunc: func(
			ctx context.Context,
			id int,
		) (model.Task, error) {
			return model.Task{}, model.ErrTaskNotFound
		},
	}

	queue := &mockTaskQueue{
		pushTaskFunc: func(
			ctx context.Context,
			taskID int,
		) error {
			return nil
		},
	}

	service := NewTaskService(repo, queue)

	_, err := service.GetTaskByID(ctx, 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf(
			"expected ErrTaskNotFound, got %v",
			err,
		)
	}
}

func TestTaskService_GetAllTasks_Success(t *testing.T) {
	ctx := context.Background()

	expectedTasks := []model.Task{
		{
			ID:      1,
			UserID:  10,
			Type:    "email",
			Payload: "first",
			Status:  "done",
		},
		{
			ID:      2,
			UserID:  20,
			Type:    "sms",
			Payload: "second",
			Status:  "pending",
		},
	}

	repo := &mockTaskRepository{
		getAllTasksFunc: func(
			ctx context.Context,
		) ([]model.Task, error) {
			return expectedTasks, nil
		},
	}

	queue := &mockTaskQueue{
		pushTaskFunc: func(
			ctx context.Context,
			taskID int,
		) error {
			return nil
		},
	}

	service := NewTaskService(repo, queue)

	result, err := service.GetAllTasks(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(expectedTasks) {
		t.Fatalf(
			"expected %d tasks, got %d",
			len(expectedTasks),
			len(result),
		)
	}

	if result[0].ID != expectedTasks[0].ID {
		t.Fatalf(
			"expected first task ID %d, got %d",
			expectedTasks[0].ID,
			result[0].ID,
		)
	}
}

func TestTaskService_GetAllTasks_Error(t *testing.T) {
	ctx := context.Background()

	expectedErr := errors.New("database error")

	repo := &mockTaskRepository{
		getAllTasksFunc: func(
			ctx context.Context,
		) ([]model.Task, error) {
			return nil, expectedErr
		},
	}

	queue := &mockTaskQueue{
		pushTaskFunc: func(
			ctx context.Context,
			taskID int,
		) error {
			return nil
		},
	}

	service := NewTaskService(repo, queue)

	_, err := service.GetAllTasks(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

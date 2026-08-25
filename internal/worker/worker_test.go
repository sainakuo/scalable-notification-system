package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type mockTaskRepository struct {
	task            model.Task
	updatedStatuses []string
	retryIncrements int
	updateStatusErr error
}

func (m *mockTaskRepository) GetTaskByID(
	ctx context.Context,
	id int,
) (model.Task, error) {
	return m.task, nil
}

func (m *mockTaskRepository) UpdateStatus(
	ctx context.Context,
	taskID int,
	status string,
) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	m.updatedStatuses = append(m.updatedStatuses, status)
	return nil
}

func (m *mockTaskRepository) IncrementRetryCount(
	ctx context.Context,
	taskID int,
) error {
	m.retryIncrements++
	return nil
}

type mockQueue struct {
	pushedTaskIDs []int
	pushErr       error
}

func (m *mockQueue) PopTask(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockQueue) PushTask(ctx context.Context, taskID int) error {
	if m.pushErr != nil {
		return m.pushErr
	}
	m.pushedTaskIDs = append(m.pushedTaskIDs, taskID)
	return nil
}

func (m *mockQueue) Size(ctx context.Context) (int, error) {
	return 0, nil
}

type mockMetrics struct {
	processed int
	failed    int
	retried   int
}

func (m *mockMetrics) TaskProcessed() {
	m.processed++
}

func (m *mockMetrics) TaskFailed() {
	m.failed++
}

func (m *mockMetrics) TaskRetried() {
	m.retried++
}

func (m *mockMetrics) ObserveTaskDuration(seconds float64) {}

func (m *mockMetrics) SetQueueSize(size int) {}

func TestWorker_ProcessTask_Success(t *testing.T) {
	repo := &mockTaskRepository{
		task: model.Task{
			ID:         42,
			UserID:     10,
			Type:       "email",
			Payload:    "hello",
			Status:     "pending",
			RetryCount: 0,
		},
	}

	queue := &mockQueue{}
	sender := &mockSender{
		shouldFail: false,
	}

	metrics := &mockMetrics{}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	retryStrategy := NewRetryStrategy(3)

	w := New(
		repo,
		queue,
		sender,
		retryStrategy,
		logger,
		metrics,
		5,
		100,
	)

	processed, err := w.processTask(
		context.Background(),
		42,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !processed {
		t.Fatal("expected task to be processed")
	}

	if len(repo.updatedStatuses) != 2 {
		t.Fatalf(
			"expected 2 status updates, got %d",
			len(repo.updatedStatuses),
		)
	}

	if repo.updatedStatuses[0] != "processing" {
		t.Errorf(
			"expected first status processing, got %s",
			repo.updatedStatuses[0],
		)
	}

	if repo.updatedStatuses[1] != "done" {
		t.Errorf(
			"expected second status done, got %s",
			repo.updatedStatuses[1],
		)
	}

	if sender.callCount != 1 {
		t.Fatalf(
			"expected sender to be called once, got %d",
			sender.callCount,
		)
	}
}

func TestWorker_ProcessTask_SenderError(t *testing.T) {
	repo := &mockTaskRepository{
		task: model.Task{
			ID:         42,
			UserID:     10,
			Type:       "email",
			Payload:    "hello",
			Status:     "pending",
			RetryCount: 0,
		},
	}

	queue := &mockQueue{}

	sender := &mockSender{
		shouldFail: true,
	}

	metrics := &mockMetrics{}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	retryStrategy := NewRetryStrategy(3)

	w := New(
		repo,
		queue,
		sender,
		retryStrategy,
		logger,
		metrics,
		5,
		100,
	)

	processed, err := w.processTask(
		context.Background(),
		42,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !processed {
		t.Fatal("expected task processing to have started")
	}

	if len(repo.updatedStatuses) != 1 {
		t.Fatalf(
			"expected 1 status update, got %d",
			len(repo.updatedStatuses),
		)
	}

	if repo.updatedStatuses[0] != "processing" {
		t.Fatalf(
			"expected status processing, got %s",
			repo.updatedStatuses[0],
		)
	}
}

func TestWorker_ProcessTask_AlreadyDone(t *testing.T) {
	repo := &mockTaskRepository{
		task: model.Task{
			ID:         42,
			UserID:     10,
			Type:       "email",
			Payload:    "hello",
			Status:     "done",
			RetryCount: 0,
		},
	}

	queue := &mockQueue{}

	sender := &mockSender{
		shouldFail: false,
	}

	metrics := &mockMetrics{}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	retryStrategy := NewRetryStrategy(3)

	w := New(
		repo,
		queue,
		sender,
		retryStrategy,
		logger,
		metrics,
		5,
		100,
	)

	processed, err := w.processTask(
		context.Background(),
		42,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if processed {
		t.Fatal("expected completed task to be skipped")
	}

	if len(repo.updatedStatuses) != 0 {
		t.Fatalf(
			"expected no status updates, got %d",
			len(repo.updatedStatuses),
		)
	}

	if sender.callCount != 0 {
		t.Fatalf(
			"expected sender not to be called, got %d calls",
			sender.callCount,
		)
	}
}

func TestWorker_HandleRetry_Success(t *testing.T) {
	repo := &mockTaskRepository{
		task: model.Task{
			ID:         42,
			Status:     "processing",
			RetryCount: 0,
		},
	}

	queue := &mockQueue{}
	sender := &mockSender{}
	metrics := &mockMetrics{}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	retryStrategy := NewRetryStrategy(3)

	w := New(
		repo,
		queue,
		sender,
		retryStrategy,
		logger,
		metrics,
		5,
		100,
	)

	err := w.handleRetry(
		context.Background(),
		42,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.retryIncrements != 1 {
		t.Fatalf(
			"expected retry count to be incremented once, got %d",
			repo.retryIncrements,
		)
	}

	if len(queue.pushedTaskIDs) != 1 {
		t.Fatalf(
			"expected task to be pushed once, got %d",
			len(queue.pushedTaskIDs),
		)
	}

	if queue.pushedTaskIDs[0] != 42 {
		t.Fatalf(
			"expected task ID 42, got %d",
			queue.pushedTaskIDs[0],
		)
	}

	if metrics.retried != 1 {
		t.Fatalf(
			"expected retry metric 1, got %d",
			metrics.retried,
		)
	}
}

func TestWorker_HandleRetry_MaxRetriesReached(t *testing.T) {
	repo := &mockTaskRepository{
		task: model.Task{
			ID:         42,
			Status:     "processing",
			RetryCount: 3,
		},
	}

	queue := &mockQueue{}
	sender := &mockSender{}
	metrics := &mockMetrics{}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	retryStrategy := NewRetryStrategy(3)

	w := New(
		repo,
		queue,
		sender,
		retryStrategy,
		logger,
		metrics,
		5,
		100,
	)

	err := w.handleRetry(
		context.Background(),
		42,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.retryIncrements != 0 {
		t.Fatalf(
			"expected retry count not to be incremented, got %d",
			repo.retryIncrements,
		)
	}

	if len(queue.pushedTaskIDs) != 0 {
		t.Fatalf(
			"expected task not to be requeued, got %d pushes",
			len(queue.pushedTaskIDs),
		)
	}

	if len(repo.updatedStatuses) != 1 {
		t.Fatalf(
			"expected 1 status update, got %d",
			len(repo.updatedStatuses),
		)
	}

	if repo.updatedStatuses[0] != "failed" {
		t.Fatalf(
			"expected status failed, got %s",
			repo.updatedStatuses[0],
		)
	}

	if metrics.retried != 0 {
		t.Fatalf(
			"expected retry metric 0, got %d",
			metrics.retried,
		)
	}
}

func TestWorker_HandleRetry_QueueError(t *testing.T) {
	repo := &mockTaskRepository{
		task: model.Task{
			ID:         42,
			Status:     "processing",
			RetryCount: 0,
		},
	}

	queueErr := errors.New("queue error")

	queue := &mockQueue{
		pushErr: queueErr,
	}

	sender := &mockSender{}
	metrics := &mockMetrics{}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	retryStrategy := NewRetryStrategy(3)

	w := New(
		repo,
		queue,
		sender,
		retryStrategy,
		logger,
		metrics,
		5,
		100,
	)

	err := w.handleRetry(
		context.Background(),
		42,
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

	if repo.retryIncrements != 1 {
		t.Fatalf(
			"expected retry count incremented once, got %d",
			repo.retryIncrements,
		)
	}

	if metrics.retried != 0 {
		t.Fatalf(
			"expected retry metric 0, got %d",
			metrics.retried,
		)
	}
}

func TestWorker_HandleRetry_MarkFailedError(t *testing.T) {
	updateErr := errors.New("update status error")

	repo := &mockTaskRepository{
		task: model.Task{
			ID:         42,
			Status:     "processing",
			RetryCount: 3,
		},
		updateStatusErr: updateErr,
	}

	queue := &mockQueue{}
	sender := &mockSender{}
	metrics := &mockMetrics{}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	retryStrategy := NewRetryStrategy(3)

	w := New(
		repo,
		queue,
		sender,
		retryStrategy,
		logger,
		metrics,
		5,
		100,
	)

	err := w.handleRetry(
		context.Background(),
		42,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, updateErr) {
		t.Fatalf(
			"expected update status error, got %v",
			err,
		)
	}

	if repo.retryIncrements != 0 {
		t.Fatalf(
			"expected no retry increments, got %d",
			repo.retryIncrements,
		)
	}

	if len(queue.pushedTaskIDs) != 0 {
		t.Fatalf(
			"expected task not to be requeued, got %d pushes",
			len(queue.pushedTaskIDs),
		)
	}
}

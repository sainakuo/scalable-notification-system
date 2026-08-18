package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Job struct {
	TaskID int
}

type Worker struct {
	repo          TaskRepository
	queue         TaskQueue
	sender        NotificationSender
	retryStrategy *RetryStrategy
	logger        *slog.Logger
	workerCount   int
	jobBuffer     int
}

func New(repo TaskRepository,
	queue TaskQueue,
	sender NotificationSender,
	retryStrategy *RetryStrategy,
	logger *slog.Logger,
	workerCount int,
	jobBuffer int,
) *Worker {
	if repo == nil {
		panic("task repository is nil")
	}

	if queue == nil {
		panic("task queue is nil")
	}

	if sender == nil {
		panic("notification sender is nil")
	}

	if retryStrategy == nil {
		panic("retry strategy is nil")
	}

	if logger == nil {
		panic("logger is nil")
	}

	if workerCount <= 0 {
		panic("worker count must be greater than zero")
	}

	if jobBuffer <= 0 {
		panic("job buffer must be greater than zero")
	}

	return &Worker{
		repo:          repo,
		queue:         queue,
		sender:        sender,
		retryStrategy: retryStrategy,
		logger:        logger,
		workerCount:   workerCount,
		jobBuffer:     jobBuffer,
	}
}

func (w *Worker) consumeLoop(
	ctx context.Context,
	jobs chan<- Job,
) {
	for {
		taskID, err := w.queue.PopTask(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			w.logger.Error(
				"queue pop failed",
				"error", err,
			)
			continue
		}

		select {
		case jobs <- Job{TaskID: taskID}:
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) Run(ctx context.Context) error {
	jobs := make(chan Job, w.jobBuffer)

	var wg sync.WaitGroup

	for i := 1; i <= w.workerCount; i++ {
		wg.Add(1)

		go w.processLoop(ctx, i, jobs, &wg)
	}

	consumerDone := make(chan struct{})

	go func() {
		defer close(consumerDone)
		w.consumeLoop(ctx, jobs)
	}()

	<-ctx.Done()

	<-consumerDone

	close(jobs)

	wg.Wait()

	return nil
}

func (w *Worker) processLoop(
	ctx context.Context,
	workerID int,
	jobs <-chan Job,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for job := range jobs {
		processed, err := w.processTask(ctx, job.TaskID)

		if err != nil {
			w.logger.Error(
				"task processing failed",
				"worker_id", workerID,
				"task_id", job.TaskID,
				"error", err,
			)

			if retryErr := w.handleRetry(ctx, job.TaskID); retryErr != nil {
				w.logger.Error(
					"task retry failed",
					"worker_id", workerID,
					"task_id", job.TaskID,
					"error", retryErr,
				)
			}

			continue
		}

		if !processed {
			w.logger.Info(
				"task skipped",
				"worker_id", workerID,
				"task_id", job.TaskID,
				"reason", "already completed",
			)
			continue
		}

		w.logger.Info(
			"task processed",
			"worker_id", workerID,
			"task_id", job.TaskID,
		)
	}
}

func (w *Worker) processTask(
	ctx context.Context,
	taskID int,
) (bool, error) {
	task, err := w.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return false, fmt.Errorf("get task: %w", err)
	}

	if task.Status == "done" || task.Status == "failed" {
		return false, nil
	}

	if err := w.repo.UpdateStatus(ctx, task.ID, "processing"); err != nil {
		return false, fmt.Errorf("set task processing: %w", err)
	}

	if err := w.sender.Send(ctx, task); err != nil {
		return true, fmt.Errorf("send notification: %w", err)
	}

	if err := w.repo.UpdateStatus(ctx, task.ID, "done"); err != nil {
		return true, fmt.Errorf("set task done: %w", err)
	}

	return true, nil
}

func (w *Worker) handleRetry(
	ctx context.Context,
	taskID int,
) error {
	task, err := w.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get task for retry: %w", err)
	}

	if !w.retryStrategy.ShouldRetry(task.RetryCount) {
		if err := w.repo.UpdateStatus(ctx, taskID, "failed"); err != nil {
			return fmt.Errorf("mark task failed: %w", err)
		}

		return nil
	}

	if err := w.repo.IncrementRetryCount(ctx, taskID); err != nil {
		return fmt.Errorf("increment retry count: %w", err)
	}

	if err := w.queue.PushTask(ctx, taskID); err != nil {
		return fmt.Errorf("requeue task: %w", err)
	}

	return nil
}

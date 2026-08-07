package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/sainakuo/scalable-notification-system/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jackc/pgx/v5"
)

type TaskRepository struct {
	DB *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		DB: db,
	}
}

func (r *TaskRepository) CreateTask(ctx context.Context, task model.Task) (model.Task, error) {
	query := `
		INSERT INTO tasks (user_id, type, payload, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		task.UserID,
		task.Type,
		task.Payload,
		task.Status,
	).Scan(&task.ID, &task.CreatedAt)

	if err != nil {
		return model.Task{}, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

func (r *TaskRepository) GetTaskByID(ctx context.Context, id int) (model.Task, error) {
	query := `
		SELECT id, user_id, type, payload, status, retry_count, created_at
		FROM tasks
		WHERE id = $1
	`

	var task model.Task

	err := r.DB.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.Type,
		&task.Payload,
		&task.Status,
		&task.RetryCount,
		&task.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Task{}, model.ErrTaskNotFound
	}

	if err != nil {
		return model.Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return task, nil
}

func (r *TaskRepository) GetAllTasks(ctx context.Context) ([]model.Task, error) {
	query := `
		SELECT id, user_id, type, payload, status, retry_count, created_at
		FROM tasks
		ORDER BY created_at DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query all tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var task model.Task

		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Type,
			&task.Payload,
			&task.Status,
			&task.RetryCount,
			&task.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, taskID int, status string) error {

	query := `
		UPDATE tasks
		SET status = $1
		WHERE id = $2
	`

	_, err := r.DB.Exec(
		ctx,
		query,
		status,
		taskID,
	)

	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}

	return err
}

func (r *TaskRepository) IncrementRetryCount(ctx context.Context, taskID int) error {
	query := `
		UPDATE tasks
		SET retry_count = retry_count + 1
		WHERE id = $1
	`

	_, err := r.DB.Exec(
		ctx,
		query,
		taskID,
	)

	if err != nil {
		return fmt.Errorf("increment retry count: %w", err)
	}

	return err
}

func (r *TaskRepository) DeleteTask(ctx context.Context, taskID int) error {
	query := `
		DELETE FROM tasks
		WHERE id = $1
	`

	_, err := r.DB.Exec(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

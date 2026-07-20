package repository

import (
	"context"

	"github.com/sainakuo/scalable-notification-system/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	DB *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		DB: db,
	}
}

func (r *TaskRepository) CreateTask(task model.Task) (model.Task, error) {
	query := `
		INSERT INTO tasks (user_id, type, payload, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.DB.QueryRow(
		context.Background(),
		query,
		task.UserID,
		task.Type,
		task.Payload,
		task.Status,
	).Scan(&task.ID, &task.CreatedAt)

	if err != nil {
		return model.Task{}, err
	}

	return task, nil
}

func (r *TaskRepository) GetTaskByID(id int) (model.Task, error) {
	query := `
		SELECT id, user_id, type, payload, status, retry_count, created_at
		FROM tasks
		WHERE id = $1
	`

	var task model.Task

	err := r.DB.QueryRow(context.Background(), query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.Type,
		&task.Payload,
		&task.Status,
		&task.RetryCount,
		&task.CreatedAt,
	)

	if err != nil {
		return model.Task{}, err
	}

	return task, nil
}

func (r *TaskRepository) GetAllTasks() ([]model.Task, error) {
	query := `
		SELECT id, user_id, type, payload, status, retry_count, created_at
		FROM tasks
		ORDER BY created_at DESC
	`

	rows, err := r.DB.Query(context.Background(), query)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateStatus(taskID int, status string) error {

	query := `
		UPDATE tasks
		SET status = $1
		WHERE id = $2
	`

	_, err := r.DB.Exec(
		context.Background(),
		query,
		status,
		taskID,
	)

	return err
}

func (r *TaskRepository) IncrementRetryCount(taskID int) error {
	query := `
		UPDATE tasks
		SET retry_count = retry_count + 1
		WHERE id = $1
	`

	_, err := r.DB.Exec(
		context.Background(),
		query,
		taskID,
	)

	return err
}

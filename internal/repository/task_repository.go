package repository

import (
	"context"

	"github.com/sainakuo/scalable-notification-system/internal/model"

	"github.com/jackc/pgx/v5"
)

type TaskRepository struct {
	DB *pgx.Conn
}

func NewTaskRepository(db *pgx.Conn) *TaskRepository {
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

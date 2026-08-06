package handler

import (
	"time"

	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type TaskResponse struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func toTaskResponse(task model.Task) TaskResponse {
	return TaskResponse{
		ID:        task.ID,
		UserID:    task.UserID,
		Type:      task.Type,
		Payload:   task.Payload,
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
	}
}

func toTaskResponses(tasks []model.Task) []TaskResponse {
	responses := make([]TaskResponse, 0, len(tasks))

	for _, task := range tasks {
		responses = append(responses, toTaskResponse(task))
	}

	return responses
}

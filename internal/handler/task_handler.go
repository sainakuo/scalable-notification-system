package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sainakuo/scalable-notification-system/internal/model"
	"github.com/sainakuo/scalable-notification-system/internal/queue"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
)

type TaskHandler struct {
	Repo  *repository.TaskRepository
	Queue *queue.RedisQueue
}

func NewTaskHandler(repo *repository.TaskRepository, taskQueue *queue.RedisQueue) *TaskHandler {
	return &TaskHandler{
		Repo:  repo,
		Queue: taskQueue,
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task model.Task

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	task.Status = "pending"

	createdTask, err := h.Repo.CreateTask(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create task",
		})
		return
	}

	err = h.Queue.PushTask(context.Background(), createdTask.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to push task to queue",
		})
		return
	}

	c.JSON(http.StatusCreated, createdTask)
}

func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task id",
		})
		return
	}

	task, err := h.Repo.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "task not found",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.Repo.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get tasks",
		})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

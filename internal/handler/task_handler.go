package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sainakuo/scalable-notification-system/internal/model"
)

type TaskHandler struct {
	service TaskService
	logger  *slog.Logger
}

func NewTaskHandler(taskService TaskService, logger *slog.Logger) *TaskHandler {
	if taskService == nil {
		panic("task service is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}
	return &TaskHandler{
		service: taskService,
		logger:  logger,
	}
}

// CreateTask godoc
// @Summary Create a new task
// @Description Creates a task, saves it to PostgreSQL and pushes task ID to Redis queue
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body CreateTaskRequest true "Task data"
// @Success 201 {object} TaskResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	task := model.Task{
		UserID:  req.UserID,
		Type:    req.Type,
		Payload: req.Payload,
	}

	createdTask, err := h.service.CreateTask(c.Request.Context(), task)

	requestID := RequestIDFromContext(c.Request.Context())

	if err != nil {
		h.logger.Error(
			"create task failed",
			"request_id", requestID,
			"user_id", req.UserID,
			"task_type", req.Type,
			"error", err,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	h.logger.Info(
		"task created",
		"request_id", requestID,
		"task_id", createdTask.ID,
		"user_id", createdTask.UserID,
		"task_type", createdTask.Type,
	)

	c.JSON(http.StatusCreated, toTaskResponse(createdTask))
}

// GetTaskByID godoc
// @Summary Get task by ID
// @Description Returns task by ID
// @Tags tasks
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} TaskResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task id",
		})
		return
	}

	task, err := h.service.GetTaskByID(c.Request.Context(), id)

	requestID := RequestIDFromContext(c.Request.Context())

	if err != nil {
		if errors.Is(err, model.ErrTaskNotFound) {
			h.logger.Info(
				"task not found",
				"request_id", requestID,
				"task_id", id,
			)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "task not found",
			})
			return
		}
		h.logger.Error(
			"get task failed",
			"request_id", requestID,
			"task_id", id,
			"error", err,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, toTaskResponse(task))
}

// GetAllTasks godoc
// @Summary Get all tasks
// @Description Returns all tasks ordered by creation date
// @Tags tasks
// @Produce json
// @Success 200 {array} TaskResponse
// @Failure 500 {object} map[string]string
// @Router /tasks [get]
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.service.GetAllTasks(c.Request.Context())
	if err != nil {
		h.logger.Error(
			"get all tasks failed",
			"error", err,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, toTaskResponses(tasks))
}

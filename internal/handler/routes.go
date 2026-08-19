package handler

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, taskHandler *TaskHandler, healthHandler *HealthHandler) {
	router.GET("/health", healthHandler.Check)

	router.POST("/tasks", taskHandler.CreateTask)
	router.GET("/tasks", taskHandler.GetAllTasks)
	router.GET("/tasks/:id", taskHandler.GetTaskByID)
}

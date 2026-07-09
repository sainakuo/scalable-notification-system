package handler

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, taskHandler *TaskHandler) {
	router.GET("/health", HealthCheck)

	router.POST("/tasks", taskHandler.CreateTask)
	router.GET("/tasks", taskHandler.GetAllTasks)
	router.GET("/tasks/:id", taskHandler.GetTaskByID)
}

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "api",
	})
}

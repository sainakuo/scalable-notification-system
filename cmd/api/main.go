package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/handler"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
)

func main() {
	db, err := config.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close(context.Background())

	taskRepo := repository.NewTaskRepository(db)
	taskHandler := handler.NewTaskHandler(taskRepo)

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "api",
		})
	})

	router.POST("/tasks", taskHandler.CreateTask)

	router.GET("/tasks", taskHandler.GetAllTasks)
	router.GET("/tasks/:id", taskHandler.GetTaskByID)

	router.Run(":8080")
}

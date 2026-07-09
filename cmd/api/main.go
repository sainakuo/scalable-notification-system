package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/handler"
	"github.com/sainakuo/scalable-notification-system/internal/queue"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	db, err := config.ConnectDB(cfg.DatabaseURL())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(ctx)

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	taskQueue := queue.NewRedisQueue(redisClient)

	taskRepo := repository.NewTaskRepository(db)
	taskHandler := handler.NewTaskHandler(taskRepo, taskQueue)

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

	router.Run(":" + cfg.APIPort)
}

package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

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

	redisClient := config.ConnectRedis(cfg.RedisAddr)

	taskQueue := queue.NewRedisQueue(redisClient)

	taskRepo := repository.NewTaskRepository(db)
	taskHandler := handler.NewTaskHandler(taskRepo, taskQueue)

	router := gin.Default()

	handler.RegisterRoutes(router, taskHandler)

	router.Run(":" + cfg.APIPort)
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/sainakuo/scalable-notification-system/docs"
	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/handler"
	"github.com/sainakuo/scalable-notification-system/internal/queue"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Scalable Notification System API
// @version 1.0
// @description REST API for asynchronous task processing with Go, Gin, PostgreSQL, Redis, Workers and gRPC.
// @host localhost:8080
// @BasePath /
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
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	handler.RegisterRoutes(router, taskHandler)

	server := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.Println("API server started on port", cfg.APIPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down API server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatal("API server forced to shutdown:", err)
	}

	log.Println("API server exited")
}

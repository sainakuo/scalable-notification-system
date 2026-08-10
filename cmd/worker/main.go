package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/queue"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
	"github.com/sainakuo/scalable-notification-system/internal/sender"
	"github.com/sainakuo/scalable-notification-system/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	cfg := config.LoadConfig()

	db, err := config.ConnectPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	redisClient := config.ConnectRedis(cfg.RedisAddr)
	defer redisClient.Close()

	grpcConn, notificationClient, err := config.ConnectNotificationService(cfg.GRPCSenderAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcConn.Close()

	taskRepo := repository.NewTaskRepository(db)
	taskQueue := queue.NewRedisQueue(redisClient)
	notificationSender := sender.NewGRPCSender(notificationClient)

	taskWorker := worker.New(
		taskRepo,
		taskQueue,
		notificationSender,
		5,
		100,
		3,
	)

	log.Println("Worker started")

	if err := taskWorker.Run(ctx); err != nil &&
		!errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}

	log.Println("Worker stopped")
}

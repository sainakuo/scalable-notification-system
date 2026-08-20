package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/logger"
	"github.com/sainakuo/scalable-notification-system/internal/metrics"
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

	appLogger := logger.New()

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
	retryStrategy := worker.NewRetryStrategy(3)

	workerMetrics := metrics.NewWorkerMetrics()
	workerMetrics.Register()

	metricsServer := &http.Server{
		Addr:    ":9090",
		Handler: promhttp.Handler(),
	}

	go func() {
		log.Println("worker metrics server started on :9090")

		if err := metricsServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Println("metrics server error:", err)
		}
	}()

	taskWorker := worker.New(
		taskRepo,
		taskQueue,
		notificationSender,
		retryStrategy,
		appLogger,
		workerMetrics,
		5,
		100,
	)

	log.Println("Worker started")

	if err := taskWorker.Run(ctx); err != nil &&
		!errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Println("metrics server shutdown error:", err)
	}

	log.Println("Worker stopped")
}

package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/logger"
	"github.com/sainakuo/scalable-notification-system/internal/metrics"
	"github.com/sainakuo/scalable-notification-system/internal/queue"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
	"github.com/sainakuo/scalable-notification-system/internal/sender"
	"github.com/sainakuo/scalable-notification-system/internal/worker"
)

type WorkerApp struct {
	DB         *pgxpool.Pool
	Redis      *redis.Client
	GRPCConn   *grpc.ClientConn
	TaskWorker *worker.Worker
}

func BuildWorker(
	ctx context.Context,
	cfg config.Config,
) (*WorkerApp, error) {
	db, err := config.ConnectPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	redisClient := config.ConnectRedis(cfg.RedisAddr)

	grpcConn, notificationClient, err :=
		config.ConnectNotificationService(cfg.GRPCSenderAddr)
	if err != nil {
		db.Close()
		_ = redisClient.Close()

		return nil, fmt.Errorf("connect notification service: %w", err)
	}

	taskRepo := repository.NewTaskRepository(db)
	taskQueue := queue.NewRedisQueue(redisClient)
	notificationSender := sender.NewGRPCSender(notificationClient)

	retryStrategy := worker.NewRetryStrategy(3)

	appLogger := logger.New()

	workerMetrics := metrics.NewWorkerMetrics()
	workerMetrics.Register()

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

	return &WorkerApp{
		DB:         db,
		Redis:      redisClient,
		GRPCConn:   grpcConn,
		TaskWorker: taskWorker,
	}, nil
}

func (a *WorkerApp) Close() {
	if a.GRPCConn != nil {
		_ = a.GRPCConn.Close()
	}

	if a.Redis != nil {
		_ = a.Redis.Close()
	}

	if a.DB != nil {
		a.DB.Close()
	}
}

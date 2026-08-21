package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/handler"
	"github.com/sainakuo/scalable-notification-system/internal/health"
	"github.com/sainakuo/scalable-notification-system/internal/logger"
	"github.com/sainakuo/scalable-notification-system/internal/queue"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
	"github.com/sainakuo/scalable-notification-system/internal/service"
)

type API struct {
	DB            *pgxpool.Pool
	Redis         *redis.Client
	TaskHandler   *handler.TaskHandler
	HealthHandler *handler.HealthHandler
}

func BuildAPI(
	ctx context.Context,
	cfg config.Config,
) (*API, error) {
	db, err := config.ConnectPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	redisClient := config.ConnectRedis(cfg.RedisAddr)

	taskRepo := repository.NewTaskRepository(db)
	taskQueue := queue.NewRedisQueue(redisClient)
	taskService := service.NewTaskService(taskRepo, taskQueue)

	appLogger := logger.New()

	taskHandler := handler.NewTaskHandler(
		taskService,
		appLogger,
	)

	postgresChecker := health.NewPostgresChecker(db)
	redisChecker := health.NewRedisChecker(redisClient)

	healthHandler := handler.NewHealthHandler(
		postgresChecker,
		redisChecker,
	)

	return &API{
		DB:            db,
		Redis:         redisClient,
		TaskHandler:   taskHandler,
		HealthHandler: healthHandler,
	}, nil
}

func (a *API) Close() {
	if a.Redis != nil {
		_ = a.Redis.Close()
	}

	if a.DB != nil {
		a.DB.Close()
	}
}

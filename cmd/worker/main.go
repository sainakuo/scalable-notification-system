package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
	notificationpb "github.com/sainakuo/scalable-notification-system/proto/notificationpb"
)

const maxRetries = 3

type Job struct {
	TaskID int
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.LoadConfig()

	db, err := config.ConnectPostgres(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	redisClient := config.ConnectRedis(cfg.RedisAddr)

	grpcConn, notificationClient, err := config.ConnectNotificationService(cfg.GRPCSenderAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcConn.Close()

	taskRepo := repository.NewTaskRepository(db)

	jobs := make(chan Job, 100)

	var wg sync.WaitGroup
	workerCount := 5

	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go startWorker(ctx, i, jobs, taskRepo, redisClient, notificationClient, &wg)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			result, err := redisClient.BRPop(ctx, 0, "tasks_queue").Result()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Println("Redis error:", err)
				continue
			}

			taskID, err := strconv.Atoi(result[1])
			if err != nil {
				log.Println("Invalid task id:", result[1])
				continue
			}

			select {
			case jobs <- Job{TaskID: taskID}:
			case <-ctx.Done():
				return
			}
		}
	}()

	fmt.Println("Worker pool started...")

	<-quit

	log.Println("Shutting down worker...")
	cancel()
	close(jobs)
	wg.Wait()

	log.Println("Worker stopped")
}

func startWorker(
	ctx context.Context,
	workerID int,
	jobs <-chan Job,
	taskRepo *repository.TaskRepository,
	redisClient *redis.Client,
	notificationClient notificationpb.NotificationServiceClient,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for job := range jobs {
		fmt.Println("Worker", workerID, "processing task", job.TaskID)

		err := processTask(ctx, job.TaskID, taskRepo, notificationClient)

		if err != nil {
			log.Println("Worker", workerID, "failed task", job.TaskID, "error:", err)

			task, getErr := taskRepo.GetTaskByID(ctx, job.TaskID)
			if getErr != nil {
				log.Println("Failed to get task after error:", getErr)
				continue
			}

			if task.RetryCount < maxRetries {
				err = taskRepo.IncrementRetryCount(ctx, job.TaskID)
				if err != nil {
					log.Println("Failed to increment retry count:", err)
					continue
				}

				err = redisClient.LPush(ctx, "tasks_queue", strconv.Itoa(job.TaskID)).Err()
				if err != nil {
					log.Println("Failed to requeue task:", err)
					continue
				}

				log.Println("Task requeued:", job.TaskID)
			} else {
				err = taskRepo.UpdateStatus(ctx, job.TaskID, "failed")
				if err != nil {
					log.Println("Failed to mark task as failed:", err)
				}

				log.Println("Task marked as failed:", job.TaskID)
			}

			continue
		}

		fmt.Println("Worker", workerID, "finished task", job.TaskID)
	}
}

func processTask(ctx context.Context, taskID int, taskRepo *repository.TaskRepository, notificationClient notificationpb.NotificationServiceClient) error {
	task, err := taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	err = taskRepo.UpdateStatus(ctx, task.ID, "processing")
	if err != nil {
		return err
	}

	// Temporary fake error for testing retry
	if task.Type == "fail" {
		return fmt.Errorf("simulated task processing error")
	}

	response, err := notificationClient.SendNotification(
		ctx,
		&notificationpb.SendNotificationRequest{
			TaskId:  int32(task.ID),
			UserId:  int32(task.UserID),
			Type:    task.Type,
			Payload: task.Payload,
		},
	)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("notification sender failed: %s", response.Message)
	}

	err = taskRepo.UpdateStatus(ctx, task.ID, "done")
	if err != nil {
		return err
	}

	return nil
}

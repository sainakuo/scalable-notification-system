package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
)

const maxRetries = 3

type Job struct {
	TaskID int
}

func main() {
	ctx := context.Background()

	db, err := config.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close(ctx)

	redisClient := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		},
	)

	taskRepo := repository.NewTaskRepository(db)

	jobs := make(chan Job, 100)

	workerCount := 5

	for i := 1; i <= workerCount; i++ {
		go startWorker(i, jobs, taskRepo, redisClient)
	}

	fmt.Println("Worker pool started...")

	for {
		result, err := redisClient.BRPop(
			ctx,
			0,
			"tasks_queue",
		).Result()

		if err != nil {
			log.Println("Redis error: ", err)
			continue
		}

		taskID, err := strconv.Atoi(
			result[1],
		)

		if err != nil {
			log.Println("Invalid task id:", result[1])
			continue
		}

		jobs <- Job{TaskID: taskID}

	}
}

func startWorker(workerID int, jobs <-chan Job, taskRepo *repository.TaskRepository, redisClient *redis.Client) {
	for job := range jobs {
		fmt.Println("Worker", workerID, "processing task", job.TaskID)

		err := processTask(job.TaskID, taskRepo)

		if err != nil {
			log.Println("Worker", workerID, "failed task", job.TaskID, "error:", err)

			task, getErr := taskRepo.GetTaskByID(job.TaskID)
			if getErr != nil {
				log.Println("Failed to get task after error:", getErr)
				continue
			}

			if task.RetryCount < maxRetries {
				err = taskRepo.IncrementRetryCount(job.TaskID)
				if err != nil {
					log.Println("Failed to increment retry count:", err)
					continue
				}

				err = redisClient.LPush(context.Background(), "tasks_queue", strconv.Itoa(job.TaskID)).Err()
				if err != nil {
					log.Println("Failed to requeue task:", err)
					continue
				}

				log.Println("Task requeued:", job.TaskID)
			} else {
				err = taskRepo.UpdateStatus(job.TaskID, "failed")
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

func processTask(taskID int, taskRepo *repository.TaskRepository) error {
	task, err := taskRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	err = taskRepo.UpdateStatus(task.ID, "processing")
	if err != nil {
		return err
	}

	// Temporary fake error for testing retry
	if task.Type == "fail" {
		return fmt.Errorf("simulated task processing error")
	}

	time.Sleep(1 * time.Second)

	err = taskRepo.UpdateStatus(task.ID, "done")
	if err != nil {
		return err
	}

	return nil
}

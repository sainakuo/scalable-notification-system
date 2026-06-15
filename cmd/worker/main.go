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
		go startWorker(i, jobs, taskRepo)
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

func startWorker(workerID int, jobs <-chan Job, taskRepo *repository.TaskRepository) {
	for job := range jobs {
		fmt.Println("Worker", workerID, "processing task", job.TaskID)

		task, err := taskRepo.GetTaskByID(job.TaskID)

		if err != nil {
			log.Println("Failed to get task:", err)
			continue
		}

		err = taskRepo.UpdateStatus(
			task.ID,
			"processing",
		)
		if err != nil {
			log.Println("Failed to update status:", err)
			continue
		}

		time.Sleep(1 * time.Second)

		err = taskRepo.UpdateStatus(
			task.ID,
			"done",
		)
		if err != nil {
			log.Println("Failed to update status:", err)
			continue
		}

		fmt.Println("Worker", workerID, "finished task", task.ID)

	}

}

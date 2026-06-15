package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"

	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/repository"
)

func main() {
	db, err := config.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close(context.Background())

	redisClient := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		},
	)

	taskRepo := repository.NewTaskRepository(db)

	for {
		result, err := redisClient.BRPop(
			context.Background(),
			0,
			"tasks_queue",
		).Result()

		if err != nil {
			log.Println(err)
			continue
		}

		taskID, err := strconv.Atoi(
			result[1],
		)

		if err != nil {
			continue
		}

		task, err := taskRepo.GetTaskByID(taskID)

		if err != nil {
			continue
		}

		fmt.Println(
			"Processing task:",
			task.ID,
			task.Type,
		)

		taskRepo.UpdateStatus(
			task.ID,
			"processing",
		)

		taskRepo.UpdateStatus(
			task.ID,
			"done",
		)

	}

}

package model

import "time"

type Task struct {
	ID         int
	UserID     int
	Type       string
	Payload    string
	Status     string
	RetryCount int
	CreatedAt  time.Time
}

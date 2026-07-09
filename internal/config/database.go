package config

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func ConnectDB(databaseURL string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to PostgreSQL")

	return conn, nil
}

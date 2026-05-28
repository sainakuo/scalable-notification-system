package config

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func ConnectDB() (*pgx.Conn, error) {
	connString := "postgres://postgres:postgres@localhost:5432/sns_db"

	conn, err := pgx.Connect(context.Background(), connString)

	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to PostgreSQL")

	return conn, nil
}

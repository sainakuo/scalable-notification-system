package health

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresChecker struct {
	db *pgxpool.Pool
}

func NewPostgresChecker(db *pgxpool.Pool) *PostgresChecker {
	if db == nil {
		panic("postgres pool is nil")
	}

	return &PostgresChecker{
		db: db,
	}
}

func (c *PostgresChecker) Check(ctx context.Context) error {
	if err := c.db.Ping(ctx); err != nil {
		return fmt.Errorf("postgres health check: %w", err)
	}

	return nil
}

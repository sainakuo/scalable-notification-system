package sender

import "context"

type Deduplicator interface {
	IsProcessed(ctx context.Context, taskID int) (bool, error)
	MarkProcessed(ctx context.Context, taskID int) error
}

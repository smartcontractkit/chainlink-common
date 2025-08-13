package core

import (
	"context"
	"time"
)

type KeyValueStore interface {
	Store(ctx context.Context, key string, val []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	PruneExpiredEntries(ctx context.Context, maxAge time.Duration) (int64, error)
}

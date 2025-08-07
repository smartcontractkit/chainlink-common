package test

import (
	"context"
	"time"
)

type KeyValueStore struct {
}

func (t KeyValueStore) Store(ctx context.Context, key string, val []byte) error {
	return nil
}

func (t KeyValueStore) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (t KeyValueStore) PruneExpiredEntries(ctx context.Context, maxAge time.Duration) (int64, error) {
	return 0, nil
}

package test

import (
	"context"
)

type KeyValueStore struct {
}

func (t KeyValueStore) Store(ctx context.Context, key string, val any) error {
	return nil
}

func (t KeyValueStore) Get(ctx context.Context, key string, dest any) error {
	return nil
}

func (t KeyValueStore) StoreBytes(_ context.Context, key string, val []byte) error {
	return nil
}

func (t KeyValueStore) GetBytes(_ context.Context, key string) ([]byte, error) {
	return nil, nil
}

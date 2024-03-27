package types

import "context"

type KeyValueStore interface {
	Store(ctx context.Context, key string, val any) error
	Get(ctx context.Context, key string, dest any) error
	StoreBytes(ctx context.Context, key string, val []byte) error
	GetBytes(ctx context.Context, key string) ([]byte, error)
}

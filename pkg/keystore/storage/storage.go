package storage

import "context"

// Storage is where we store the encrypted key material.
type Storage interface {
	GetEncryptedKeystore(ctx context.Context) ([]byte, error)
	PutEncryptedKeystore(ctx context.Context, encryptedKeystore []byte) error
}

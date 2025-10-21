package storage

import (
	"context"
	"os"
)

const readWritePerms = os.FileMode(0600)

var _ Storage = &FileStorage{}

// FileStorage implements Storage using a file
type FileStorage struct {
	name string
}

func NewFileStorage(name string) *FileStorage {
	return &FileStorage{
		name: name,
	}
}

func (f *FileStorage) GetEncryptedKeystore(ctx context.Context) ([]byte, error) {
	return os.ReadFile(f.name)
}

func (f *FileStorage) PutEncryptedKeystore(ctx context.Context, encryptedKeystore []byte) error {
	return os.WriteFile(f.name, encryptedKeystore, readWritePerms)
}

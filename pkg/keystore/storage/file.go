package storage

import (
	"context"
	"os"
)

type FileStorage struct {
	filePath string
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{filePath: filePath}
}

func (s *FileStorage) GetEncryptedKeystore(ctx context.Context) ([]byte, error) {
	return os.ReadFile(s.filePath)
}

func (s *FileStorage) PutEncryptedKeystore(ctx context.Context, encryptedKeystore []byte) error {
	return os.WriteFile(s.filePath, encryptedKeystore, 0600)
}

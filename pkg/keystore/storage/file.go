package storage

import (
	"context"
	"fmt"
	"os"
)

type FileStorage struct {
	filePath string
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
	}
	return &FileStorage{filePath: filePath}, nil
}

func (s *FileStorage) GetEncryptedKeystore(ctx context.Context) ([]byte, error) {
	return os.ReadFile(s.filePath)
}

func (s *FileStorage) PutEncryptedKeystore(ctx context.Context, encryptedKeystore []byte) error {
	return os.WriteFile(s.filePath, encryptedKeystore, 0600)
}

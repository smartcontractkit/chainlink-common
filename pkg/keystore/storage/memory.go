package storage

import (
	"context"
	"sync"
)

// MemoryStorage implements Storage using in-memory storage
type MemoryStorage struct {
	mu   sync.RWMutex
	data []byte
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: []byte{}, // Initialize with empty byte slice instead of nil
	}
}

// GetEncryptedKeystore returns the stored encrypted keystore data
func (m *MemoryStorage) GetEncryptedKeystore(ctx context.Context) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy of the data
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

// PutEncryptedKeystore stores the encrypted keystore data
func (m *MemoryStorage) PutEncryptedKeystore(ctx context.Context, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store a copy of the data, treating nil as empty bytes
	if data == nil {
		m.data = []byte{}
	} else {
		m.data = make([]byte, len(data))
		copy(m.data, data)
	}

	return nil
}

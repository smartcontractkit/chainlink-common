package keystore

import (
	"context"
	"fmt"
)

type GetKeysRequest struct {
	Names []string // Empty slice means get all keys
}

type GetKeysResponse struct {
	Keys []KeyInfo
}

type Reader interface {
	GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error)
}

func (k *keystore) GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// If no names specified, return all keys
	if len(req.Names) == 0 {
		keys := make([]KeyInfo, 0, len(k.keystore))
		for name, key := range k.keystore {
			keys = append(keys, KeyInfo{
				Name:      name,
				KeyType:   key.keyType,
				PublicKey: key.publicKey(),
				Metadata:  key.metadata,
			})
		}
		return GetKeysResponse{Keys: keys}, nil
	}

	// Return specific keys
	keys := make([]KeyInfo, 0, len(req.Names))
	for _, name := range req.Names {
		key, ok := k.keystore[name]
		if !ok {
			return GetKeysResponse{}, fmt.Errorf("key not found: %s", name)
		}

		keys = append(keys, KeyInfo{
			Name:      name,
			KeyType:   key.keyType,
			PublicKey: key.publicKey(),
			Metadata:  key.metadata,
		})
	}

	return GetKeysResponse{Keys: keys}, nil
}

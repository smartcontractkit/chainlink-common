package keystore

import (
	"context"
	"fmt"
)

type ListKeysRequest struct{}

type ListKeysResponse struct {
	Keys []KeyInfo
}

type GetKeysRequest struct {
	Names []string
}

type GetKeysResponse struct {
	Keys []GetKeyResponse
}

type GetKeyRequest struct {
	Name string
}

type GetKeyResponse struct {
	KeyInfo KeyInfo
}

type Reader interface {
	ListKeys(ctx context.Context, req ListKeysRequest) (ListKeysResponse, error)
	GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error)
}

func (k *keystore) ListKeys(ctx context.Context, req ListKeysRequest) (ListKeysResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	keys := make([]KeyInfo, 0, len(k.keystore))
	for name, key := range k.keystore {
		keys = append(keys, KeyInfo{
			Name:      name,
			KeyType:   key.keyType,
			PublicKey: key.publicKey(),
			Metadata:  key.metadata,
		})
	}

	return ListKeysResponse{Keys: keys}, nil
}

func (k *keystore) GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	responses := make([]GetKeyResponse, 0, len(req.Names))
	for _, name := range req.Names {
		key, ok := k.keystore[name]
		if !ok {
			return GetKeysResponse{}, fmt.Errorf("key not found: %s", name)
		}

		responses = append(responses, GetKeyResponse{
			KeyInfo: KeyInfo{
				Name:      name,
				KeyType:   key.keyType,
				PublicKey: key.publicKey(),
				Metadata:  key.metadata,
			},
		})
	}

	return GetKeysResponse{Keys: responses}, nil
}

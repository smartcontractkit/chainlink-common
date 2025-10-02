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
	KeyNames []string
}

type GetKeysResponse struct {
	Keys []GetKeyResponse
}

type GetKeyRequest struct {
	KeyName string
}

type GetKeyResponse struct {
	KeyInfo KeyInfo
}

type Reader interface {
	GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error)
}

func (k *keystore) GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	// If no names provided, return all keys
	if len(req.KeyNames) == 0 {
		responses := make([]GetKeyResponse, 0, len(k.keystore))
		for name, key := range k.keystore {
			responses = append(responses, GetKeyResponse{
				KeyInfo: KeyInfo{
					Name:      name,
					KeyType:   key.keyType,
					PublicKey: key.publicKey,
					Metadata:  key.metadata,
				},
			})
		}
		return GetKeysResponse{Keys: responses}, nil
	}

	responses := make([]GetKeyResponse, 0, len(req.KeyNames))
	for _, name := range req.KeyNames {
		key, ok := k.keystore[name]
		if !ok {
			return GetKeysResponse{}, fmt.Errorf("key not found: %s", name)
		}
		responses = append(responses, GetKeyResponse{
			KeyInfo: KeyInfo{
				Name:      name,
				KeyType:   key.keyType,
				PublicKey: key.publicKey,
				Metadata:  key.metadata,
			},
		})
	}

	return GetKeysResponse{Keys: responses}, nil
}

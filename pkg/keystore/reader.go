package keystore

import "context"

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
	return ListKeysResponse{}, nil
}

func (k *keystore) GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error) {
	return GetKeysResponse{}, nil
}

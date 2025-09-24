package keystore

import "context"

type ListKeysRequest struct{}

type ListKeysResponse struct {
	Keys []KeyInfo
}

type GetKeyRequest struct {
	Name string
}

type GetKeyResponse struct {
	KeyInfo KeyInfo
}

type Reader interface {
	ListKeys(ctx context.Context, req ListKeysRequest) (ListKeysResponse, error)
	GetKey(ctx context.Context, req GetKeyRequest) (GetKeyResponse, error)
}

func (k *keystore) ListKeys(ctx context.Context, req ListKeysRequest) (ListKeysResponse, error) {
	return ListKeysResponse{}, nil
}

func (k *keystore) GetKey(ctx context.Context, req GetKeyRequest) (GetKeyResponse, error) {
	return GetKeyResponse{}, nil
}

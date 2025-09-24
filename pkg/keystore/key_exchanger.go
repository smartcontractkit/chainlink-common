package keystore

import "context"

type DeriveSharedSecretRequest struct {
	LocalKeyName string
	RemotePubKey []byte
}

type DeriveSharedSecretResponse struct {
	SharedSecret []byte
}

type KeyExchanger interface {
	DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error)
}

func (k *keystore) DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error) {
	return DeriveSharedSecretResponse{}, nil
}

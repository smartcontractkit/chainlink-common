package keystore

import (
	"context"
)

type EncryptRequest struct {
	KeyName string
	Data    []byte
}

type EncryptResponse struct {
	EncryptedData []byte
}

type DecryptRequest struct {
	KeyName       string
	EncryptedData []byte
}

type DecryptResponse struct {
	Data []byte
}

type DeriveSharedSecretRequest struct {
	LocalKeyName string
	RemotePubKey []byte // Maybe this naming is confusing?
}

type DeriveSharedSecretResponse struct {
	SharedSecret []byte
}

type Encryptor interface {
	Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error)
	Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error)
	DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error)
}

// TODO: Encryptor implementation.
func (k *keystore) Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error) {
	return EncryptResponse{}, nil
}

func (k *keystore) Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error) {
	return DecryptResponse{}, nil
}

func (k *keystore) DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error) {
	return DeriveSharedSecretResponse{}, nil
}

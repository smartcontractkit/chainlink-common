package keystore

import (
	"context"
)

const (
	X25519 KeyType = "X25519"
	// TODO: Support P256
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

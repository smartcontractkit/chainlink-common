package keystore

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore/internal"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
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

func (k *keystore) Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.KeyName]
	if !ok {
		return EncryptResponse{}, fmt.Errorf("key not found: %s", req.KeyName)
	}
	switch key.keyType {
	case X25519:
		encrypted, err := box.SealAnonymous(nil, req.Data, (*[32]byte)(key.publicKey()), rand.Reader)
		if err != nil {
			return EncryptResponse{}, fmt.Errorf("failed to encrypt data: %w", err)
		}
		return EncryptResponse{
			EncryptedData: encrypted,
		}, nil
	default:
		return EncryptResponse{}, fmt.Errorf("unsupported key type: %s", key.keyType)
	}
}

func (k *keystore) Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.KeyName]
	if !ok {
		return DecryptResponse{}, fmt.Errorf("key not found: %s", req.KeyName)
	}
	switch key.keyType {
	case X25519:
		decrypted, ok := box.OpenAnonymous(nil, req.EncryptedData, (*[32]byte)(key.publicKey()), (*[32]byte)(internal.Bytes(key.privateKey)))
		if !ok {
			return DecryptResponse{}, fmt.Errorf("failed to decrypt data")
		}
		return DecryptResponse{
			Data: decrypted,
		}, nil
	default:
		return DecryptResponse{}, fmt.Errorf("unsupported key type: %s", key.keyType)
	}
}

func (k *keystore) DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.LocalKeyName]
	if !ok {
		return DeriveSharedSecretResponse{}, fmt.Errorf("key not found: %s", req.LocalKeyName)
	}
	switch key.keyType {
	case X25519:
		if len(req.RemotePubKey) != 32 {
			return DeriveSharedSecretResponse{}, fmt.Errorf("remote public key must be 32 bytes")
		}
		sharedSecret, err := curve25519.X25519(internal.Bytes(key.privateKey), req.RemotePubKey)
		if err != nil {
			return DeriveSharedSecretResponse{}, fmt.Errorf("failed to derive shared secret: %w", err)
		}
		return DeriveSharedSecretResponse{
			SharedSecret: sharedSecret,
		}, nil
	default:
		return DeriveSharedSecretResponse{}, fmt.Errorf("unsupported key type: %s", key.keyType)
	}
}

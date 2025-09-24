package keystore

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore/serialization"
	"google.golang.org/protobuf/proto"
)

type CreateKeyRequest struct {
	Name    string
	KeyType KeyType
}

type CreateKeyResponse struct {
	KeyInfo KeyInfo
}

type DeleteKeyRequest struct {
	Name string
}

type DeleteKeyResponse struct{}

type ImportKeyRequest struct {
	Name    string
	KeyType KeyType
	Data    []byte
}

type ImportKeyResponse struct {
	PublicKey []byte
}

type ExportKeyRequest struct {
	Name string
}

type ExportKeyResponse struct {
	Data []byte
}

type Admin interface {
	CreateKey(ctx context.Context, req CreateKeyRequest) (CreateKeyResponse, error)
	DeleteKey(ctx context.Context, req DeleteKeyRequest) (DeleteKeyResponse, error)
	ImportKey(ctx context.Context, req ImportKeyRequest) (ImportKeyResponse, error)
	ExportKey(ctx context.Context, req ExportKeyRequest) (ExportKeyResponse, error)
}

func (k *keystore) CreateKey(ctx context.Context, req CreateKeyRequest) (CreateKeyResponse, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Clone the keystore.
	ks := proto.Clone(k.keystore).(*serialization.Keystore)

	// Mutate it with the new key.
	switch req.KeyType {
	case Ed25519:
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return CreateKeyResponse{}, fmt.Errorf("failed to generate Ed25519 key: %w", err)
		}
		ks.Ed25519Keys = append(ks.Ed25519Keys, &serialization.Ed25519Key{
			Name:       req.Name,
			PrivateKey: privateKey,
			CreatedAt:  time.Now().Unix(),
		})
	case Secp256k1:
		return CreateKeyResponse{}, fmt.Errorf("Secp256k1 is not supported yet")
	}
	// Persist it to storage.
	if err := save(k.storage, k.password, k.keystore); err != nil {
		return CreateKeyResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}
	// If we succeed to save, update the in memory keystore.
	k.keystore = ks
	return CreateKeyResponse{}, nil
}

func (k *keystore) DeleteKey(ctx context.Context, req DeleteKeyRequest) (DeleteKeyResponse, error) {
	return DeleteKeyResponse{}, nil
}

func (k *keystore) ImportKey(ctx context.Context, req ImportKeyRequest) (ImportKeyResponse, error) {
	return ImportKeyResponse{}, nil
}

func (k *keystore) ExportKey(ctx context.Context, req ExportKeyRequest) (ExportKeyResponse, error) {
	return ExportKeyResponse{}, nil
}

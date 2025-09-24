package keystore

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"maps"
	"time"

	gethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/internal"
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
	KeyInfo KeyInfo
	Data    []byte
}

type Admin interface {
	CreateKey(ctx context.Context, req CreateKeyRequest) (CreateKeyResponse, error)
	DeleteKey(ctx context.Context, req DeleteKeyRequest) (DeleteKeyResponse, error)
	ImportKey(ctx context.Context, req ImportKeyRequest) (ImportKeyResponse, error)
	ExportKey(ctx context.Context, req ExportKeyRequest) (ExportKeyResponse, error)
}

func (ks *keystore) CreateKey(ctx context.Context, req CreateKeyRequest) (CreateKeyResponse, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Clone the keystore.
	ksCopy := maps.Clone(ks.keystore)

	// Mutate it with the new key.
	switch req.KeyType {
	case Ed25519:
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return CreateKeyResponse{}, fmt.Errorf("failed to generate Ed25519 key: %w", err)
		}
		ksCopy[req.Name] = key{
			keyType:    req.KeyType,
			privateKey: internal.NewRaw(privateKey),
			createdAt:  time.Now(),
		}
	case Secp256k1:
		privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		if err != nil {
			return CreateKeyResponse{}, fmt.Errorf("failed to generate Secp256k1 key: %w", err)
		}
		ksCopy[req.Name] = key{
			keyType:    req.KeyType,
			privateKey: internal.NewRaw(privateKeyECDSA.D.Bytes()),
			createdAt:  time.Now(),
		}
		return CreateKeyResponse{}, fmt.Errorf("Secp256k1 is not supported yet")
	}
	// Persist it to storage.
	if err := save(ks.storage, ks.password, ksCopy); err != nil {
		return CreateKeyResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}
	// If we succeed to save, update the in memory keystore.
	ks.keystore = ksCopy
	return CreateKeyResponse{}, nil
}

func (k *keystore) DeleteKey(ctx context.Context, req DeleteKeyRequest) (DeleteKeyResponse, error) {
	return DeleteKeyResponse{}, nil
}

func (k *keystore) ImportKey(ctx context.Context, req ImportKeyRequest) (ImportKeyResponse, error) {
	return ImportKeyResponse{}, nil
}

func (k *keystore) ExportKey(ctx context.Context, req ExportKeyRequest) (ExportKeyResponse, error) {
	key, ok := k.keystore[req.Name]
	if !ok {
		return ExportKeyResponse{}, fmt.Errorf("key not found: %s", req.Name)
	}
	exportedKey, err := gethkeystore.EncryptDataV3(internal.Bytes(key.privateKey), []byte(k.password), gethkeystore.StandardScryptN, gethkeystore.StandardScryptP)
	if err != nil {
		return ExportKeyResponse{}, fmt.Errorf("failed to export key: %w", err)
	}
	exportedKeyData, err := json.Marshal(exportedKey)
	if err != nil {
		return ExportKeyResponse{}, fmt.Errorf("failed to marshal exported key: %w", err)
	}
	return ExportKeyResponse{
		KeyInfo: KeyInfo{
			Name:      req.Name,
			KeyType:   key.keyType,
			PublicKey: []byte{}, // TODO
		},
		Data: exportedKeyData,
	}, nil
}

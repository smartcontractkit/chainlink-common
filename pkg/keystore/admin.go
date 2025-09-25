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

	"golang.org/x/crypto/curve25519"

	gethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/internal"
)

type CreateKeysRequest struct {
	Keys []CreateKeyRequest
}

type CreateKeyRequest struct {
	Name    string
	KeyType KeyType
}

type CreateKeysResponse struct {
	Keys []CreateKeyResponse
}

type CreateKeyResponse struct {
	KeyInfo KeyInfo
}

type DeleteKeysRequest struct {
	Names []string
}

type DeleteKeysResponse struct{}

type ImportKeysRequest struct {
	Keys []ImportKeyRequest
}

type ImportKeyRequest struct {
	Name    string
	KeyType KeyType
	Data    []byte
}

type ImportKeysResponse struct {
	Keys []ImportKeyResponse
}

type ImportKeyResponse struct {
	PublicKey []byte
}

type ExportKeysRequest struct {
	Names []string
}

type ExportKeysResponse struct {
	Keys []ExportKeyResponse
}

type ExportKeyRequest struct {
	Name string
}

type ExportKeyResponse struct {
	KeyInfo KeyInfo
	Data    []byte
}

type Admin interface {
	CreateKeys(ctx context.Context, req CreateKeysRequest) (CreateKeysResponse, error)
	DeleteKeys(ctx context.Context, req DeleteKeysRequest) (DeleteKeysResponse, error)
	ImportKeys(ctx context.Context, req ImportKeysRequest) (ImportKeysResponse, error)
	ExportKeys(ctx context.Context, req ExportKeysRequest) (ExportKeysResponse, error)
}

func (ks *keystore) CreateKeys(ctx context.Context, req CreateKeysRequest) (CreateKeysResponse, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Clone the keystore.
	ksCopy := maps.Clone(ks.keystore)
	var responses []CreateKeyResponse

	// Create all keys
	for _, keyReq := range req.Keys {
		switch keyReq.KeyType {
		case Ed25519:
			_, privateKey, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate Ed25519 key: %w", err)
			}
			ksCopy[keyReq.Name] = key{
				keyType:    keyReq.KeyType,
				privateKey: internal.NewRaw(privateKey),
				createdAt:  time.Now(),
			}
		case Secp256k1:
			privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate Secp256k1 key: %w", err)
			}
			ksCopy[keyReq.Name] = key{
				keyType:    keyReq.KeyType,
				privateKey: internal.NewRaw(privateKeyECDSA.D.Bytes()),
				createdAt:  time.Now(),
			}
		case X25519:
			encryptionKey := [curve25519.ScalarSize]byte{}
			_, err := rand.Read(encryptionKey[:])
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate Curve25519 key: %w", err)
			}
			ksCopy[keyReq.Name] = key{
				keyType:    keyReq.KeyType,
				privateKey: internal.NewRaw(encryptionKey[:]),
				createdAt:  time.Now(),
			}
		default:
			return CreateKeysResponse{}, fmt.Errorf("unsupported key type: %s", keyReq.KeyType)
		}

		responses = append(responses, CreateKeyResponse{
			KeyInfo: KeyInfo{
				Name:      keyReq.Name,
				KeyType:   keyReq.KeyType,
				PublicKey: ksCopy[keyReq.Name].publicKey(),
			},
		})
	}

	// Persist it to storage.
	if err := save(ks.storage, ks.password, ksCopy); err != nil {
		return CreateKeysResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}
	// If we succeed to save, update the in memory keystore.
	ks.keystore = ksCopy
	return CreateKeysResponse{Keys: responses}, nil
}

func (k *keystore) DeleteKeys(ctx context.Context, req DeleteKeysRequest) (DeleteKeysResponse, error) {
	return DeleteKeysResponse{}, nil
}

func (k *keystore) ImportKeys(ctx context.Context, req ImportKeysRequest) (ImportKeysResponse, error) {
	return ImportKeysResponse{}, nil
}

func (k *keystore) ExportKeys(ctx context.Context, req ExportKeysRequest) (ExportKeysResponse, error) {
	var responses []ExportKeyResponse

	for _, name := range req.Names {
		key, ok := k.keystore[name]
		if !ok {
			return ExportKeysResponse{}, fmt.Errorf("key not found: %s", name)
		}
		exportedKey, err := gethkeystore.EncryptDataV3(internal.Bytes(key.privateKey), []byte(k.password), gethkeystore.StandardScryptN, gethkeystore.StandardScryptP)
		if err != nil {
			return ExportKeysResponse{}, fmt.Errorf("failed to export key %s: %w", name, err)
		}
		exportedKeyData, err := json.Marshal(exportedKey)
		if err != nil {
			return ExportKeysResponse{}, fmt.Errorf("failed to marshal exported key %s: %w", name, err)
		}
		responses = append(responses, ExportKeyResponse{
			KeyInfo: KeyInfo{
				Name:      name,
				KeyType:   key.keyType,
				PublicKey: key.publicKey(),
			},
			Data: exportedKeyData,
		})
	}

	return ExportKeysResponse{Keys: responses}, nil
}

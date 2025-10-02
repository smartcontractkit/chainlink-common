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

type SetMetadataRequest struct {
	Updates []SetMetadataUpdate
}

type SetMetadataUpdate struct {
	KeyName  string
	Metadata []byte
}

type SetMetadataResponse struct{}

type Admin interface {
	CreateKeys(ctx context.Context, req CreateKeysRequest) (CreateKeysResponse, error)
	DeleteKeys(ctx context.Context, req DeleteKeysRequest) (DeleteKeysResponse, error)
	ImportKeys(ctx context.Context, req ImportKeysRequest) (ImportKeysResponse, error)
	ExportKeys(ctx context.Context, req ExportKeysRequest) (ExportKeysResponse, error)
	SetMetadata(ctx context.Context, req SetMetadataRequest) (SetMetadataResponse, error)
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
				metadata:   []byte{},
			}
		case EcdsaSecp256k1:
			privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate EcdsaSecp256k1 key: %w", err)
			}
			ksCopy[keyReq.Name] = key{
				keyType:    keyReq.KeyType,
				privateKey: internal.NewRaw(privateKeyECDSA.D.Bytes()),
				createdAt:  time.Now(),
				metadata:   []byte{},
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
				metadata:   []byte{},
			}
		default:
			return CreateKeysResponse{}, fmt.Errorf("unsupported key type: %s", keyReq.KeyType)
		}

		responses = append(responses, CreateKeyResponse{
			KeyInfo: KeyInfo{
				Name:      keyReq.Name,
				KeyType:   keyReq.KeyType,
				PublicKey: ksCopy[keyReq.Name].publicKey(),
				Metadata:  ksCopy[keyReq.Name].metadata,
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
	k.mu.Lock()
	defer k.mu.Unlock()

	ksCopy := maps.Clone(k.keystore)
	for _, name := range req.Names {
		delete(ksCopy, name)
	}
	if err := save(k.storage, k.password, ksCopy); err != nil {
		return DeleteKeysResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}
	k.keystore = ksCopy
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
				Metadata:  key.metadata,
			},
			Data: exportedKeyData,
		})
	}

	return ExportKeysResponse{Keys: responses}, nil
}

func (ks *keystore) SetMetadata(ctx context.Context, req SetMetadataRequest) (SetMetadataResponse, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ksCopy := maps.Clone(ks.keystore)

	for _, update := range req.Updates {
		key, ok := ksCopy[update.KeyName]
		if !ok {
			return SetMetadataResponse{}, fmt.Errorf("key not found: %s", update.KeyName)
		}

		key.metadata = update.Metadata
		ksCopy[update.KeyName] = key
	}

	if err := save(ks.storage, ks.password, ksCopy); err != nil {
		return SetMetadataResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}

	ks.keystore = ksCopy
	return SetMetadataResponse{}, nil
}

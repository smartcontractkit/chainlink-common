package keystore

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"maps"
	"time"

	"golang.org/x/crypto/curve25519"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/internal"
)

type CreateKeysRequest struct {
	Keys []CreateKeyRequest
}

type CreateKeyRequest struct {
	KeyName string
	KeyType KeyType
}

type CreateKeysResponse struct {
	Keys []CreateKeyResponse
}

type CreateKeyResponse struct {
	KeyInfo KeyInfo
}

type DeleteKeysRequest struct {
	KeyNames []string
}

type DeleteKeysResponse struct{}

type ImportKeysRequest struct {
	Keys []ImportKeyRequest
}

type ImportKeyRequest struct {
	KeyName string
	KeyType KeyType
	Data    []byte
}

type ImportKeysResponse struct{}

type ExportKeyParam struct {
	KeyName string
	Enc     EncryptionParams
}

type ExportKeysRequest struct {
	Keys []ExportKeyParam
}

type ExportKeysResponse struct {
	Keys []ExportKeyResponse
}

type ExportKeyResponse struct {
	KeyName string
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

	ksCopy := maps.Clone(ks.keystore)
	var responses []CreateKeyResponse
	for _, keyReq := range req.Keys {
		if _, ok := ksCopy[keyReq.KeyName]; ok {
			return CreateKeysResponse{}, fmt.Errorf("key already exists: %s", keyReq.KeyName)
		}
		switch keyReq.KeyType {
		case Ed25519:
			_, privateKey, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate Ed25519 key: %w", err)
			}
			publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(privateKey), keyReq.KeyType)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to get public key from private key: %w", err)
			}
			ksCopy[keyReq.KeyName] = key{
				keyType:    keyReq.KeyType,
				privateKey: internal.NewRaw(privateKey),
				publicKey:  publicKey,
				createdAt:  time.Now(),
				metadata:   []byte{},
			}
		case EcdsaSecp256k1:
			privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate EcdsaSecp256k1 key: %w", err)
			}
			publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(privateKey.D.Bytes()), keyReq.KeyType)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to get public key from private key: %w", err)
			}
			ksCopy[keyReq.KeyName] = key{
				keyType:    keyReq.KeyType,
				privateKey: internal.NewRaw(privateKey.D.Bytes()),
				publicKey:  publicKey,
				createdAt:  time.Now(),
				metadata:   []byte{},
			}
		case X25519:
			privateKey := [curve25519.ScalarSize]byte{}
			_, err := rand.Read(privateKey[:])
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate Curve25519 key: %w", err)
			}
			publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(privateKey[:]), keyReq.KeyType)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to get public key from private key: %w", err)
			}
			ksCopy[keyReq.KeyName] = key{
				keyType:    keyReq.KeyType,
				privateKey: internal.NewRaw(privateKey[:]),
				publicKey:  publicKey,
				createdAt:  time.Now(),
				metadata:   []byte{},
			}
		default:
			return CreateKeysResponse{}, fmt.Errorf("unsupported key type: %s", keyReq.KeyType)
		}

		responses = append(responses, CreateKeyResponse{
			KeyInfo: KeyInfo{
				Name:      keyReq.KeyName,
				KeyType:   keyReq.KeyType,
				PublicKey: ksCopy[keyReq.KeyName].publicKey,
				Metadata:  ksCopy[keyReq.KeyName].metadata,
			},
		})
	}

	// Persist it to storage.
	if err := save(ks.storage, ks.enc, ksCopy); err != nil {
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
	for _, name := range req.KeyNames {
		if _, ok := ksCopy[name]; !ok {
			return DeleteKeysResponse{}, fmt.Errorf("key not found: %s", name)
		}
		delete(ksCopy, name)
	}
	if err := save(k.storage, k.enc, ksCopy); err != nil {
		return DeleteKeysResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}
	k.keystore = ksCopy
	return DeleteKeysResponse{}, nil
}

func (k *keystore) ImportKeys(ctx context.Context, req ImportKeysRequest) (ImportKeysResponse, error) {
	return ImportKeysResponse{}, nil
}

func (k *keystore) ExportKeys(ctx context.Context, req ExportKeysRequest) (ExportKeysResponse, error) {
	return ExportKeysResponse{}, nil
}

func (ks *keystore) SetMetadata(ctx context.Context, req SetMetadataRequest) (SetMetadataResponse, error) {
	return SetMetadataResponse{}, nil
}

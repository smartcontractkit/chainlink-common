package keystore

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"maps"
	"time"

	"golang.org/x/crypto/curve25519"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

var (
	ErrKeyAlreadyExists   = fmt.Errorf("key already exists")
	ErrInvalidKeyName     = fmt.Errorf("invalid key name")
	ErrKeyNotFound        = fmt.Errorf("key not found")
	ErrUnsupportedKeyType = fmt.Errorf("unsupported key type")
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

// UnimplementedAdmin returns ErrUnimplemented for all Admin methods.
type UnimplementedAdmin struct{}

func (UnimplementedAdmin) CreateKeys(ctx context.Context, req CreateKeysRequest) (CreateKeysResponse, error) {
	return CreateKeysResponse{}, fmt.Errorf("Admin.CreateKeys: %w", ErrUnimplemented)
}

func (UnimplementedAdmin) DeleteKeys(ctx context.Context, req DeleteKeysRequest) (DeleteKeysResponse, error) {
	return DeleteKeysResponse{}, fmt.Errorf("Admin.DeleteKeys: %w", ErrUnimplemented)
}

func (UnimplementedAdmin) ImportKeys(ctx context.Context, req ImportKeysRequest) (ImportKeysResponse, error) {
	return ImportKeysResponse{}, fmt.Errorf("Admin.ImportKeys: %w", ErrUnimplemented)
}

func (UnimplementedAdmin) ExportKeys(ctx context.Context, req ExportKeysRequest) (ExportKeysResponse, error) {
	return ExportKeysResponse{}, fmt.Errorf("Admin.ExportKeys: %w", ErrUnimplemented)
}

func (UnimplementedAdmin) SetMetadata(ctx context.Context, req SetMetadataRequest) (SetMetadataResponse, error) {
	return SetMetadataResponse{}, fmt.Errorf("Admin.SetMetadata: %w", ErrUnimplemented)
}

func ValidKeyName(name string) error {
	if name == "" {
		return fmt.Errorf("key name cannot be empty")
	}
	// Just a sanity bound.
	if len(name) > 1_000 {
		return fmt.Errorf("key name cannot be longer than 1000 characters")
	}
	return nil
}

// CreateKeys creates multiple keys in a single operation. The response preserves the order of the request.
// It's atomic - either all keys are created or none are created.
func (ks *keystore) CreateKeys(ctx context.Context, req CreateKeysRequest) (CreateKeysResponse, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ksCopy := maps.Clone(ks.keystore)
	var responses []CreateKeyResponse
	for _, keyReq := range req.Keys {
		if _, ok := ksCopy[keyReq.KeyName]; ok {
			return CreateKeysResponse{}, fmt.Errorf("%w: %s", ErrKeyAlreadyExists, keyReq.KeyName)
		}
		if err := ValidKeyName(keyReq.KeyName); err != nil {
			return CreateKeysResponse{}, fmt.Errorf("%w: %s", ErrInvalidKeyName, err)
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
			ksCopy[keyReq.KeyName] = newKey(keyReq.KeyType, internal.NewRaw(privateKey), publicKey, time.Now(), []byte{})
		case ECDSA_S256:
			privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate ECDSA_S256 key: %w", err)
			}
			publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(privateKey.D.Bytes()), keyReq.KeyType)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to get public key from private key: %w", err)
			}
			ksCopy[keyReq.KeyName] = newKey(keyReq.KeyType, internal.NewRaw(privateKey.D.Bytes()), publicKey, time.Now(), []byte{})
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
			ksCopy[keyReq.KeyName] = newKey(keyReq.KeyType, internal.NewRaw(privateKey[:]), publicKey, time.Now(), []byte{})
		case ECDH_P256:
			privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate ECDH_P256 key: %w", err)
			}
			publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(privateKey.Bytes()), keyReq.KeyType)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to get public key from private key: %w", err)
			}
			ksCopy[keyReq.KeyName] = newKey(keyReq.KeyType, internal.NewRaw(privateKey.Bytes()), publicKey, time.Now(), []byte{})
		default:
			return CreateKeysResponse{}, fmt.Errorf("%w: %s", ErrUnsupportedKeyType, keyReq.KeyType)
		}

		created := ksCopy[keyReq.KeyName].createdAt
		k := ksCopy[keyReq.KeyName]
		responses = append(responses, CreateKeyResponse{
			KeyInfo: newKeyInfo(keyReq.KeyName, keyReq.KeyType, created, k.publicKey, k.metadata),
		})
	}

	// Persist it to storage.
	if err := ks.save(ctx, ksCopy); err != nil {
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
			return DeleteKeysResponse{}, fmt.Errorf("%w: %s", ErrKeyNotFound, name)
		}
		delete(ksCopy, name)
	}
	if err := k.save(ctx, ksCopy); err != nil {
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

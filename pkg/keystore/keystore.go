package keystore

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/curve25519"

	gethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/serialization"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"google.golang.org/protobuf/proto"
)

type KeyType string

type KeyInfo struct {
	Name      string
	KeyType   KeyType
	PublicKey []byte
}

type Keystore interface {
	Admin
	Reader
	Signer
	Encryptor
}

type key struct {
	keyType    KeyType
	privateKey internal.Raw
	createdAt  time.Time
}

func (k key) publicKey() []byte {
	switch k.keyType {
	case Ed25519:
		return ed25519.PublicKey(internal.Bytes(k.privateKey))
	case Secp256k1:
		return nil // TODO: derive public key
	case X25519:
		rv, err := curve25519.X25519(internal.Bytes(k.privateKey)[:], curve25519.Basepoint)
		// Shouldn't ever happen?
		if err != nil {
			panic(err)
		}
		var rvFixed [curve25519.PointSize]byte
		copy(rvFixed[:], rv)
		return rvFixed[:]
	default:
		// Could have a whole separate admin interface for
		// purely symmetric keys, but maybe not needed if we only ever use
		// asymmetric key exchange protocols like X25519 for encryption?
		return nil
	}
}

type keystore struct {
	mu sync.RWMutex
	// Keystore is the in memory keys.
	// Probably makes sense to have maps per key type and have actual
	// typed keys
	keystore map[string]key
	// Storage is used to persisting encrypted key material.
	storage  storage.Storage
	password string
}

func NewKeystore(storage storage.Storage, password string) (Keystore, error) {
	// Load the keystore from the storage.
	// TODO: create new empty keystore if storage empty (idempotent)
	ks, err := load(storage, password)
	if err != nil {
		return nil, fmt.Errorf("failed to load keystore: %w", err)
	}
	return &keystore{
		mu:       sync.RWMutex{},
		keystore: ks,
		storage:  storage,
		password: password,
	}, nil
}

func load(storage storage.Storage, password string) (map[string]key, error) {
	encryptedKeystore, err := storage.GetEncryptedKeystore(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted keystore: %w", err)
	}
	encryptedSecrets := gethkeystore.CryptoJSON{}
	err = json.Unmarshal(encryptedKeystore, &encryptedSecrets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal encrypted keystore: %w", err)
	}
	decryptedKeystore, err := gethkeystore.DecryptDataV3(encryptedSecrets, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}
	keystorepb := &serialization.Keystore{}
	err = proto.Unmarshal(decryptedKeystore, keystorepb)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal keystore: %w", err)
	}
	keystore := make(map[string]key)
	for _, k := range keystorepb.Keys {
		keystore[k.Name] = key{
			createdAt:  time.Unix(k.CreatedAt, 0),
			keyType:    KeyType(k.KeyType),
			privateKey: internal.NewRaw(k.PrivateKey),
		}
	}
	return keystore, nil
}

func save(storage storage.Storage, password string, keystore map[string]key) error {
	keystorepb := serialization.Keystore{
		Keys: make([]*serialization.Key, 0),
	}
	for name, k := range keystore {
		keystorepb.Keys = append(keystorepb.Keys, &serialization.Key{
			Name:       name,
			KeyType:    string(k.keyType),
			PrivateKey: internal.Bytes(k.privateKey),
			CreatedAt:  k.createdAt.Unix(),
		})
	}
	rawKeystore, err := proto.Marshal(&keystorepb)
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}
	// TODO: Could parameterize these.
	encryptedSecrets, err := gethkeystore.EncryptDataV3(rawKeystore, []byte(password), gethkeystore.StandardScryptN, gethkeystore.StandardScryptP)
	if err != nil {
		return fmt.Errorf("failed to encrypt keystore: %w", err)
	}
	encryptedSecretsBytes, err := json.Marshal(encryptedSecrets)
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted keystore: %w", err)
	}
	err = storage.PutEncryptedKeystore(context.Background(), encryptedSecretsBytes)
	if err != nil {
		return fmt.Errorf("failed to put encrypted keystore: %w", err)
	}
	return nil
}

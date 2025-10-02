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
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/serialization"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"google.golang.org/protobuf/proto"
)

type KeyType string

// KeyInfo is the information about a key in the keystore.
// Public key may be empty for non-asymmetric key types.
type KeyInfo struct {
	Name      string
	KeyType   KeyType
	PublicKey []byte
	Metadata  []byte
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
	metadata   []byte

	// Derived from private key on loading from storage.
	// Cached here for convenience.
	publicKey []byte
}

func publicKeyFromPrivateKey(privateKeyBytes internal.Raw, keyType KeyType) ([]byte, error) {
	switch keyType {
	case Ed25519:
		return ed25519.PublicKey(internal.Bytes(privateKeyBytes)), nil
	case EcdsaSecp256k1:
		// Here we use SEC1 (uncompressed) format for ECDSA public keys.
		// EVM addresses are derived from this format.
		// We use the geth crypto library for secp256k1 support
		// because stdlib doesn't support it.
		privateKey, err := gethcrypto.ToECDSA(internal.Bytes(privateKeyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to convert private key to ECDSA private key: %w", err)
		}
		pubKey := gethcrypto.FromECDSAPub(&privateKey.PublicKey)
		return pubKey, nil
	case X25519:
		rv, err := curve25519.X25519(internal.Bytes(privateKeyBytes)[:], curve25519.Basepoint)
		if err != nil {
			return nil, fmt.Errorf("failed to derive shared secret: %w", err)
		}
		return rv, nil
	default:
		// Some types may not have a public key.
		return []byte{}, nil
	}
}

type keystore struct {
	mu       sync.RWMutex
	keystore map[string]key
	storage  storage.Storage
	password string
}

func NewKeystore(storage storage.Storage, password string) (Keystore, error) {
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

	// If no data exists, return empty keystore
	if encryptedKeystore == nil || len(encryptedKeystore) == 0 {
		return make(map[string]key), nil
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
		publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(k.PrivateKey), KeyType(k.KeyType))
		if err != nil {
			return nil, fmt.Errorf("failed to get public key from private key: %w", err)
		}
		keystore[k.Name] = key{
			createdAt:  time.Unix(k.CreatedAt, 0),
			keyType:    KeyType(k.KeyType),
			privateKey: internal.NewRaw(k.PrivateKey),
			publicKey:  publicKey,
			metadata:   k.Metadata,
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
			Metadata:   k.metadata,
		})
	}
	rawKeystore, err := proto.Marshal(&keystorepb)
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}
	// TODO: Could parameterize these.
	// Scrypt supposedly impacts performance
	// significantly, let's benchmark that.
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

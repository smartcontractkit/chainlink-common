package keystore

import (
	"context"
	"crypto/rand"

	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore/serialization"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"golang.org/x/crypto/scrypt"
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
	KeyExchanger
}

type keystore struct {
	// Keystore is the in memory keys.
	keystore *serialization.Keystore
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
		keystore: ks,
		storage:  storage,
		password: password,
	}, nil
}

func load(storage storage.Storage, password string) (*serialization.Keystore, error) {
	encryptedKeystore, err := storage.GetEncryptedKeystore(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted keystore: %w", err)
	}
	encryptedSecrets := &EncryptedSecrets{}
	err = json.Unmarshal(encryptedKeystore, encryptedSecrets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal encrypted keystore: %w", err)
	}
	decryptedKeystore, err := decrypt(password, encryptedSecrets)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}
	keystore := &serialization.Keystore{}
	err = proto.Unmarshal(decryptedKeystore, keystore)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal keystore: %w", err)
	}
	return keystore, nil
}

func save(storage storage.Storage, password string, keystore *serialization.Keystore) error {
	rawKeystore, err := proto.Marshal(keystore)
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}
	encryptedSecrets, err := encrypt(password, rawKeystore)
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

const (
	scryptN = 32768
	scryptR = 8
	scryptP = 1
	keyLen  = 32
	version = 1
)

type EncryptedSecrets struct {
	IV             string `json:"iv"`              // Base64 encoded IV
	Ciphertext     string `json:"ciphertext"`      // Base64 encoded encrypted data
	Tag            string `json:"tag"`             // Base64 encoded authentication tag
	AssociatedData string `json:"associated_data"` // Associated data
	ScryptSalt     string `json:"scrypt_salt"`     // Base64 encoded scrypt salt
	Version        uint64 `json:"version"`         // Version number
}

func encrypt(password string, data []byte) (*EncryptedSecrets, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key, err := scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, data, []byte("keystore"))

	return &EncryptedSecrets{
		IV:             base64.StdEncoding.EncodeToString(iv),
		Ciphertext:     base64.StdEncoding.EncodeToString(ciphertext),
		Tag:            "", // GCM includes tag in ciphertext
		AssociatedData: "keystore",
		ScryptSalt:     base64.StdEncoding.EncodeToString(salt),
		Version:        version,
	}, nil
}

func decrypt(password string, encryptedSecrets *EncryptedSecrets) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(encryptedSecrets.ScryptSalt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(encryptedSecrets.IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedSecrets.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	key, err := scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertext, []byte(encryptedSecrets.AssociatedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

package file

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/scrypt"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/serialization"
)

const (
	scryptN = 32768
	scryptR = 8
	scryptP = 1
	keyLen  = 32
	version = 1
)

// EncryptedSecrets represents the JSON file format for the encrypted keystore
type EncryptedSecrets struct {
	IV             string `json:"iv"`              // Base64 encoded IV
	Ciphertext     string `json:"ciphertext"`      // Base64 encoded encrypted data
	Tag            string `json:"tag"`             // Base64 encoded authentication tag
	AssociatedData string `json:"associated_data"` // Associated data
	ScryptSalt     string `json:"scrypt_salt"`     // Base64 encoded scrypt salt
	Version        uint64 `json:"version"`         // Version number
}

type FileKeystore struct {
	password string
	filePath string
	cache    *serialization.Keystore
}

var _ keystore.Keystore = &FileKeystore{}

func NewFileKeystore(password string, filePath string) (*FileKeystore, error) {
	ks := &FileKeystore{
		password: password,
		filePath: filePath,
		cache:    nil,
	}

	if err := ks.loadKeystoreFromDisk(context.Background()); err != nil {
		return nil, err
	}

	return ks, nil
}

func (f *FileKeystore) loadKeystoreFromDisk(ctx context.Context) error {
	if _, err := os.Stat(f.filePath); os.IsNotExist(err) {
		f.cache = &serialization.Keystore{
			Ed25519Keys:   []*serialization.Ed25519Key{},
			Secp256K1Keys: []*serialization.Secp256K1Key{},
		}
		return nil
	}

	encryptedData, err := os.ReadFile(f.filePath)
	if err != nil {
		return fmt.Errorf("failed to read keystore file: %w", err)
	}

	encryptedSecrets := &EncryptedSecrets{}
	if err := json.Unmarshal(encryptedData, encryptedSecrets); err != nil {
		return fmt.Errorf("failed to unmarshal encrypted secrets: %w", err)
	}

	decryptedData, err := f.decrypt(encryptedSecrets)
	if err != nil {
		return fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	ks := &serialization.Keystore{}
	if err := proto.Unmarshal(decryptedData, ks); err != nil {
		return fmt.Errorf("failed to unmarshal keystore: %w", err)
	}

	f.cache = ks
	return nil
}

func (f *FileKeystore) saveKeystore(ctx context.Context, ks *serialization.Keystore) error {
	data, err := proto.Marshal(ks)
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}

	encryptedSecrets, err := f.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt keystore: %w", err)
	}

	encryptedData, err := json.MarshalIndent(encryptedSecrets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted secrets: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(f.filePath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(f.filePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	f.cache = ks
	return nil
}

func (f *FileKeystore) encrypt(data []byte) (*EncryptedSecrets, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key, err := scrypt.Key([]byte(f.password), salt, scryptN, scryptR, scryptP, keyLen)
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

func (f *FileKeystore) decrypt(encryptedSecrets *EncryptedSecrets) ([]byte, error) {
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

	key, err := scrypt.Key([]byte(f.password), salt, scryptN, scryptR, scryptP, keyLen)
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

func (f *FileKeystore) CreateKey(ctx context.Context, name string, keyType keystore.KeyType) (keystore.KeyInfo, error) {
	if f.keyExists(f.cache, name) {
		return keystore.KeyInfo{}, fmt.Errorf("key with name %s already exists", name)
	}

	var keyInfo keystore.KeyInfo
	now := time.Now().Unix()

	newKeystore := &serialization.Keystore{
		Ed25519Keys:   make([]*serialization.Ed25519Key, len(f.cache.Ed25519Keys)),
		Secp256K1Keys: make([]*serialization.Secp256K1Key, len(f.cache.Secp256K1Keys)),
	}

	copy(newKeystore.Ed25519Keys, f.cache.Ed25519Keys)
	copy(newKeystore.Secp256K1Keys, f.cache.Secp256K1Keys)

	switch keyType {
	case keystore.Ed25519:
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return keystore.KeyInfo{}, fmt.Errorf("failed to generate Ed25519 key: %w", err)
		}

		newKeystore.Ed25519Keys = append(newKeystore.Ed25519Keys, &serialization.Ed25519Key{
			Name:       name,
			PrivateKey: privateKey,
			CreatedAt:  now,
		})

		publicKey := privateKey.Public().(ed25519.PublicKey)
		keyInfo = keystore.KeyInfo{
			Name:      name,
			KeyType:   keyType,
			PublicKey: publicKey,
		}

	case keystore.Secp256k1:
		return keystore.KeyInfo{}, fmt.Errorf("secp256k1 key generation requires external library - not implemented yet")

	default:
		return keystore.KeyInfo{}, fmt.Errorf("unsupported key type: %s", keyType)
	}

	if err := f.saveKeystore(ctx, newKeystore); err != nil {
		return keystore.KeyInfo{}, err
	}

	return keyInfo, nil
}

func (f *FileKeystore) DeleteKey(ctx context.Context, name string) error {
	newKeystore := &serialization.Keystore{
		Ed25519Keys:   make([]*serialization.Ed25519Key, 0, len(f.cache.Ed25519Keys)),
		Secp256K1Keys: make([]*serialization.Secp256K1Key, 0, len(f.cache.Secp256K1Keys)),
	}

	keyFound := false

	for _, key := range f.cache.Ed25519Keys {
		if key.Name != name {
			newKeystore.Ed25519Keys = append(newKeystore.Ed25519Keys, key)
		} else {
			keyFound = true
		}
	}

	for _, key := range f.cache.Secp256K1Keys {
		if key.Name != name {
			newKeystore.Secp256K1Keys = append(newKeystore.Secp256K1Keys, key)
		} else {
			keyFound = true
		}
	}

	if !keyFound {
		return fmt.Errorf("key with name %s not found", name)
	}

	return f.saveKeystore(ctx, newKeystore)
}

func (f *FileKeystore) ImportKey(ctx context.Context, name string, keyType keystore.KeyType, data []byte) ([]byte, error) {
	if f.keyExists(f.cache, name) {
		return nil, fmt.Errorf("key with name %s already exists", name)
	}

	now := time.Now().Unix()

	newKeystore := &serialization.Keystore{
		Ed25519Keys:   make([]*serialization.Ed25519Key, len(f.cache.Ed25519Keys)),
		Secp256K1Keys: make([]*serialization.Secp256K1Key, len(f.cache.Secp256K1Keys)),
	}

	copy(newKeystore.Ed25519Keys, f.cache.Ed25519Keys)
	copy(newKeystore.Secp256K1Keys, f.cache.Secp256K1Keys)

	switch keyType {
	case keystore.Ed25519:
		if len(data) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid Ed25519 private key size: expected %d, got %d", ed25519.PrivateKeySize, len(data))
		}

		privateKey := ed25519.PrivateKey(data)
		publicKey := privateKey.Public().(ed25519.PublicKey)

		newKeystore.Ed25519Keys = append(newKeystore.Ed25519Keys, &serialization.Ed25519Key{
			Name:       name,
			PrivateKey: privateKey,
			CreatedAt:  now,
		})

		if err := f.saveKeystore(ctx, newKeystore); err != nil {
			return nil, err
		}

		return publicKey, nil

	case keystore.Secp256k1:
		return nil, fmt.Errorf("secp256k1 key import requires external library - not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

func (f *FileKeystore) GetKey(ctx context.Context, name string) (keystore.KeyInfo, error) {
	for _, key := range f.cache.Ed25519Keys {
		if key.Name == name {
			privateKey := ed25519.PrivateKey(key.PrivateKey)
			publicKey := privateKey.Public().(ed25519.PublicKey)

			return keystore.KeyInfo{
				Name:      key.Name,
				KeyType:   keystore.Ed25519,
				PublicKey: publicKey,
			}, nil
		}
	}

	for _, key := range f.cache.Secp256K1Keys {
		if key.Name == name {
			return keystore.KeyInfo{
				Name:      key.Name,
				KeyType:   keystore.Secp256k1,
				PublicKey: nil, // secp256k1 not implemented yet
			}, nil
		}
	}

	return keystore.KeyInfo{}, fmt.Errorf("key with name %s not found", name)
}

func (f *FileKeystore) ListKeys(ctx context.Context) ([]keystore.KeyInfo, error) {
	var keys []keystore.KeyInfo

	for _, key := range f.cache.Ed25519Keys {
		privateKey := ed25519.PrivateKey(key.PrivateKey)
		publicKey := privateKey.Public().(ed25519.PublicKey)

		keys = append(keys, keystore.KeyInfo{
			Name:      key.Name,
			KeyType:   keystore.Ed25519,
			PublicKey: publicKey,
		})
	}

	for _, key := range f.cache.Secp256K1Keys {
		keys = append(keys, keystore.KeyInfo{
			Name:      key.Name,
			KeyType:   keystore.Secp256k1,
			PublicKey: nil, // secp256k1 not implemented yet
		})
	}

	return keys, nil
}

func (f *FileKeystore) Sign(ctx context.Context, name string, data []byte) ([]byte, error) {
	for _, key := range f.cache.Ed25519Keys {
		if key.Name == name {
			privateKey := ed25519.PrivateKey(key.PrivateKey)
			return ed25519.Sign(privateKey, data), nil
		}
	}

	for _, key := range f.cache.Secp256K1Keys {
		if key.Name == name {
			return nil, fmt.Errorf("secp256k1 signing requires external library - not implemented yet")
		}
	}

	return nil, fmt.Errorf("key with name %s not found", name)
}

func (f *FileKeystore) Verify(ctx context.Context, name string, data []byte, signature []byte) (bool, error) {
	for _, key := range f.cache.Ed25519Keys {
		if key.Name == name {
			privateKey := ed25519.PrivateKey(key.PrivateKey)
			publicKey := privateKey.Public().(ed25519.PublicKey)
			return ed25519.Verify(publicKey, data, signature), nil
		}
	}

	for _, key := range f.cache.Secp256K1Keys {
		if key.Name == name {
			return false, fmt.Errorf("secp256k1 verification requires external library - not implemented yet")
		}
	}

	return false, fmt.Errorf("key with name %s not found", name)
}

func (f *FileKeystore) keyExists(ks *serialization.Keystore, name string) bool {
	for _, key := range ks.Ed25519Keys {
		if key.Name == name {
			return true
		}
	}
	for _, key := range ks.Secp256K1Keys {
		if key.Name == name {
			return true
		}
	}
	return false
}

package keystore

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/box"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

// Opaque error messages to prevent information leakage
var (
	ErrSharedSecretFailed = errors.New("shared secret derivation failed")
	ErrEncryptionFailed   = errors.New("encryption operation failed")
	ErrDecryptionFailed   = errors.New("decryption operation failed")
)

type EncryptRequest struct {
	RemoteKeyType KeyType
	RemotePubKey  []byte
	Data          []byte
}

type EncryptResponse struct {
	EncryptedData []byte
}

type DecryptRequest struct {
	KeyName       string
	EncryptedData []byte
}

type DecryptResponse struct {
	Data []byte
}

type DeriveSharedSecretRequest struct {
	KeyName      string
	RemotePubKey []byte
}

type DeriveSharedSecretResponse struct {
	SharedSecret []byte
}

const (
	// Maximum payload size for encrypt/decrypt operations (100kb)
	// Note just an initial limit, we may want to increase this in the future.
	MaxEncryptionPayloadSize = 100 * 1024
)

const (
	nonceSizeECDHP256  = 12
	ephPubSizeECDHP256 = 65
)

var (
	// Domain separation for HKDF-SHA256 based AES-GCM keys.
	infoAESGCM = []byte("keystore:ecdh-p256:aes-gcm:hkdf-sha256:v1")
)

// Encryptor is an interfaces for hybrid encryption (key exchange + encryption) operations.
type Encryptor interface {
	Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error)
	Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error)
	// DeriveSharedSecret: Derives a shared secret between the key specified
	// and the remote public key. WARNING: Using the shared secret should only be used directly in
	// cases where very custom encryption schemes are needed and you know
	// exactly what you are doing.
	DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error)
}

// UnimplementedEncryptor returns ErrUnimplemented for all Encryptor methods.
// Clients should embed this struct to ensure forward compatibility with changes to the Encryptor interface.
type UnimplementedEncryptor struct{}

func (UnimplementedEncryptor) Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error) {
	return EncryptResponse{}, fmt.Errorf("Encryptor.Encrypt: %w", ErrUnimplemented)
}

func (UnimplementedEncryptor) Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error) {
	return DecryptResponse{}, fmt.Errorf("Encryptor.Decrypt: %w", ErrUnimplemented)
}

func (UnimplementedEncryptor) DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error) {
	return DeriveSharedSecretResponse{}, fmt.Errorf("Encryptor.DeriveSharedSecret: %w", ErrUnimplemented)
}

func (k *keystore) Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error) {
	if len(req.Data) > MaxEncryptionPayloadSize {
		return EncryptResponse{}, ErrEncryptionFailed
	}

	switch req.RemoteKeyType {
	case X25519:
		encrypted, err := k.encryptX25519Anonymous(req.Data, req.RemotePubKey)
		if err != nil {
			return EncryptResponse{}, err
		}
		return EncryptResponse{
			EncryptedData: encrypted,
		}, nil
	case ECDH_P256:
		encrypted, err := k.encryptECDHP256Anonymous(req.Data, req.RemotePubKey)
		if err != nil {
			return EncryptResponse{}, err
		}
		return EncryptResponse{EncryptedData: encrypted}, nil
	default:
		return EncryptResponse{}, ErrEncryptionFailed
	}
}

func (k *keystore) Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if len(req.EncryptedData) == 0 || len(req.EncryptedData) > MaxEncryptionPayloadSize*2 {
		return DecryptResponse{}, ErrDecryptionFailed
	}

	key, ok := k.keystore[req.KeyName]
	if !ok {
		return DecryptResponse{}, ErrDecryptionFailed
	}

	switch key.keyType {
	case X25519:
		decrypted, err := k.decryptX25519Anonymous(req.EncryptedData, key.privateKey, key.publicKey)
		if err != nil {
			return DecryptResponse{}, err
		}
		return DecryptResponse{Data: decrypted}, nil
	case ECDH_P256:
		decrypted, err := k.decryptECDHP256Anonymous(req.EncryptedData, key.privateKey)
		if err != nil {
			return DecryptResponse{}, err
		}
		return DecryptResponse{Data: decrypted}, nil
	default:
		return DecryptResponse{}, ErrDecryptionFailed
	}
}

func (k *keystore) DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.KeyName]
	if !ok {
		return DeriveSharedSecretResponse{}, ErrSharedSecretFailed
	}

	switch key.keyType {
	case X25519:
		if len(req.RemotePubKey) != 32 {
			return DeriveSharedSecretResponse{}, ErrSharedSecretFailed
		}
		sharedSecret, err := curve25519.X25519(internal.Bytes(key.privateKey), req.RemotePubKey)
		if err != nil {
			return DeriveSharedSecretResponse{}, ErrSharedSecretFailed
		}
		return DeriveSharedSecretResponse{
			SharedSecret: sharedSecret,
		}, nil
	case ECDH_P256:
		// P-256 uncompressed public keys are 65 bytes (0x04 || x || y)
		if len(req.RemotePubKey) != 65 {
			return DeriveSharedSecretResponse{}, ErrSharedSecretFailed
		}
		curve := ecdh.P256()
		priv, err := curve.NewPrivateKey(internal.Bytes(key.privateKey))
		if err != nil {
			return DeriveSharedSecretResponse{}, ErrSharedSecretFailed
		}
		remotePub, err := curve.NewPublicKey(req.RemotePubKey)
		if err != nil {
			return DeriveSharedSecretResponse{}, ErrSharedSecretFailed
		}
		shared, err := priv.ECDH(remotePub)
		if err != nil {
			return DeriveSharedSecretResponse{}, ErrSharedSecretFailed
		}
		return DeriveSharedSecretResponse{SharedSecret: shared}, nil
	default:
		return DeriveSharedSecretResponse{}, ErrSharedSecretFailed
	}
}

// encryptX25519Anonymous performs X25519 anonymous encryption using NaCl box
func (k *keystore) encryptX25519Anonymous(data []byte, remotePubKey []byte) ([]byte, error) {
	if len(remotePubKey) != 32 {
		return nil, ErrEncryptionFailed
	}

	// Use SealAnonymous for anonymous encryption
	encrypted, err := box.SealAnonymous(nil, data, (*[32]byte)(remotePubKey), rand.Reader)
	if err != nil {
		return nil, ErrEncryptionFailed
	}
	return encrypted, nil
}

// decryptX25519Anonymous performs X25519 anonymous decryption using NaCl box
func (k *keystore) decryptX25519Anonymous(encryptedData []byte, privateKey internal.Raw, publicKey []byte) ([]byte, error) {
	if len(publicKey) != 32 {
		return nil, ErrDecryptionFailed
	}

	// Use OpenAnonymous for anonymous decryption
	decrypted, ok := box.OpenAnonymous(nil, encryptedData, (*[32]byte)(publicKey), (*[32]byte)(internal.Bytes(privateKey)))
	if !ok {
		return nil, ErrDecryptionFailed
	}
	if len(decrypted) == 0 {
		// box.OpenAnonymous will return a nil slice if the ciphertext is empty
		return []byte{}, nil
	}
	return decrypted, nil
}

// encryptECDHP256Anonymous performs ECDH-P256 anonymous encryption
// We follow the same general idea as box:
// 1. Generate an ephemeral key pair
// 2. Derive a shared secret using the ephemeral private key + recipient public key
// 3. Derive a nonce from the ephemeral public key + recipient public key: SHA-256(ephPubKey || recipientPubKey)[:12]
// 4. Derive an AES-256-GCM key from the shared secret and nonce
// 5. Encrypt the data with AES-GCM, including both nonce and ephemeral public key in AAD for complete authentication
// 6. Embed ephemeral public key and nonce in the result
func (k *keystore) encryptECDHP256Anonymous(data []byte, remotePubKey []byte) ([]byte, error) {
	curve := ecdh.P256()
	if len(remotePubKey) != 65 {
		return nil, ErrEncryptionFailed
	}

	// Generate ephemeral key pair for this encryption
	ephPriv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	// Remote public key must be on the P256 curve for the shared secret to work
	recipientPub, err := curve.NewPublicKey(remotePubKey)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	// Derive shared secret using ephemeral private key + recipient public key
	shared, err := ephPriv.ECDH(recipientPub)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	// Derive deterministic nonce from ephemeral public key + recipient public key
	// This is the same approach taken by box.SealAnonymous.
	nonce := deriveNonce(ephPriv.PublicKey().Bytes(), remotePubKey)

	// Derive AES-256-GCM key from shared secret and nonce
	derivedKey, err := deriveAESKeyFromSharedSecret(shared, nonce, infoAESGCM)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	// Encrypt with AES-GCM
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, ErrEncryptionFailed
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	// Include both nonce and ephemeral public key in AAD for complete authentication
	// AAD is [12 byte nonce] [65 byte ephemeral public key]
	ciphertext := gcm.Seal(nil, nonce, data, append(nonce[:], ephPriv.PublicKey().Bytes()...))

	// Embed ephemeral public key and nonce in the result
	// [12 byte nonce] [65 byte ephemeral public key] [AES-GCM ciphertext]
	result := encodeECDHP256Anonymous(nonce[:], ephPriv.PublicKey().Bytes(), ciphertext)
	return result, nil
}

func encodeECDHP256Anonymous(nonce []byte, ephPub []byte, ciphertext []byte) []byte {
	var result []byte
	result = append(result, nonce[:]...)
	result = append(result, ephPub...)
	result = append(result, ciphertext...)
	return result
}

func decodeECDHP256Anonymous(encryptedData []byte) ([]byte, []byte, []byte, error) {
	if len(encryptedData) < ephPubSizeECDHP256+nonceSizeECDHP256 {
		return nil, nil, nil, ErrDecryptionFailed
	}
	nonceBytes := encryptedData[:nonceSizeECDHP256]
	ephPubBytes := encryptedData[nonceSizeECDHP256 : nonceSizeECDHP256+ephPubSizeECDHP256]
	ciphertext := encryptedData[nonceSizeECDHP256+ephPubSizeECDHP256:]
	return nonceBytes, ephPubBytes, ciphertext, nil
}

// decryptECDHP256Anonymous performs ECDH-P256 anonymous decryption
func (k *keystore) decryptECDHP256Anonymous(encryptedData []byte, privateKey internal.Raw) ([]byte, error) {
	if len(encryptedData) < ephPubSizeECDHP256+nonceSizeECDHP256 {
		return nil, ErrDecryptionFailed
	}

	nonce, ephPubBytes, ciphertext, err := decodeECDHP256Anonymous(encryptedData)
	if err != nil {
		return nil, err
	}

	curve := ecdh.P256()
	ephPub, err := curve.NewPublicKey(ephPubBytes)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	priv, err := curve.NewPrivateKey(internal.Bytes(privateKey))
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	shared, err := priv.ECDH(ephPub)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	derivedKey, err := deriveAESKeyFromSharedSecret(shared, nonce[:], infoAESGCM)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// Include both nonce and ephemeral public key in AAD for complete authentication
	aad := append(nonce[:], ephPubBytes...)
	pt, err := gcm.Open(nil, nonce[:], ciphertext, aad)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	if len(pt) == 0 {
		// Return empty slice instead of nil for consistency
		return []byte{}, nil
	}
	return pt, nil
}

// deriveNonce creates a deterministic nonce from two public keys
func deriveNonce(pub1, pub2 []byte) []byte {
	h := sha256.New()
	h.Write(pub1)
	h.Write(pub2)
	return h.Sum(nil)[:nonceSizeECDHP256]
}

func deriveAESKeyFromSharedSecret(sharedSecret []byte, salt []byte, info []byte) ([]byte, error) {
	r := hkdf.New(sha256.New, sharedSecret, salt, info)
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("hkdf: %w", err)
	}
	return key, nil
}

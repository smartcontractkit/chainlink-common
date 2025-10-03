package keystore

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/box"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

type EncryptRequest struct {
	KeyName      string
	RemotePubKey []byte
	Data         []byte
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
	LocalKeyName string
	RemotePubKey []byte
}

type DeriveSharedSecretResponse struct {
	SharedSecret []byte
}

const (
	aesGCMNonceSize = 12
	hkdfSaltSize    = 16

	// ciphertext framing for EcdhP256 + HKDF-SHA256 + AES-GCM
	// [1B version=1][2B ephLen][ephPub][1B saltLen][salt][1B nonceLen][nonce][ciphertext]
	encVersionV1 byte = 1

	algP256HKDFAESGCM = "ecdh-p256+hkdf-sha256+aes-256-gcm"
)

func hkdfAESGCMKey(sharedSecret, salt, info []byte, keyLen int) ([]byte, error) {
	r := hkdf.New(sha256.New, sharedSecret, salt, info)
	key := make([]byte, keyLen)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("hkdf: %w", err)
	}
	return key, nil
}

type encAAD struct {
	V     byte   `json:"v"`
	Alg   string `json:"alg"`
	EPK   []byte `json:"epk"`
	Salt  []byte `json:"salt"`
	Nonce []byte `json:"nonce"`
}

type encEnvelope struct {
	V     byte   `json:"v"`
	Alg   string `json:"alg"`
	EPK   []byte `json:"epk"`
	Salt  []byte `json:"salt"`
	Nonce []byte `json:"nonce"`
	CT    []byte `json:"ct"`
}

// Encryptor is an interfaces for hybrid encryption (key exchange + encryption) operations.
// WARNING: Using the shared secret should only be used directly in
// cases where very custom encryption schemes are needed and you know
// exactly what you are doing.
type Encryptor interface {
	Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error)
	Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error)
	DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error)

	mustEmbedUnimplemented()
}

func (k *keystore) Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.KeyName]
	if !ok {
		return EncryptResponse{}, fmt.Errorf("key not found: %s", req.KeyName)
	}
	switch key.keyType {
	case X25519:
		if len(req.RemotePubKey) != 32 {
			return EncryptResponse{}, fmt.Errorf("remote public key must be 32 bytes for X25519")
		}
		encrypted, err := box.SealAnonymous(nil, req.Data, (*[32]byte)(req.RemotePubKey), rand.Reader)
		if err != nil {
			return EncryptResponse{}, fmt.Errorf("failed to encrypt data: %w", err)
		}
		return EncryptResponse{
			EncryptedData: encrypted,
		}, nil
	case EcdhP256:
		curve := ecdh.P256()
		if len(req.RemotePubKey) == 0 {
			return EncryptResponse{}, fmt.Errorf("remote public key required for EcdhP256")
		}
		recipientPub, err := curve.NewPublicKey(req.RemotePubKey)
		if err != nil {
			return EncryptResponse{}, fmt.Errorf("invalid P-256 public key: %w", err)
		}
		// Ephemeral key pair
		ephPriv, err := curve.GenerateKey(rand.Reader)
		if err != nil {
			return EncryptResponse{}, fmt.Errorf("failed to generate ephemeral key: %w", err)
		}
		shared, err := ephPriv.ECDH(recipientPub)
		if err != nil {
			return EncryptResponse{}, fmt.Errorf("ecdh failed: %w", err)
		}
		// Derive AES-256-GCM key
		salt := make([]byte, hkdfSaltSize)
		if _, err := rand.Read(salt); err != nil {
			return EncryptResponse{}, fmt.Errorf("salt generation failed: %w", err)
		}
		info := []byte("keystore:ecdh-p256:aes-gcm:hkdf-sha256:v1")
		aeadKey, err := hkdfAESGCMKey(shared, salt, info, 32)
		if err != nil {
			return EncryptResponse{}, err
		}
		block, err := aes.NewCipher(aeadKey)
		if err != nil {
			return EncryptResponse{}, fmt.Errorf("aes: %w", err)
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return EncryptResponse{}, fmt.Errorf("gcm: %w", err)
		}
		nonce := make([]byte, aesGCMNonceSize)
		if _, err := rand.Read(nonce); err != nil {
			return EncryptResponse{}, fmt.Errorf("nonce generation failed: %w", err)
		}
		ephPub := ephPriv.PublicKey().Bytes()
		head := encAAD{
			V:     encVersionV1,
			Alg:   algP256HKDFAESGCM,
			EPK:   ephPub,
			Salt:  salt,
			Nonce: nonce,
		}
		aadBytes, err := json.Marshal(head)
		if err != nil {
			return EncryptResponse{}, fmt.Errorf("aad marshal: %w", err)
		}
		ciphertext := gcm.Seal(nil, nonce, req.Data, aadBytes)
		env := encEnvelope{
			V:     encVersionV1,
			Alg:   algP256HKDFAESGCM,
			EPK:   ephPub,
			Salt:  salt,
			Nonce: nonce,
			CT:    ciphertext,
		}
		out, err := json.Marshal(env)
		if err != nil {
			return EncryptResponse{}, fmt.Errorf("envelope marshal: %w", err)
		}
		return EncryptResponse{EncryptedData: out}, nil
	default:
		return EncryptResponse{}, fmt.Errorf("unsupported key type: %s", key.keyType)
	}
}

func (k *keystore) Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.KeyName]
	if !ok {
		return DecryptResponse{}, fmt.Errorf("key not found: %s", req.KeyName)
	}
	switch key.keyType {
	case X25519:
		decrypted, ok := box.OpenAnonymous(nil, req.EncryptedData, (*[32]byte)(key.publicKey), (*[32]byte)(internal.Bytes(key.privateKey)))
		if !ok {
			return DecryptResponse{}, fmt.Errorf("failed to decrypt data")
		}
		return DecryptResponse{
			Data: decrypted,
		}, nil
	case EcdhP256:
		var env encEnvelope
		if err := json.Unmarshal(req.EncryptedData, &env); err != nil {
			return DecryptResponse{}, fmt.Errorf("envelope unmarshal: %w", err)
		}
		if env.V != encVersionV1 || env.Alg != algP256HKDFAESGCM {
			return DecryptResponse{}, fmt.Errorf("unsupported envelope version/alg")
		}
		curve := ecdh.P256()
		priv, err := curve.NewPrivateKey(internal.Bytes(key.privateKey))
		if err != nil {
			return DecryptResponse{}, fmt.Errorf("invalid P-256 private key: %w", err)
		}
		ephPub, err := curve.NewPublicKey(env.EPK)
		if err != nil {
			return DecryptResponse{}, fmt.Errorf("invalid P-256 ephemeral public key: %w", err)
		}
		shared, err := priv.ECDH(ephPub)
		if err != nil {
			return DecryptResponse{}, fmt.Errorf("ecdh failed: %w", err)
		}
		info := []byte("keystore:ecdh-p256:aes-gcm:hkdf-sha256:v1")
		aeadKey, err := hkdfAESGCMKey(shared, env.Salt, info, 32)
		if err != nil {
			return DecryptResponse{}, err
		}
		block, err := aes.NewCipher(aeadKey)
		if err != nil {
			return DecryptResponse{}, fmt.Errorf("aes: %w", err)
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return DecryptResponse{}, fmt.Errorf("gcm: %w", err)
		}
		aad := encAAD{V: env.V, Alg: env.Alg, EPK: env.EPK, Salt: env.Salt, Nonce: env.Nonce}
		aadBytes, err := json.Marshal(aad)
		if err != nil {
			return DecryptResponse{}, fmt.Errorf("aad marshal: %w", err)
		}
		pt, err := gcm.Open(nil, env.Nonce, env.CT, aadBytes)
		if err != nil {
			return DecryptResponse{}, fmt.Errorf("gcm open: %w", err)
		}
		return DecryptResponse{Data: pt}, nil
	default:
		return DecryptResponse{}, fmt.Errorf("unsupported key type: %s", key.keyType)
	}
}

func (k *keystore) DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.LocalKeyName]
	if !ok {
		return DeriveSharedSecretResponse{}, fmt.Errorf("key not found: %s", req.LocalKeyName)
	}
	switch key.keyType {
	case X25519:
		if len(req.RemotePubKey) != 32 {
			return DeriveSharedSecretResponse{}, fmt.Errorf("remote public key must be 32 bytes")
		}
		sharedSecret, err := curve25519.X25519(internal.Bytes(key.privateKey), req.RemotePubKey)
		if err != nil {
			return DeriveSharedSecretResponse{}, fmt.Errorf("failed to derive shared secret: %w", err)
		}
		return DeriveSharedSecretResponse{
			SharedSecret: sharedSecret,
		}, nil
	case EcdhP256:
		curve := ecdh.P256()
		priv, err := curve.NewPrivateKey(internal.Bytes(key.privateKey))
		if err != nil {
			return DeriveSharedSecretResponse{}, fmt.Errorf("invalid P-256 private key: %w", err)
		}
		remotePub, err := curve.NewPublicKey(req.RemotePubKey)
		if err != nil {
			return DeriveSharedSecretResponse{}, fmt.Errorf("invalid P-256 public key: %w", err)
		}
		shared, err := priv.ECDH(remotePub)
		if err != nil {
			return DeriveSharedSecretResponse{}, fmt.Errorf("ecdh failed: %w", err)
		}
		return DeriveSharedSecretResponse{SharedSecret: shared}, nil
	default:
		return DeriveSharedSecretResponse{}, fmt.Errorf("unsupported key type: %s", key.keyType)
	}
}

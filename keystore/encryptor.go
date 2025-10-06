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

// Opaque error messages to prevent information leakage
var (
	ErrEncryptionFailed = fmt.Errorf("encryption operation failed")
	ErrDecryptionFailed = fmt.Errorf("decryption operation failed")
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
	// 16 byte is the standard NIST recommended salt size for HKDF.
	hkdfSaltSize = 16
	// 1 byte is the version for the encryption envelope.
	encVersionV1 byte = 1
)

var (
	// Domain separation for HKDF-SHA256 based AES-GCM keys.
	infoAESGCM = []byte("keystore:ecdh-p256:aes-gcm:hkdf-sha256:v1")
)

type encHeader struct {
	Version            byte   `json:"version"`
	Alg                string `json:"alg"`
	EphemeralPublicKey []byte `json:"ephemeral_public_key"`
	Salt               []byte `json:"salt"`
	Nonce              []byte `json:"nonce"`
}

type encEnvelope struct {
	encHeader
	CipherText []byte `json:"ciphertext"`
}

// Encryptor is an interfaces for hybrid encryption (key exchange + encryption) operations.
// WARNING: Using the shared secret should only be used directly in
// cases where very custom encryption schemes are needed and you know
// exactly what you are doing.
type Encryptor interface {
	Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error)
	Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error)
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
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Validate request parameters without leaking information
	if req.KeyName == "" || len(req.Data) == 0 {
		return EncryptResponse{}, ErrEncryptionFailed
	}

	key, ok := k.keystore[req.KeyName]
	if !ok {
		// Don't leak key existence - return same error as other failures
		return EncryptResponse{}, ErrEncryptionFailed
	}

	switch key.keyType {
	case X25519:
		if len(req.RemotePubKey) != 32 {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		encrypted, err := box.SealAnonymous(nil, req.Data, (*[32]byte)(req.RemotePubKey), rand.Reader)
		if err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		return EncryptResponse{
			EncryptedData: encrypted,
		}, nil
	case EcdhP256:
		curve := ecdh.P256()
		if len(req.RemotePubKey) == 0 {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		// Remote public key must be on the P256 curve for the shared secret to work.
		recipientPub, err := curve.NewPublicKey(req.RemotePubKey)
		if err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		// Create an ephemeral keypair on the P256 curve used for encryption.
		ephPriv, err := curve.GenerateKey(rand.Reader)
		if err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		// The magic here is the the receipient can compute the same
		// shared secret because ephPriv*G*recipientPriv = ephPub*G.
		// This lets them derive the same ephemeral key used for encryption
		// so they can decrypt the ciphertext.
		shared, err := ephPriv.ECDH(recipientPub)
		if err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		// Derive AES-256-GCM key
		// The reason we do this is so that we can use symmetric encryption (more efficient)
		// This is part of any standard hybrid encryption scheme.
		// We include random salt to prevent rainbow table attacks (i.e. preventing
		// attackers from tracking a mapping of encryption data to plaintext)
		salt := make([]byte, hkdfSaltSize)
		if _, err := rand.Read(salt); err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		derivedKey, err := deriveAESKeyFromSharedSecret(shared, salt, infoAESGCM)
		if err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		block, err := aes.NewCipher(derivedKey)
		if err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err := rand.Read(nonce); err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		ephPub := ephPriv.PublicKey().Bytes()
		head := encHeader{
			Version:            encVersionV1,
			Alg:                EcdhP256.String(),
			EphemeralPublicKey: ephPub,
			Salt:               salt,
			Nonce:              nonce,
		}
		aadBytes, err := json.Marshal(head)
		if err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		// Critical to include the header parameters as additional authenticated data.
		// Prevents a MITM from changing the header.
		ciphertext := gcm.Seal(nil, nonce, req.Data, aadBytes)
		env := encEnvelope{
			encHeader:  head,
			CipherText: ciphertext,
		}
		out, err := json.Marshal(env)
		if err != nil {
			return EncryptResponse{}, ErrEncryptionFailed
		}
		return EncryptResponse{EncryptedData: out}, nil
	default:
		return EncryptResponse{}, ErrEncryptionFailed
	}
}

func (k *keystore) Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Validate request parameters without leaking information

	if req.KeyName == "" || len(req.EncryptedData) == 0 {
		return DecryptResponse{}, ErrDecryptionFailed
	}

	key, ok := k.keystore[req.KeyName]
	if !ok {
		// Don't leak key existence - return same error as other failures
		return DecryptResponse{}, ErrDecryptionFailed
	}

	switch key.keyType {
	case X25519:
		decrypted, ok := box.OpenAnonymous(nil, req.EncryptedData, (*[32]byte)(key.publicKey), (*[32]byte)(internal.Bytes(key.privateKey)))
		if !ok {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		return DecryptResponse{
			Data: decrypted,
		}, nil
	case EcdhP256:
		var env encEnvelope
		if err := json.Unmarshal(req.EncryptedData, &env); err != nil {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		if env.Version != encVersionV1 || env.Alg != string(EcdhP256) {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		curve := ecdh.P256()
		priv, err := curve.NewPrivateKey(internal.Bytes(key.privateKey))
		if err != nil {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		ephPub, err := curve.NewPublicKey(env.EphemeralPublicKey)
		if err != nil {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		shared, err := priv.ECDH(ephPub)
		if err != nil {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		derivedKey, err := deriveAESKeyFromSharedSecret(shared, env.Salt, infoAESGCM)
		if err != nil {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		block, err := aes.NewCipher(derivedKey)
		if err != nil {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		aad := encHeader{Version: env.Version, Alg: env.Alg, EphemeralPublicKey: env.EphemeralPublicKey, Salt: env.Salt, Nonce: env.Nonce}
		aadBytes, err := json.Marshal(aad)
		if err != nil {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		pt, err := gcm.Open(nil, env.Nonce, env.CipherText, aadBytes)
		if err != nil {
			return DecryptResponse{}, ErrDecryptionFailed
		}
		return DecryptResponse{Data: pt}, nil
	default:
		return DecryptResponse{}, ErrDecryptionFailed
	}
}

func (k *keystore) DeriveSharedSecret(ctx context.Context, req DeriveSharedSecretRequest) (DeriveSharedSecretResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Validate request parameters without leaking information
	if req.LocalKeyName == "" || len(req.RemotePubKey) == 0 {
		return DeriveSharedSecretResponse{}, ErrEncryptionFailed
	}

	key, ok := k.keystore[req.LocalKeyName]
	if !ok {
		// Don't leak key existence - return same error as other failures
		return DeriveSharedSecretResponse{}, ErrEncryptionFailed
	}

	switch key.keyType {
	case X25519:
		if len(req.RemotePubKey) != 32 {
			return DeriveSharedSecretResponse{}, ErrEncryptionFailed
		}
		sharedSecret, err := curve25519.X25519(internal.Bytes(key.privateKey), req.RemotePubKey)
		if err != nil {
			return DeriveSharedSecretResponse{}, ErrEncryptionFailed
		}
		return DeriveSharedSecretResponse{
			SharedSecret: sharedSecret,
		}, nil
	case EcdhP256:
		curve := ecdh.P256()
		priv, err := curve.NewPrivateKey(internal.Bytes(key.privateKey))
		if err != nil {
			return DeriveSharedSecretResponse{}, ErrEncryptionFailed
		}
		remotePub, err := curve.NewPublicKey(req.RemotePubKey)
		if err != nil {
			return DeriveSharedSecretResponse{}, ErrEncryptionFailed
		}
		shared, err := priv.ECDH(remotePub)
		if err != nil {
			return DeriveSharedSecretResponse{}, ErrEncryptionFailed
		}
		return DeriveSharedSecretResponse{SharedSecret: shared}, nil
	default:
		return DeriveSharedSecretResponse{}, ErrEncryptionFailed
	}
}

func deriveAESKeyFromSharedSecret(sharedSecret []byte, salt []byte, info []byte) ([]byte, error) {
	r := hkdf.New(sha256.New, sharedSecret, salt, info)
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("hkdf: %w", err)
	}
	return key, nil
}

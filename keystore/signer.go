package keystore

import (
	"context"
	"fmt"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ed25519"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

type SignRequest struct {
	KeyName string
	Data    []byte
}

type SignResponse struct {
	Signature []byte
}

type VerifyRequest struct {
	KeyName   string
	Data      []byte
	Signature []byte
}

type VerifyResponse struct {
	Valid bool
}

type Signer interface {
	Sign(ctx context.Context, req SignRequest) (SignResponse, error)
	Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error)
}

// UnimplementedSigner returns ErrUnimplemented for all Signer methods.
type UnimplementedSigner struct{}

func (UnimplementedSigner) Sign(ctx context.Context, req SignRequest) (SignResponse, error) {
	return SignResponse{}, fmt.Errorf("Signer.Sign: %w", ErrUnimplemented)
}

func (UnimplementedSigner) Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error) {
	return VerifyResponse{}, fmt.Errorf("Signer.Verify: %w", ErrUnimplemented)
}

func (k *keystore) Sign(ctx context.Context, req SignRequest) (SignResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.KeyName]
	if !ok {
		return SignResponse{}, fmt.Errorf("key not found: %s", req.KeyName)
	}
	switch key.keyType {
	case Ed25519:
		privateKey := ed25519.PrivateKey(internal.Bytes(key.privateKey))
		signature := ed25519.Sign(privateKey, req.Data)
		return SignResponse{
			Signature: signature,
		}, nil
	case ECDSA_S256:
		if len(req.Data) != 32 {
			return SignResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256")
		}
		privateKey, err := gethcrypto.ToECDSA(internal.Bytes(key.privateKey))
		if err != nil {
			return SignResponse{}, fmt.Errorf("failed to convert private key to ECDSA private key: %w", err)
		}
		signature, err := gethcrypto.Sign(req.Data, privateKey)
		if err != nil {
			return SignResponse{}, err
		}
		return SignResponse{
			Signature: signature,
		}, nil
	default:
		return SignResponse{}, fmt.Errorf("unsupported key type: %s", key.keyType)
	}
}

func (k *keystore) Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.KeyName]
	if !ok {
		return VerifyResponse{}, fmt.Errorf("key not found: %s", req.KeyName)
	}
	switch key.keyType {
	case Ed25519:
		if len(req.Signature) != 64 {
			return VerifyResponse{}, fmt.Errorf("signature must be 64 bytes for Ed25519")
		}
		publicKey := ed25519.PublicKey(key.publicKey)
		signature := ed25519.Verify(publicKey, req.Data, req.Signature)
		return VerifyResponse{
			Valid: signature,
		}, nil
	case ECDSA_S256:
		if len(req.Data) != 32 {
			return VerifyResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256")
		}
		if len(req.Signature) != 64 {
			return VerifyResponse{}, fmt.Errorf("signature must be 64 bytes for ECDSA_S256")
		}
		publicKey, err := gethcrypto.UnmarshalPubkey(key.publicKey)
		if err != nil {
			return VerifyResponse{}, fmt.Errorf("failed to unmarshal public key: %w", err)
		}
		signature := gethcrypto.VerifySignature(gethcrypto.FromECDSAPub(publicKey), req.Data, req.Signature)
		if err != nil {
			return VerifyResponse{}, err
		}
		return VerifyResponse{
			Valid: signature,
		}, nil
	default:
		return VerifyResponse{}, fmt.Errorf("unsupported key type: %s", key.keyType)
	}
}

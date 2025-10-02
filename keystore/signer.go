package keystore

import (
	"context"
)

const (
	// Digital signature key types.
	Ed25519        KeyType = "ed25519"
	EcdsaSecp256k1 KeyType = "ecdsa-secp256k1"
)

type SignRequest struct {
	Name string
	Data []byte
}

type SignResponse struct {
	Signature []byte
}

type VerifyRequest struct {
	Name      string
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

// TODO: Signer implementation.

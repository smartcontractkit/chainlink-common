package keystore

import "context"

const (
	// Digital signature key types.
	Ed25519   KeyType = "ed25519"
	Secp256k1 KeyType = "secp256k1"
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

func (k *keystore) Sign(ctx context.Context, req SignRequest) (SignResponse, error) {
	return SignResponse{}, nil
}

func (k *keystore) Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error) {
	return VerifyResponse{}, nil
}

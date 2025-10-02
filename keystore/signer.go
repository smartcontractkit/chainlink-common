package keystore

import (
	"context"
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

// TODO: Signer implementation.
func (k *keystore) Sign(ctx context.Context, req SignRequest) (SignResponse, error) {
	return SignResponse{}, nil
}

func (k *keystore) Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error) {
	return VerifyResponse{}, nil
}

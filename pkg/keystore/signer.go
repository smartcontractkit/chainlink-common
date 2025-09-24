package keystore

import "context"

// Request/Response types for Signer interface
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

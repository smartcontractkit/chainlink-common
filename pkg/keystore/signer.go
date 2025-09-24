package keystore

import (
	"context"
	"fmt"

	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/internal"
)

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

// Sign signs the data with the key.
// Note this provides no specific safe guards.
func (k *keystore) Sign(ctx context.Context, req SignRequest) (SignResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.Name]
	if !ok {
		return SignResponse{}, fmt.Errorf("key not found: %s", req.Name)
	}
	switch key.keyType {
	case Secp256k1:
		var privateKey *ecdsa.PrivateKey
		d := big.NewInt(0).SetBytes(internal.Bytes(key.privateKey))
		privateKey.D = d
		privateKey.PublicKey.Curve = crypto.S256()
		privateKey.PublicKey.X, privateKey.PublicKey.Y = crypto.S256().ScalarBaseMult(d.Bytes())
		signature, err := crypto.Sign(req.Data, privateKey)
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
	return VerifyResponse{}, nil
}

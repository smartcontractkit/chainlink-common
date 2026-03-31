package aptos

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// AptosLoopKeystore extends core.Keystore for Aptos-specific key access.
//
// Accounts() returns Aptos account addresses (not public keys), consistent
// with the EVM keystore pattern. The account address is stable across
// authentication key rotation.
//
// GetPublicKey() provides the Ed25519 public key for a given account address.
// This is required to build Aptos transaction signatures, which must include
// both the public key bytes and the signature bytes.
type AptosLoopKeystore interface {
	core.Keystore
	// GetPublicKey returns the hex-encoded Ed25519 public key for the given
	// account address. Returns an error if no key is found for the address.
	GetPublicKey(ctx context.Context, accountAddr string) (string, error)
}

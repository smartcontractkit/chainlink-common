package nodeauth

import (
	"encoding/hex"
	"fmt"
)

// To avoid circular dependencies, we replicate the Signer interface from chainlink/core/services/p2p/types.
// Signer interface: https://github.com/smartcontractkit/chainlink/blob/592b4fbbe8b4d623fca65a0c8e8bba31fe651ed8/core/services/p2p/types/types.go
// Each node will implement the Signer interface.
type Signer interface {
	Sign(msg []byte) (signature []byte, err error)
}

// NodeJWTSigningMethod is an implementation of the jwt.SigningMethod interface that allows nodes to use their native signing mechanisms (e.g., ECDSA) to generate JWT tokens.
// All nodes will implement the Signer interface.
// This is a wrapper around the Signer interface to be compatible with the jwt.SigningMethod interface.
// jwt.SigningMethod interface: https://pkg.go.dev/github.com/golang-jwt/jwt/v5#SigningMethod
type NodeJWTSigningMethod struct{}

func (d *NodeJWTSigningMethod) Verify(signingString string, signature []byte, key interface{}) error {
	// Verification is handled server-side through blockchain signature recovery
	// This method is required by the jwt.SigningMethod interface but not used in our client flow
	return fmt.Errorf("verification not supported - handled by server-side signature recovery")
}

// Sign() uses the node's signer implementation to sign and generate a JWT token.
// Each node must implement the Signer interface.
func (d *NodeJWTSigningMethod) Sign(signingString string, key interface{}) ([]byte, error) {
	signer, ok := key.(Signer)
	if !ok {
		return nil, fmt.Errorf("key must implement Signer interface, got %T", key)
	}

	messageBytes := []byte(signingString)

	// Call the node's signer implementation
	signature, err := signer.Sign(messageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT with node signer: %w", err)
	}

	hexSignature := hex.EncodeToString(signature)
	return []byte(hexSignature), nil
}

func (d *NodeJWTSigningMethod) Alg() string {
	return "ES256K" // ECDSA algorithm on secp256k1 curve (Ethereum-style)
}

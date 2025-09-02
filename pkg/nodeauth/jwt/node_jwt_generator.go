package jwt

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/types"
	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/utils"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// NodeJWTGenerator implements the JWTGenerator interface.
type NodeJWTGenerator struct {
	signer    *core.Ed25519Signer // The Ed25519Signer to sign the JWT (no private key exposure)
	csaPubKey ed25519.PublicKey   // the ed25519 public key (signature key's counterpart) of the node to verify the JWT's signature.
}

// NewNodeJWTGenerator creates a new node JWT generator
func NewNodeJWTGenerator(signer *core.Ed25519Signer, csaPubKey ed25519.PublicKey) *NodeJWTGenerator {
	return &NodeJWTGenerator{
		signer:    signer,
		csaPubKey: csaPubKey,
	}
}

// CreateJWTForRequest creates a JWT token for the given request
func (m *NodeJWTGenerator) CreateJWTForRequest(req any) (string, error) {
	if m.signer == nil {
		return "", fmt.Errorf("no signer configured")
	}

	// Create request digest for integrity
	digest := utils.CalculateRequestDigest(req)

	// Create JWT claims
	now := time.Now()
	claims := types.NodeJWTClaims{
		PublicKey: hex.EncodeToString(m.csaPubKey), // PublicKey: Node's public key to proof JWT's signature.
		Digest:    digest,                          // Digest: Request integrity hash
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hex.EncodeToString(m.csaPubKey), // Issuer: Node's CSA Public Key  // TODO: change to DON ID if node is aware of its DON ID
			Subject:   hex.EncodeToString(m.csaPubKey), // Subject: Node's CSA Public Key for on-chain verification of node-DON relationship.
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Create token with claims using jwt.SigningMethodEdDSA(built-in EdDSA signing method)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	return token.SignedString(m.signer)
}

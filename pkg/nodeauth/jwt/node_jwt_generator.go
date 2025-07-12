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
	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

// NodeJWTGenerator implements the JWTGenerator interface.
type NodeJWTGenerator struct {
	environment types.EnvironmentName
	signer      *core.Ed25519Signer // The Ed25519Signer to sign the JWT (no private key exposure)
	csaPubKey   ed25519.PublicKey   // the ed25519 public key (signature key's counterpart) of the node to verify the JWT's signature.
	p2pId       p2ptypes.PeerID     // node p2pId to identify this node on-chain
}

// NewNodeJWTGenerator creates a new node JWT generator
func NewNodeJWTGenerator(signer *core.Ed25519Signer, csaPubKey ed25519.PublicKey, p2pId p2ptypes.PeerID, environment types.EnvironmentName) *NodeJWTGenerator {
	return &NodeJWTGenerator{
		environment: environment,
		signer:      signer,
		csaPubKey:   csaPubKey,
		p2pId:       p2pId,
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
		P2PId:       m.p2pId.String(),                // P2PId: Node's on-chain P2P ID for on-chain verification of node-DON relationship.
		PublicKey:   hex.EncodeToString(m.csaPubKey), // PublicKey: Node's public key to proof JWT's signature.
		Environment: string(m.environment),           // Environment: Environment for which the JWT token is generated.
		Digest:      digest,                          // Digest: Request integrity hash
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.p2pId.String(), // Issuer: Node's P2P ID  // TODO: change to DON ID if node is aware of its DON ID
			Subject:   m.p2pId.String(), // Subject: Node's P2P ID for on-chain verification of node-DON relationship.
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Create token with claims using jwt.SigningMethodEdDSA(built-in EdDSA signing method)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	return token.SignedString(m.signer)
}

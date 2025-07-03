package nodeauth

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/utils"
)

// NodeJWTGenerator implements the JWTGenerator interface.
// It is used to generate JWT tokens for node-initiated requests.
type NodeJWTGenerator struct {
	environment EnvironmentName
	privateKey  ed25519.PrivateKey // The ed25519 private key(csa) of the node to sign the JWT.
	publicKey   ed25519.PublicKey  // the ed25519 public key (csa counterpart) of the node to verify the JWT's signature.
	p2pId       ed25519.PublicKey  // the ed25519 public key (node p2pId to identify this node on-chain)

}

// NewNodeJWTGenerator creates a new node JWT generator
func NewNodeJWTGenerator(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, p2pId ed25519.PublicKey, environment EnvironmentName) *NodeJWTGenerator {
	return &NodeJWTGenerator{
		environment: environment,
		privateKey:  privateKey,
		publicKey:   publicKey,
		p2pId:       p2pId,
	}
}

// CreateJWTForRequest creates a JWT token for the given request
func (m *NodeJWTGenerator) CreateJWTForRequest(req any) (string, error) {
	if m.privateKey == nil {
		return "", fmt.Errorf("no private key configured")
	}

	// Create request digest for integrity
	digest := utils.CalculateRequestDigest(req)

	// Create JWT claims
	now := time.Now()
	claims := NodeJWTClaims{
		P2PId:       hex.EncodeToString(m.p2pId),     // P2PId: Node's on-chain P2P ID for on-chain verification of node-DON relationship.
		PublicKey:   hex.EncodeToString(m.publicKey), // PublicKey: Node's public key to proof JWT's signature.
		Environment: string(m.environment),           // Environment: Environment for which the JWT token is generated.
		Digest:      digest,                          // Digest: Request integrity hash
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hex.EncodeToString(m.p2pId), // Issuer: Node's P2P ID  // TODO: change to DON ID if node is aware of its DON ID
			Subject:   hex.EncodeToString(m.p2pId), // Subject: Node's P2P ID for on-chain verification of node-DON relationship.
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Create token with claims using Ed25519(EdDSASigningMethod) signing method
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	// Sign the token with the private key
	return token.SignedString(m.privateKey)
}

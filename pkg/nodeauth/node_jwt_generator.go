package nodeauth

import (
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
	signer      Signer   // Each node must implement the Signer interface in node_jwt_signer.go.
	p2pId       [32]byte // p2pId is the on-chain P2P ID of the node.
	publicKey   [32]byte // Node's public key to verify JWT's signature.
}

// NewNodeJWTGenerator creates a new node JWT generator
func NewNodeJWTGenerator(signer Signer, p2pId [32]byte, publicKey [32]byte) *NodeJWTGenerator {
	return &NodeJWTGenerator{
		signer:    signer,
		p2pId:     p2pId,
		publicKey: publicKey,
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
	claims := NodeJWTClaims{
		P2PId:       hex.EncodeToString(m.p2pId[:]),     // P2PId: Node's on-chain P2P ID for on-chain verification of node-DON relationship.
		PublicKey:   hex.EncodeToString(m.publicKey[:]), // PublicKey: Node's public key to proof JWT's signature.
		Environment: string(m.environment),              // Environment: Environment for which the JWT token is generated.
		Digest:      digest,                             // Digest: Request integrity hash
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hex.EncodeToString(m.p2pId[:]),    // Issuer: Node's P2P ID  // TODO: change to DON ID if node is aware of its DON ID
			Subject:   hex.EncodeToString(m.p2pId[:]),    // Subject: Node's P2P ID for on-chain verification of node-DON relationship.
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Create token with claims (using node-specific signing method)
	token := jwt.NewWithClaims(&NodeJWTSigningMethod{}, claims)

	// Sign the token - pass the signer as the key
	return token.SignedString(m.signer)
}

package nodeauth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	workflowJWTExpiration                            = 5 * time.Minute
	workflowDONType                                  = "workflowDON"
	EnvironmentNameProductionMainnet EnvironmentName = "production_mainnet"
	EnvironmentNameProductionTestnet EnvironmentName = "production_testnet"
)

// EnvironmentName represents the environment for which the JWT token is generated
type EnvironmentName string

// JWTManager handles JWT token creation and management for workflow clients
type JWTManager interface {
	// CreateJWTForRequest creates a JWT token for the given request
	CreateJWTForRequest(req any) (string, error)
}

type NodeJWTManager struct {
	environment EnvironmentName
	signer      Signer   // Each node must implement the Signer interface in node_jwt_signer.go.
	p2pId       [32]byte // p2pId is the on-chain P2P ID of the node.
	publicKey   [32]byte
	// Node's public key to proof JWT's signature.
	// Should be ABI encoded version of the node's address:
	// Shsould be the signer field in : https://github.com/smartcontractkit/chainlink-evm/blob/develop/contracts/src/v0.8/workflow/dev/v2/CapabilitiesRegistry.sol#L65
}

// NodeJWTClaims represents the JWT claims payload for node-initiated requests.
type NodeJWTClaims struct {
	P2PId       string `json:"p2pId"`
	PublicKey   string `json:"public_key"`
	Environment string `json:"environment"`
	Digest      string `json:"digest"`
	jwt.RegisteredClaims
}

// NewNodeJWTManager creates a new node JWT manager
func NewNodeJWTManager(signer Signer, p2pId [32]byte, publicKey [32]byte) *NodeJWTManager {
	return &NodeJWTManager{
		signer:    signer,
		p2pId:     p2pId,
		publicKey: publicKey,
	}
}

// CreateJWTForRequest creates a JWT token for the given request
// Note: workflowId is NOT included in JWT claims - it's included in the GRPC request payload struct.
func (m *NodeJWTManager) CreateJWTForRequest(req any) (string, error) {
	if m.signer == nil {
		return "", fmt.Errorf("no signer configured")
	}

	// Create request digest for integrity
	digest := m.DigestFromRequest(req)

	// Create JWT claims
	now := time.Now()
	claims := NodeJWTClaims{
		P2PId:       hex.EncodeToString(m.p2pId[:]),     // P2PId: Node's on-chain P2P ID for on-chain verification of node-DON relationship.
		PublicKey:   hex.EncodeToString(m.publicKey[:]), // PublicKey: Node's public key to proof JWT's signature.
		Environment: string(m.environment),              // Environment: Environment for which the JWT token is generated.
		Digest:      digest,                             // Digest: Request integrity hash
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    workflowDONType,                // Issuer: Node's DON ID  // TODO: change to DON ID if node is aware of its DON ID
			Subject:   hex.EncodeToString(m.p2pId[:]), // Subject: Node's P2P ID for on-chain verification of node-DON relationship.
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Create token with claims (using node-specific signing method)
	token := jwt.NewWithClaims(&NodeJWTSigningMethod{}, claims)

	// Sign the token - pass the signer as the key
	return token.SignedString(m.signer)
}

// DigestFromRequest creates a SHA256 digest of the request for integrity verification
func (m *NodeJWTManager) DigestFromRequest(req any) string {
	// Create canonical string representation
	var canonical string
	if s, ok := req.(fmt.Stringer); ok {
		canonical = s.String()
	} else {
		canonical = fmt.Sprintf("%v", req)
	}

	// Hash and encode as hex
	hash := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(hash[:])
}

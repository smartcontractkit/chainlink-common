package jwt

import (
	"context"
	"crypto/ed25519"

	nodeauthtypes "github.com/smartcontractkit/chainlink-common/pkg/nodeauth/types"
)

// ---------- JWT Generator Interfaces ----------
// JWTGenerator handles JWT token creation.
type JWTGenerator interface {
	// CreateJWTForRequest creates a JWT token for the given request
	CreateJWTForRequest(req any) (string, error)
}

// ---------- JWT Authenticator - Related Interfaces ----------
// JWTAuthenticator handles JWT token authentication.
type JWTAuthenticator interface {
	// AuthenticateJWT authenticates the JWT token for the given request and return the JWT claims.
	// If the JWT token is invalid, the function will return nil claims and error.
	AuthenticateJWT(ctx context.Context, tokenString string, originalRequest any) (bool, *nodeauthtypes.NodeJWTClaims, error)
}

// NodeAuthProvider interface for node <-> DON auth provider
// Each service that uses NodeJWTAuthenticator must provide an implementation for this interface.
type NodeAuthProvider interface {

	// IsNodePubKeyTrusted checks if a node's public key is trusted
	// Usually, this is done by checking the node aginst DON's on-chain topology.
	// The check can be done aginst on-chain contracts or cache, depending on the each service's implementation.
	IsNodePubKeyTrusted(ctx context.Context, publicKey ed25519.PublicKey) (bool, error)
}

package jwt

import (
	"context"
	"crypto/ed25519"

	"github.com/golang-jwt/jwt/v5"
)

// ---------- JWT Payload - Related Types ----------
// NodeJWTClaims represents the JWT claims payload for node-initiated requests.
type NodeJWTClaims struct {
	P2PId       string `json:"p2pId" validate:"required"`
	PublicKey   string `json:"public_key" validate:"required"`
	Environment string `json:"environment" validate:"required"`
	Digest      string `json:"digest" validate:"required"`
	jwt.RegisteredClaims
}

// EnvironmentName represents the environment for which the JWT token is generated
type EnvironmentName string

// ---------- JWT Generator Interfaces ----------
// JWTGenerator handles JWT token creation.
type JWTGenerator interface {
	// CreateJWTForRequest creates a JWT token for the given request
	CreateJWTForRequest(req any) (string, error)
}

// ---------- JWT Authenticator - Related Interfaces ----------
// JWTAuthenticator handles JWT token authentication.
type JWTAuthenticator interface {
	// AuthenticateJWT authenticates the JWT token for the given request
	AuthenticateJWT(ctx context.Context, tokenString string, originalRequest any) (bool, error)
}

// NodeTopologyProvider interface for node <-> DON topology provider
// Each service that uses NodeJWTAuthenticator must provide an implementation for this interface.
type NodeTopologyProvider interface {

	// IsNodePubKeyTrusted checks if a node's public key is trusted
	// Usually, this is done by checking the node aginst DON's on-chain topology.
	// The check can be done aginst on-chain contracts or cache, depending on the each service's implementation.
	IsNodePubKeyTrusted(ctx context.Context, p2pId ed25519.PublicKey, publicKey ed25519.PublicKey) (bool, error)
}

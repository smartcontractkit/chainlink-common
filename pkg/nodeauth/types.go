package nodeauth

import (
	"context"

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

// ---------- JWT Validator - Related Interfaces ----------
// JWTValidator handles JWT token validation.
type JWTValidator interface {
	// ValidateJWT validates the JWT token for the given request
	ValidateJWT(ctx context.Context, tokenString string, originalRequest any) (bool, error)
}

// NodeTopologyProvider interface for node <-> DON topology authorization check.
// Each service that uses NodeJWTValidator must provide an implementation for this interface.
type NodeTopologyProvider interface {
	// IsNodeAuthorized checks if a node is authorized
	// Usually, this is done by checking the node aginst DON's on-chain topology.
	// The check can be done aginst on-chain contracts or cache, depending on the each service's implementation.
	IsNodeAuthorized(ctx context.Context, p2pId string, publicKey [32]byte) (bool, error)
}



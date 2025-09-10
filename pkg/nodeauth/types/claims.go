package types

import (
	"github.com/golang-jwt/jwt/v5"
)

// NodeJWTClaims represents the JWT claims payload for node-initiated requests.
type NodeJWTClaims struct {
	PublicKey string `json:"public_key" validate:"required"`
	Digest    string `json:"digest" validate:"required"`
	jwt.RegisteredClaims
}

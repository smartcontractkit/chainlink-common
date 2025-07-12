package types

import (
	"github.com/golang-jwt/jwt/v5"
)

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

const (
	EnvironmentNameProductionMainnet EnvironmentName = "production_mainnet"
	EnvironmentNameProductionTestnet EnvironmentName = "production_testnet"
)

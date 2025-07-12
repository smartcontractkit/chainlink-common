package jwt

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/types"
	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/utils"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

// Test request type
type mockRequest struct {
	Field string
}

func (r mockRequest) String() string {
	return r.Field
}

// Helper function to create test signer and keys
func createTestSigner() (*core.Ed25519Signer, ed25519.PublicKey, p2ptypes.PeerID) {
	// Generate a private key for signing
	csaPubKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Failed to generate Ed25519 key pair: " + err.Error())
	}

	// Create ed25519 signer from the mock node's csa private key
	signFn := func(ctx context.Context, account string, data []byte) (signed []byte, err error) {
		return ed25519.Sign(privateKey, data), nil
	}

	signer, err := core.NewEd25519Signer(hex.EncodeToString(csaPubKey), signFn)
	if err != nil {
		panic("Failed to create Ed25519Signer: " + err.Error())
	}

	// Generate a separate public key for p2pId
	p2pIdKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Failed to generate Ed25519 p2pId: " + err.Error())
	}

	// Create PeerID from ed25519 public key
	p2pId, err := p2ptypes.PeerIDFromPublicKey(p2pIdKey)
	if err != nil {
		panic("Failed to create PeerID from public key: " + err.Error())
	}

	return signer, csaPubKey, p2pId
}

func TestNodeJWTGenerator_CreateJWTForRequest(t *testing.T) {
	// prepare test data
	signer, csaPubKey, p2pId := createTestSigner()
	req := mockRequest{Field: "test request"}

	jwtGenerator := NewNodeJWTGenerator(signer, csaPubKey, p2pId, types.EnvironmentNameProductionTestnet)

	jwtToken, err := jwtGenerator.CreateJWTForRequest(req)
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	// verify expected JWT claim values
	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &types.NodeJWTClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(*types.NodeJWTClaims)
	require.True(t, ok)
	assert.Equal(t, p2pId.String(), claims.Issuer)

	expectedP2PIdStr := p2pId.String()
	expectedCSAPubKeyHex := hex.EncodeToString(csaPubKey)

	assert.Equal(t, expectedP2PIdStr, claims.Subject)
	assert.Equal(t, expectedP2PIdStr, claims.P2PId)
	assert.Equal(t, expectedCSAPubKeyHex, claims.PublicKey)

	expectedDigest := utils.CalculateRequestDigest(req)
	assert.Equal(t, expectedDigest, claims.Digest)

	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)

	assert.Equal(t, p2pId.String(), claims.P2PId, "P2P ID should match string representation")

	decodedCSAPubKey, err := hex.DecodeString(claims.PublicKey)
	require.NoError(t, err)
	assert.Equal(t, csaPubKey, ed25519.PublicKey(decodedCSAPubKey), "Decoded CSA public key should match original")
}

func TestNodeJWTGenerator_DigestTampering(t *testing.T) {
	// prepare test data
	signer, csaPubKey, p2pId := createTestSigner()
	originalReq := mockRequest{Field: "original"}
	tamperedReq := mockRequest{Field: "tampered"}

	jwtGenerator := NewNodeJWTGenerator(signer, csaPubKey, p2pId, types.EnvironmentNameProductionTestnet)

	jwtToken, err := jwtGenerator.CreateJWTForRequest(originalReq)
	require.NoError(t, err)

	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &types.NodeJWTClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(*types.NodeJWTClaims)
	require.True(t, ok)

	originalDigest := utils.CalculateRequestDigest(originalReq)
	tamperedDigest := utils.CalculateRequestDigest(tamperedReq)

	assert.Equal(t, originalDigest, claims.Digest)
	assert.NotEqual(t, tamperedDigest, claims.Digest, "Digest should be different for tampered request")
}

func TestNodeJWTGenerator_NoSigner(t *testing.T) {
	// Test that generator returns error when no signer is configured
	_, csaPubKey, p2pId := createTestSigner()
	req := mockRequest{Field: "test"}

	jwtGenerator := NewNodeJWTGenerator(nil, csaPubKey, p2pId, types.EnvironmentNameProductionTestnet)

	_, err := jwtGenerator.CreateJWTForRequest(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no signer configured")
}

func TestNodeJWTGenerator_ValidateSignature(t *testing.T) {
	// prepare test data
	signer, csaPubKey, p2pId := createTestSigner()
	req := mockRequest{Field: "test request"}

	jwtGenerator := NewNodeJWTGenerator(signer, csaPubKey, p2pId, types.EnvironmentNameProductionTestnet)

	jwtToken, err := jwtGenerator.CreateJWTForRequest(req)
	require.NoError(t, err)

	// Parse and validate the JWT token with signature verification
	token, err := jwt.ParseWithClaims(jwtToken, &types.NodeJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is EdDSA
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, assert.AnError
		}
		return csaPubKey, nil
	})

	require.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*types.NodeJWTClaims)
	require.True(t, ok)
	assert.NotEmpty(t, claims.P2PId)
}

package nodeauth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/utils"
)

// mockRequest is a simple type that implements fmt.Stringer.
type mockRequest struct {
	Field string
}

func (d mockRequest) String() string {
	return d.Field
}

// Helper function to create test Ed25519 keys
func createTestKeys() (ed25519.PrivateKey, ed25519.PublicKey, ed25519.PublicKey) {
	// Generate a private key for signing
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Failed to generate Ed25519 key pair: " + err.Error())
	}

	// Generate a separate public key for p2pId
	p2pId, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Failed to generate Ed25519 p2pId: " + err.Error())
	}

	return privateKey, publicKey, p2pId
}

func TestNodeJWTGenerator_CreateJWTForRequest(t *testing.T) {
	// prepare test data
	privateKey, publicKey, p2pId := createTestKeys()
	req := mockRequest{Field: "test request"}

	jwtGenerator := NewNodeJWTGenerator(privateKey, publicKey, p2pId, EnvironmentNameProductionTestnet)

	jwtToken, err := jwtGenerator.CreateJWTForRequest(req)
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	// verify expected values
	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &NodeJWTClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(*NodeJWTClaims)
	require.True(t, ok)
	assert.Equal(t, hex.EncodeToString(p2pId), claims.Issuer)

	expectedP2PIdHex := hex.EncodeToString(p2pId)
	expectedPublicKeyHex := hex.EncodeToString(publicKey)

	assert.Equal(t, expectedP2PIdHex, claims.Subject)
	assert.Equal(t, expectedP2PIdHex, claims.P2PId)
	assert.Equal(t, expectedPublicKeyHex, claims.PublicKey)

	expectedDigest := utils.CalculateRequestDigest(req)
	assert.Equal(t, expectedDigest, claims.Digest)

	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)

	// verify hex encoding of P2P ID and public key
	decodedP2PId, err := hex.DecodeString(claims.P2PId)
	require.NoError(t, err)
	assert.Equal(t, p2pId, ed25519.PublicKey(decodedP2PId), "Decoded P2P ID should match original")

	decodedPublicKey, err := hex.DecodeString(claims.PublicKey)
	require.NoError(t, err)
	assert.Equal(t, publicKey, ed25519.PublicKey(decodedPublicKey), "Decoded public key should match original")
}

func TestNodeJWTGenerator_DigestTampering(t *testing.T) {
	privateKey, publicKey, p2pId := createTestKeys()
	jwtGenerator := NewNodeJWTGenerator(privateKey, publicKey, p2pId, EnvironmentNameProductionTestnet)
	req := mockRequest{Field: "original"}

	// Create JWT for original and altered request
	jwtToken, err := jwtGenerator.CreateJWTForRequest(req)
	require.NoError(t, err)
	reqAltered := mockRequest{Field: "tampered"}
	digestAltered := utils.CalculateRequestDigest(reqAltered)

	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &NodeJWTClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(*NodeJWTClaims)
	require.True(t, ok)

	// The digest in the token should not match the altered request
	assert.NotEqual(t, digestAltered, claims.Digest, "Expected JWT digest to not match altered request")
}

func TestNodeJWTGenerator_NoPrivateKey(t *testing.T) {
	_, publicKey, p2pId := createTestKeys()
	jwtGenerator := NewNodeJWTGenerator(nil, publicKey, p2pId, EnvironmentNameProductionTestnet)

	req := mockRequest{Field: "test request"}

	_, err := jwtGenerator.CreateJWTForRequest(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no private key configured")
}

func TestNodeJWTGenerator_ValidateSignature(t *testing.T) {
	privateKey, publicKey, p2pId := createTestKeys()
	jwtGenerator := NewNodeJWTGenerator(privateKey, publicKey, p2pId, EnvironmentNameProductionTestnet)

	req := mockRequest{Field: "test request"}

	jwtToken, err := jwtGenerator.CreateJWTForRequest(req)
	require.NoError(t, err)

	// Parse and verify the JWT token
	token, err := jwt.ParseWithClaims(jwtToken, &NodeJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Make sure it's signed with EdDSA
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return publicKey, nil
	})

	require.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*NodeJWTClaims)
	require.True(t, ok)
	assert.Equal(t, hex.EncodeToString(p2pId), claims.P2PId)
	assert.Equal(t, hex.EncodeToString(publicKey), claims.PublicKey)
}

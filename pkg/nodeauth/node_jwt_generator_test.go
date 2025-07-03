package nodeauth

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/utils"
)

func init() {
	// Register the custom signing method with the JWT library for testing
	jwt.RegisterSigningMethod("EdDSA", func() jwt.SigningMethod {
		return &NodeJWTSigningMethod{}
	})
}

// // ---------- Mock Request Type ----------
// mockRequest is a simple type that implements fmt.Stringer.
type mockRequest struct {
	Field string
}

func (d mockRequest) String() string {
	return d.Field
}

// Helper function to create test p2pId and publicKey
func createTestData() ([32]byte, [32]byte) {
	var p2pId [32]byte
	copy(p2pId[:], "test-p2p-id-123456789012345678901234")

	var publicKey [32]byte
	copy(publicKey[:], "test-public-key-1234567890123456")

	return p2pId, publicKey
}

func TestNodeJWTGenerator_CreateJWTForRequest(t *testing.T) {
	// prepare test data
	mockSig := mocks.NewSigner(t)
	p2pId, publicKey := createTestData()
	req := mockRequest{Field: "test request"}

	// set up mock expectations
	signature := make([]byte, 65)
	signature[64] = 0x1b
	mockSig.EXPECT().Sign(mock.AnythingOfType("[]uint8")).Return(signature, nil).Once()

	jwtGenerator := NewNodeJWTGenerator(mockSig, p2pId, publicKey)

	jwtToken, err := jwtGenerator.CreateJWTForRequest(req)
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	// verify expected values
	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &NodeJWTClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(*NodeJWTClaims)
	require.True(t, ok)
	assert.Equal(t, hex.EncodeToString(p2pId[:]), claims.Issuer)

	expectedP2PIdHex := hex.EncodeToString(p2pId[:])
	expectedPublicKeyHex := hex.EncodeToString(publicKey[:])

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
	var decodedP2PIdArray [32]byte
	copy(decodedP2PIdArray[:], decodedP2PId)
	assert.Equal(t, p2pId, decodedP2PIdArray, "Decoded P2P ID should match original")
	decodedPublicKey, err := hex.DecodeString(claims.PublicKey)
	require.NoError(t, err)
	var decodedPublicKeyArray [32]byte
	copy(decodedPublicKeyArray[:], decodedPublicKey)
	assert.Equal(t, publicKey, decodedPublicKeyArray, "Decoded public key should match original")
}

func TestNodeJWTGenerator_DigestTampering(t *testing.T) {
	mockSig := mocks.NewSigner(t)
	p2pId, publicKey := createTestData()
	jwtGenerator := NewNodeJWTGenerator(mockSig, p2pId, publicKey)
	req := mockRequest{Field: "original"}

	mockSig.EXPECT().Sign(mock.AnythingOfType("[]uint8")).Return([]byte("mock-signature"), nil).Maybe()

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

func TestNodeJWTGenerator_NoSigner(t *testing.T) {
	p2pId, publicKey := createTestData()
	jwtGenerator := NewNodeJWTGenerator(nil, p2pId, publicKey)

	req := mockRequest{Field: "test request"}

	_, err := jwtGenerator.CreateJWTForRequest(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no signer configured")
}

func TestNodeJWTGenerator_SigningError(t *testing.T) {
	mockSig := mocks.NewSigner(t)
	p2pId, publicKey := createTestData()
	jwtGenerator := NewNodeJWTGenerator(mockSig, p2pId, publicKey)

	req := mockRequest{Field: "test request"}

	mockSig.EXPECT().Sign(mock.AnythingOfType("[]uint8")).Return(nil, fmt.Errorf("mock signing error")).Maybe()

	_, err := jwtGenerator.CreateJWTForRequest(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mock signing error")
}

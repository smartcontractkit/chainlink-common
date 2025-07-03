package nodeauth

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"testing"
	"time"

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

// Helper function to create a valid JWT token for testing
func createValidJWTToken(t *testing.T, p2pId, publicKey [32]byte, request any) string {
	mockSigner := mocks.NewSigner(t)
	mockSigner.EXPECT().Sign(mock.AnythingOfType("[]uint8")).Return([]byte("mock-signature"), nil)

	generator := NewNodeJWTGenerator(mockSigner, p2pId, publicKey)
	token, err := generator.CreateJWTForRequest(request)
	require.NoError(t, err)
	return token
}

// Helper function to create test data
func createValidatorTestData() ([32]byte, [32]byte, mockRequest) {
	var p2pId [32]byte
	copy(p2pId[:], "test-p2p-id-123456789012345678901234")

	var publicKey [32]byte
	copy(publicKey[:], "test-public-key-1234567890123456")

	request := mockRequest{Field: "test request"}

	return p2pId, publicKey, request
}

func TestNodeJWTValidator_ValidateJWT_Success(t *testing.T) {
	// Setup
	p2pId, publicKey, request := createValidatorTestData()

	mockTopologyProvider := mocks.NewNodeTopologyProvider(t)
	logger := slog.Default()
	validator := NewNodeJWTValidator(mockTopologyProvider, logger)

	// Create valid JWT token
	tokenString := createValidJWTToken(t, p2pId, publicKey, request)

	// Mock topology provider to return authorized
	mockTopologyProvider.EXPECT().
		IsNodeAuthorized(mock.Anything, hex.EncodeToString(p2pId[:]), publicKey).
		Return(true, nil).
		Once()

	// Execute
	isValid, err := validator.ValidateJWT(context.Background(), tokenString, request)

	// Assert
	assert.True(t, isValid)
	assert.NoError(t, err)
}

func TestNodeJWTValidator_ValidateJWT_InvalidJWTFormat(t *testing.T) {
	// Setup
	mockTopologyProvider := mocks.NewNodeTopologyProvider(t)
	logger := slog.Default()
	validator := NewNodeJWTValidator(mockTopologyProvider, logger)

	// Execute with invalid JWT
	isValid, err := validator.ValidateJWT(context.Background(), "invalid.jwt.token", nil)

	// Assert
	assert.False(t, isValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse and validate JWT claims")
}

func TestNodeJWTValidator_ValidateJWT_InvalidPublicKey(t *testing.T) {
	// Setup
	p2pId, _, request := createValidatorTestData()

	mockTopologyProvider := mocks.NewNodeTopologyProvider(t)
	logger := slog.Default()
	validator := NewNodeJWTValidator(mockTopologyProvider, logger)

	// Create JWT with invalid public key
	now := time.Now()
	claims := NodeJWTClaims{
		P2PId:       hex.EncodeToString(p2pId[:]),
		PublicKey:   "invalid-hex-key", // Invalid hex
		Environment: "test",
		Digest:      utils.CalculateRequestDigest(request),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Execute
	isValid, err := validator.ValidateJWT(context.Background(), tokenString, request)

	// Assert
	assert.False(t, isValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key format")
}

func TestNodeJWTValidator_ValidateJWT_DigestMismatch(t *testing.T) {
	// Setup
	p2pId, publicKey, request := createValidatorTestData()

	mockTopologyProvider := mocks.NewNodeTopologyProvider(t)
	logger := slog.Default()
	validator := NewNodeJWTValidator(mockTopologyProvider, logger)

	// Create JWT for one request
	tokenString := createValidJWTToken(t, p2pId, publicKey, request)

	// Try to validate with different request (digest mismatch)
	differentRequest := mockRequest{Field: "different request"}

	// Execute
	isValid, err := validator.ValidateJWT(context.Background(), tokenString, differentRequest)

	// Assert
	assert.False(t, isValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request integrity check failed")
}

func TestNodeJWTValidator_ValidateJWT_ExpiredToken(t *testing.T) {
	// Setup
	p2pId, publicKey, request := createValidatorTestData()

	mockTopologyProvider := mocks.NewNodeTopologyProvider(t)
	logger := slog.Default()
	validator := NewNodeJWTValidator(mockTopologyProvider, logger)

	// Create expired JWT
	pastTime := time.Now().Add(-1 * time.Hour)
	claims := NodeJWTClaims{
		P2PId:       hex.EncodeToString(p2pId[:]),
		PublicKey:   hex.EncodeToString(publicKey[:]),
		Environment: "test",
		Digest:      utils.CalculateRequestDigest(request),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(pastTime), // Expired
			IssuedAt:  jwt.NewNumericDate(pastTime.Add(-5 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Execute
	isValid, err := validator.ValidateJWT(context.Background(), tokenString, request)

	// Assert
	assert.False(t, isValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT claims validation failed")
}

func TestNodeJWTValidator_ValidateJWT_UnauthorizedNode(t *testing.T) {
	// Setup
	p2pId, publicKey, request := createValidatorTestData()

	mockTopologyProvider := mocks.NewNodeTopologyProvider(t)
	logger := slog.Default()
	validator := NewNodeJWTValidator(mockTopologyProvider, logger)

	// Create valid JWT token
	tokenString := createValidJWTToken(t, p2pId, publicKey, request)

	// Mock topology provider to return unauthorized
	mockTopologyProvider.EXPECT().
		IsNodeAuthorized(mock.Anything, hex.EncodeToString(p2pId[:]), publicKey).
		Return(false, nil).
		Once()

	// Execute
	isValid, err := validator.ValidateJWT(context.Background(), tokenString, request)

	// Assert
	assert.False(t, isValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized node")
}

func TestNodeJWTValidator_ValidateJWT_TopologyProviderError(t *testing.T) {
	// Setup
	p2pId, publicKey, request := createValidatorTestData()

	mockTopologyProvider := mocks.NewNodeTopologyProvider(t)
	logger := slog.Default()
	validator := NewNodeJWTValidator(mockTopologyProvider, logger)

	// Create valid JWT token
	tokenString := createValidJWTToken(t, p2pId, publicKey, request)

	// Mock topology provider to return error
	mockTopologyProvider.EXPECT().
		IsNodeAuthorized(mock.Anything, hex.EncodeToString(p2pId[:]), publicKey).
		Return(false, fmt.Errorf("topology provider error")).
		Once()

	// Execute
	isValid, err := validator.ValidateJWT(context.Background(), tokenString, request)

	// Assert
	assert.False(t, isValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node validation failed")
}

func TestNodeJWTValidator_ParseJWTClaims_Success(t *testing.T) {
	// Setup
	p2pId, publicKey, request := createValidatorTestData()

	logger := slog.Default()
	validator := NewNodeJWTValidator(nil, logger)

	// Create valid JWT token
	tokenString := createValidJWTToken(t, p2pId, publicKey, request)

	// Execute
	claims, err := validator.parseJWTClaims(tokenString)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, hex.EncodeToString(p2pId[:]), claims.P2PId)
	assert.Equal(t, hex.EncodeToString(publicKey[:]), claims.PublicKey)
}

func TestNodeJWTValidator_ParseJWTClaims_InvalidFormat(t *testing.T) {
	// Setup
	logger := slog.Default()
	validator := NewNodeJWTValidator(nil, logger)

	// Execute
	claims, err := validator.parseJWTClaims("invalid.jwt")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid JWT format")
}

func TestNodeJWTValidator_DecodePublicKey_Success(t *testing.T) {
	// Setup
	logger := slog.Default()
	validator := NewNodeJWTValidator(nil, logger)

	originalKey := [32]byte{}
	copy(originalKey[:], "test-public-key-1234567890123456")
	publicKeyHex := hex.EncodeToString(originalKey[:])

	// Execute
	decodedKey, err := validator.decodePublicKey(publicKeyHex)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, originalKey, decodedKey)
}

func TestNodeJWTValidator_DecodePublicKey_InvalidHex(t *testing.T) {
	// Setup
	logger := slog.Default()
	validator := NewNodeJWTValidator(nil, logger)

	// Execute
	_, err := validator.decodePublicKey("invalid-hex-string")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex encoding")
}

func TestNodeJWTValidator_VerifyRequestDigest_Success(t *testing.T) {
	// Setup
	logger := slog.Default()
	validator := NewNodeJWTValidator(nil, logger)

	request := mockRequest{Field: "test request"}
	expectedDigest := utils.CalculateRequestDigest(request)

	claims := &NodeJWTClaims{
		Digest: expectedDigest,
	}

	// Execute
	err := validator.verifyRequestDigest(claims, request)

	// Assert
	assert.NoError(t, err)
}

func TestNodeJWTValidator_VerifyRequestDigest_Mismatch(t *testing.T) {
	// Setup
	logger := slog.Default()
	validator := NewNodeJWTValidator(nil, logger)

	request := mockRequest{Field: "test request"}
	differentRequest := mockRequest{Field: "different request"}
	wrongDigest := utils.CalculateRequestDigest(differentRequest)

	claims := &NodeJWTClaims{
		Digest: wrongDigest,
	}

	// Execute
	err := validator.verifyRequestDigest(claims, request)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "digest mismatch")
}

func TestNodeJWTValidator_VerifyStandardClaims_Success(t *testing.T) {
	// Setup
	logger := slog.Default()
	validator := NewNodeJWTValidator(nil, logger)

	now := time.Now()
	claims := &NodeJWTClaims{
		P2PId:       "746573742d7032702d69642d3132333435363738393031323334353637383930", 
		PublicKey:   "746573742d7075626c69632d6b65792d31323334353637383930313233343536", 
		Environment: "test",
		Digest:      "746573742d646967657374000000000000000000000000000000000000000000", // All hex valid
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Minute)),
		},
	}

	// Execute
	err := validator.verifyStandardClaims(claims)

	// Assert
	assert.NoError(t, err)
}

func TestNodeJWTValidator_VerifyStandardClaims_MissingRequiredFields(t *testing.T) {
	// Setup
	logger := slog.Default()
	validator := NewNodeJWTValidator(nil, logger)

	claims := &NodeJWTClaims{
		// Missing required fields
		P2PId: "", 
	}

	// Execute
	err := validator.verifyStandardClaims(claims)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "claims validation failed")
}

package nodeauth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
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

// Test request type
type testRequest struct {
	Field string
}

func (r testRequest) String() string {
	return r.Field
}

// Helper function to create test keys
func createValidatorTestKeys() (ed25519.PrivateKey, ed25519.PublicKey, ed25519.PublicKey) {
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

// Helper function to create a valid JWT token
func createValidJWT(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, p2pId ed25519.PublicKey, request any) (string, error) {
	generator := NewNodeJWTGenerator(privateKey, publicKey, p2pId, EnvironmentNameProductionTestnet)
	return generator.CreateJWTForRequest(request)
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestNodeJWTValidator_ValidateJWT_Success(t *testing.T) {
	// Setup
	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create mock topology provider
	mockProvider := mocks.NewNodeTopologyProvider(t)

	mockProvider.EXPECT().IsNodeAuthorized(
		mock.Anything,
		p2pId,
		publicKey,
	).Return(true, nil).Once()

	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	isValid, err := validator.ValidateJWT(context.Background(), jwtToken, request)

	// Assert
	require.NoError(t, err)
	assert.True(t, isValid)
}

func TestNodeJWTValidator_ValidateJWT_InvalidTokenFormat(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())
	request := testRequest{Field: "test request"}

	// Test with invalid token
	isValid, err := validator.ValidateJWT(context.Background(), "invalid.jwt.token", request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "failed to parse and validate JWT claims")
}

func TestNodeJWTValidator_ValidateJWT_InvalidPublicKey(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())
	request := testRequest{Field: "test request"}

	// Create JWT with invalid public key
	now := time.Now()
	claims := NodeJWTClaims{
		P2PId:       "valid-p2p-id",
		PublicKey:   "invalid-hex-key",
		Environment: "test",
		Digest:      utils.CalculateRequestDigest(request),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Subject:   "test-subject",
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(ed25519.PrivateKey(make([]byte, ed25519.PrivateKeySize)))
	require.NoError(t, err)

	// Test
	isValid, err := validator.ValidateJWT(context.Background(), tokenString, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "invalid public key format")
}

func TestNodeJWTValidator_ValidateJWT_SignatureVerificationFailed(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())
	request := testRequest{Field: "test request"}

	// Create JWT with one key but sign with another
	_, rightPublicKey, p2pId := createValidatorTestKeys()
	wrongPrivateKey, _, _ := createValidatorTestKeys()

	// Create JWT with right public key but wrong private key
	now := time.Now()
	claims := NodeJWTClaims{
		P2PId:       hex.EncodeToString(p2pId),
		PublicKey:   hex.EncodeToString(rightPublicKey),
		Environment: "test",
		Digest:      utils.CalculateRequestDigest(request),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hex.EncodeToString(p2pId),
			Subject:   hex.EncodeToString(p2pId),
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(wrongPrivateKey)
	require.NoError(t, err)

	// Test
	isValid, err := validator.ValidateJWT(context.Background(), tokenString, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "JWT signature verification failed")
}

func TestNodeJWTValidator_ValidateJWT_DigestMismatch(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	privateKey, publicKey, p2pId := createValidatorTestKeys()
	originalRequest := testRequest{Field: "original"}
	tamperedRequest := testRequest{Field: "tampered"}

	// Create JWT for original request
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, originalRequest)
	require.NoError(t, err)

	// Test with tampered request
	isValid, err := validator.ValidateJWT(context.Background(), jwtToken, tamperedRequest)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "request integrity check failed")
}

func TestNodeJWTValidator_ValidateJWT_ExpiredToken(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create expired JWT
	now := time.Now()
	claims := NodeJWTClaims{
		P2PId:       hex.EncodeToString(p2pId),
		PublicKey:   hex.EncodeToString(publicKey),
		Environment: "test",
		Digest:      utils.CalculateRequestDigest(request),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hex.EncodeToString(p2pId),
			Subject:   hex.EncodeToString(p2pId),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// Test
	isValid, err := validator.ValidateJWT(context.Background(), tokenString, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "JWT signature verification failed")
}

func TestNodeJWTValidator_ValidateJWT_UnauthorizedNode(t *testing.T) {
	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create mock topology provider that returns unauthorized
	mockProvider := mocks.NewNodeTopologyProvider(t)

	mockProvider.EXPECT().IsNodeAuthorized(
		mock.Anything,
		p2pId,
		publicKey,
	).Return(false, nil).Once()

	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	isValid, err := validator.ValidateJWT(context.Background(), jwtToken, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "unauthorized node")
}

func TestNodeJWTValidator_ValidateJWT_TopologyProviderError(t *testing.T) {
	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create mock topology provider that returns error
	mockProvider := mocks.NewNodeTopologyProvider(t)

	mockProvider.EXPECT().IsNodeAuthorized(
		mock.Anything,
		p2pId,
		publicKey,
	).Return(false, fmt.Errorf("topology provider error")).Once()

	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	isValid, err := validator.ValidateJWT(context.Background(), jwtToken, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "node validation failed")
}

func TestNodeJWTValidator_parseJWTClaims_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	claims, err := validator.parseJWTClaims(jwtToken)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(p2pId), claims.P2PId)
	assert.Equal(t, hex.EncodeToString(publicKey), claims.PublicKey)
}

func TestNodeJWTValidator_parseJWTClaims_InvalidFormat(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	// Test with invalid token format
	_, err := validator.parseJWTClaims("invalid.jwt")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JWT format")
}

func TestNodeJWTValidator_decodePublicKey_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	_, publicKey, _ := createValidatorTestKeys()
	publicKeyHex := hex.EncodeToString(publicKey)

	// Test
	decodedKey, err := validator.decodePublicKey(publicKeyHex)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, publicKey, decodedKey)
}

func TestNodeJWTValidator_decodePublicKey_InvalidHex(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	// Test with invalid hex
	_, err := validator.decodePublicKey("invalid-hex")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex encoding")
}

func TestNodeJWTValidator_decodePublicKey_InvalidSize(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	// Test with wrong size (too short)
	_, err := validator.decodePublicKey("1234")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key size")
}

func TestNodeJWTValidator_decodeP2PId_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	_, _, p2pId := createValidatorTestKeys()
	p2pIdHex := hex.EncodeToString(p2pId)

	// Test
	decodedP2PId, err := validator.decodeP2PId(p2pIdHex)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, p2pId, decodedP2PId)
}

func TestNodeJWTValidator_decodeP2PId_InvalidHex(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	// Test with invalid hex
	_, err := validator.decodeP2PId("invalid-hex")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex encoding")
}

func TestNodeJWTValidator_decodeP2PId_InvalidSize(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	// Test with wrong size (too short)
	_, err := validator.decodeP2PId("1234")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid p2pId size")
}

func TestNodeJWTValidator_verifyJWTSignature_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	err = validator.verifyJWTSignature(jwtToken, publicKey)

	// Assert
	require.NoError(t, err)
}

func TestNodeJWTValidator_verifyJWTSignature_WrongKey(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	privateKey, _, p2pId := createValidatorTestKeys()
	_, wrongPublicKey, _ := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create JWT with one key pair
	jwtToken, err := createValidJWT(privateKey, wrongPublicKey, p2pId, request)
	require.NoError(t, err)

	// Try to verify with different public key
	_, differentPublicKey, _ := createValidatorTestKeys()
	err = validator.verifyJWTSignature(jwtToken, differentPublicKey)

	// Assert - should fail because we're using wrong public key
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")
}

func TestNodeJWTValidator_verifyRequestDigest_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	request := testRequest{Field: "test request"}
	claims := &NodeJWTClaims{
		Digest: utils.CalculateRequestDigest(request),
	}

	// Test
	err := validator.verifyRequestDigest(claims, request)

	// Assert
	require.NoError(t, err)
}

func TestNodeJWTValidator_verifyRequestDigest_Mismatch(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	originalRequest := testRequest{Field: "original"}
	differentRequest := testRequest{Field: "different"}

	claims := &NodeJWTClaims{
		Digest: utils.CalculateRequestDigest(originalRequest),
	}

	// Test with different request
	err := validator.verifyRequestDigest(claims, differentRequest)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "digest mismatch")
}

func TestNodeJWTValidator_verifyStandardClaims_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	now := time.Now()
	claims := &NodeJWTClaims{
		P2PId:       "valid-p2p-id",
		PublicKey:   "valid-public-key",
		Environment: "test",
		Digest:      "valid-digest",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Test
	err := validator.verifyStandardClaims(claims)

	// Assert
	require.NoError(t, err)
}

func TestNodeJWTValidator_verifyStandardClaims_MissingFields(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	// Claims with missing required fields
	claims := &NodeJWTClaims{
		P2PId: "valid-p2p-id",
		// Missing PublicKey, Environment, Digest
	}

	// Test
	err := validator.verifyStandardClaims(claims)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claims validation failed")
}

func TestNodeJWTValidator_verifyStandardClaims_ExpiredToken(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	now := time.Now()
	claims := &NodeJWTClaims{
		P2PId:       "valid-p2p-id",
		PublicKey:   "valid-public-key",
		Environment: "test",
		Digest:      "valid-digest",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}

	// Test
	err := validator.verifyStandardClaims(claims)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token expired")
}

func TestNodeJWTValidator_verifyStandardClaims_FutureIssuedAt(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	validator := NewNodeJWTValidator(mockProvider, createTestLogger())

	now := time.Now()
	claims := &NodeJWTClaims{
		P2PId:       "valid-p2p-id",
		PublicKey:   "valid-public-key",
		Environment: "test",
		Digest:      "valid-digest",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now.Add(2 * workflowJWTExpiration)), // Too far in future
		},
	}

	// Test
	err := validator.verifyStandardClaims(claims)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token issued too far in future")
}

func TestNewNodeJWTValidator(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	logger := createTestLogger()

	validator := NewNodeJWTValidator(mockProvider, logger)

	assert.NotNil(t, validator)
	assert.Equal(t, mockProvider, validator.nodeTopologyProvider)
	assert.NotNil(t, validator.parser)
	assert.NotNil(t, validator.validator)
	assert.Equal(t, logger, validator.logger)
}

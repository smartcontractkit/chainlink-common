package jwt

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

	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/jwt/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/utils"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
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
	// Create signer using the private key
	account := hex.EncodeToString(publicKey)
	signFn := func(ctx context.Context, account string, data []byte) (signed []byte, err error) {
		return ed25519.Sign(privateKey, data), nil
	}

	signer, err := core.NewEd25519Signer(account, signFn)
	if err != nil {
		return "", err
	}

	generator := NewNodeJWTGenerator(signer, publicKey, p2pId, EnvironmentNameProductionTestnet)
	return generator.CreateJWTForRequest(request)
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestNodeJWTAuthenticator_AuthenticateJWT_Success(t *testing.T) {
	// Setup
	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create mock topology provider
	mockProvider := mocks.NewNodeTopologyProvider(t)

	mockProvider.EXPECT().IsNodePubKeyTrusted(
		mock.Anything,
		p2pId,
		publicKey,
	).Return(true, nil).Once()

	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	isValid, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, request)

	// Assert
	require.NoError(t, err)
	assert.True(t, isValid)
}

func TestNodeJWTAuthenticator_AuthenticateJWT_InvalidTokenFormat(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())
	request := testRequest{Field: "test request"}

	// Test with invalid token
	isValid, err := authenticator.AuthenticateJWT(context.Background(), "invalid.jwt.token", request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "failed to parse and validate JWT claims")
}

func TestNodeJWTAuthenticator_AuthenticateJWT_InvalidPublicKey(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())
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
	isValid, err := authenticator.AuthenticateJWT(context.Background(), tokenString, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "invalid public key format")
}

func TestNodeJWTAuthenticator_AuthenticateJWT_SignatureVerificationFailed(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())
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
	isValid, err := authenticator.AuthenticateJWT(context.Background(), tokenString, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "JWT signature verification failed")
}

func TestNodeJWTAuthenticator_AuthenticateJWT_DigestMismatch(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	privateKey, publicKey, p2pId := createValidatorTestKeys()
	originalRequest := testRequest{Field: "original"}
	tamperedRequest := testRequest{Field: "tampered"}

	// Create JWT for original request
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, originalRequest)
	require.NoError(t, err)

	// Test with tampered request
	isValid, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, tamperedRequest)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "request integrity check failed")
}

func TestNodeJWTAuthenticator_AuthenticateJWT_ExpiredToken(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

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
	isValid, err := authenticator.AuthenticateJWT(context.Background(), tokenString, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "JWT signature verification failed")
}

func TestNodeJWTAuthenticator_AuthenticateJWT_UnauthorizedNode(t *testing.T) {
	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create mock topology provider that returns unauthorized
	mockProvider := mocks.NewNodeTopologyProvider(t)

	mockProvider.EXPECT().IsNodePubKeyTrusted(
		mock.Anything,
		p2pId,
		publicKey,
	).Return(false, nil).Once()

	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	isValid, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "unauthorized node")
}

func TestNodeJWTAuthenticator_AuthenticateJWT_TopologyProviderError(t *testing.T) {
	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create mock topology provider that returns error
	mockProvider := mocks.NewNodeTopologyProvider(t)

	mockProvider.EXPECT().IsNodePubKeyTrusted(
		mock.Anything,
		p2pId,
		publicKey,
	).Return(false, fmt.Errorf("topology provider error")).Once()

	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	isValid, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, request)

	// Assert
	require.Error(t, err)
	assert.False(t, isValid)
	assert.Contains(t, err.Error(), "node validation failed")
}

func TestNodeJWTAuthenticator_parseJWTClaims_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	claims, err := authenticator.parseJWTClaims(jwtToken)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(p2pId), claims.P2PId)
	assert.Equal(t, hex.EncodeToString(publicKey), claims.PublicKey)
}

func TestNodeJWTAuthenticator_parseJWTClaims_InvalidFormat(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Test with invalid token format
	_, err := authenticator.parseJWTClaims("invalid.jwt")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JWT format")
}

func TestNodeJWTAuthenticator_decodePublicKey_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	_, publicKey, _ := createValidatorTestKeys()
	publicKeyHex := hex.EncodeToString(publicKey)

	// Test
	decodedKey, err := authenticator.decodePublicKey(publicKeyHex)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, publicKey, decodedKey)
}

func TestNodeJWTAuthenticator_decodePublicKey_InvalidHex(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Test with invalid hex
	_, err := authenticator.decodePublicKey("invalid-hex")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex encoding")
}

func TestNodeJWTAuthenticator_decodePublicKey_InvalidSize(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Test with wrong size (too short)
	_, err := authenticator.decodePublicKey("1234")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key size")
}

func TestNodeJWTAuthenticator_decodeP2PId_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	_, _, p2pId := createValidatorTestKeys()
	p2pIdHex := hex.EncodeToString(p2pId)

	// Test
	decodedP2PId, err := authenticator.decodeP2PId(p2pIdHex)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, p2pId, decodedP2PId)
}

func TestNodeJWTAuthenticator_decodeP2PId_InvalidHex(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Test with invalid hex
	_, err := authenticator.decodeP2PId("invalid-hex")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex encoding")
}

func TestNodeJWTAuthenticator_decodeP2PId_InvalidSize(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Test with wrong size (too short)
	_, err := authenticator.decodeP2PId("1234")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid p2pId size")
}

func TestNodeJWTAuthenticator_verifyJWTSignature_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	privateKey, publicKey, p2pId := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create valid JWT
	jwtToken, err := createValidJWT(privateKey, publicKey, p2pId, request)
	require.NoError(t, err)

	// Test
	err = authenticator.verifyJWTSignature(jwtToken, publicKey)

	// Assert
	require.NoError(t, err)
}

func TestNodeJWTAuthenticator_verifyJWTSignature_WrongKey(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	privateKey, _, p2pId := createValidatorTestKeys()
	_, wrongPublicKey, _ := createValidatorTestKeys()
	request := testRequest{Field: "test request"}

	// Create JWT with one key pair
	jwtToken, err := createValidJWT(privateKey, wrongPublicKey, p2pId, request)
	require.NoError(t, err)

	// Try to verify with different public key
	_, differentPublicKey, _ := createValidatorTestKeys()
	err = authenticator.verifyJWTSignature(jwtToken, differentPublicKey)

	// Assert - should fail because we're using wrong public key
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")
}

func TestNodeJWTAuthenticator_verifyRequestDigest_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	request := testRequest{Field: "test request"}
	claims := &NodeJWTClaims{
		Digest: utils.CalculateRequestDigest(request),
	}

	// Test
	err := authenticator.verifyRequestDigest(claims, request)

	// Assert
	require.NoError(t, err)
}

func TestNodeJWTAuthenticator_verifyRequestDigest_Mismatch(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	originalRequest := testRequest{Field: "original"}
	differentRequest := testRequest{Field: "different"}

	claims := &NodeJWTClaims{
		Digest: utils.CalculateRequestDigest(originalRequest),
	}

	// Test with different request
	err := authenticator.verifyRequestDigest(claims, differentRequest)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "digest mismatch")
}

func TestNodeJWTAuthenticator_verifyStandardClaims_Success(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

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
	err := authenticator.verifyStandardClaims(claims)

	// Assert
	require.NoError(t, err)
}

func TestNodeJWTAuthenticator_verifyStandardClaims_MissingFields(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Claims with missing required fields
	claims := &NodeJWTClaims{
		P2PId: "valid-p2p-id",
		// Missing PublicKey, Environment, Digest
	}

	// Test
	err := authenticator.verifyStandardClaims(claims)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claims validation failed")
}

func TestNodeJWTAuthenticator_verifyStandardClaims_ExpiredToken(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

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
	err := authenticator.verifyStandardClaims(claims)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token expired")
}

func TestNodeJWTAuthenticator_verifyStandardClaims_FutureIssuedAt(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

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
	err := authenticator.verifyStandardClaims(claims)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token issued too far in future")
}

func TestNewNodeJWTAuthenticator(t *testing.T) {
	mockProvider := mocks.NewNodeTopologyProvider(t)
	logger := createTestLogger()

	authenticator := NewNodeJWTAuthenticator(mockProvider, logger)

	assert.NotNil(t, authenticator)
	assert.Equal(t, mockProvider, authenticator.nodeTopologyProvider)
	assert.NotNil(t, authenticator.parser)
	assert.NotNil(t, authenticator.validator)
	assert.Equal(t, logger, authenticator.logger)
}

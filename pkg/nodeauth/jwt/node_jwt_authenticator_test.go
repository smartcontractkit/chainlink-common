package jwt

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/jwt/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/types"
	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/utils"
	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

// Test request type
type testRequest struct {
	Field string
}

func (r testRequest) String() string {
	return r.Field
}

// Helper function to create test keys
func createValidatorTestKeys() (ed25519.PrivateKey, ed25519.PublicKey, p2ptypes.PeerID) {
	// Generate a private key for signing
	csaPubKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Failed to generate Ed25519 key pair: " + err.Error())
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

	return privateKey, csaPubKey, p2pId
}

// Helper function to create a valid JWT token
func createValidJWT(privateKey ed25519.PrivateKey, csaPubKey ed25519.PublicKey, p2pId p2ptypes.PeerID, environment types.EnvironmentName) string {
	testRequest := testRequest{Field: "test-request"}
	digest := utils.CalculateRequestDigest(testRequest)

	// Create JWT claims
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
		P2PId:       p2pId.String(),
		PublicKey:   hex.EncodeToString(csaPubKey),
		Environment: string(environment),
		Digest:      digest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p2pId.String(), // Issuer: Node's P2P ID
			Subject:   p2pId.String(), // Subject: Node's P2P ID
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic("Failed to sign JWT: " + err.Error())
	}

	return tokenString
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestNodeJWTAuthenticator_AuthenticateJWT_ValidToken(t *testing.T) {

	// Given
	privateKey, csaPubKey, p2pId := createValidatorTestKeys()
	mockProvider := &mocks.NodeAuthProvider{}
	mockProvider.On("IsNodePubKeyTrusted", mock.Anything, p2pId, csaPubKey, string(types.EnvironmentNameProductionTestnet)).Return(true, nil)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	jwtToken := createValidJWT(privateKey, csaPubKey, p2pId, types.EnvironmentNameProductionTestnet)

	// Test
	testRequest := testRequest{Field: "test-request"}
	valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

	// Assertions
	require.NoError(t, err)
	assert.True(t, valid)
	assert.NotNil(t, claims)
	assert.Equal(t, p2pId.String(), claims.P2PId)
	assert.Equal(t, string(types.EnvironmentNameProductionTestnet), claims.Environment)
	mockProvider.AssertExpectations(t)
}

func TestNodeJWTAuthenticator_AuthenticateJWT_TamperedRequest(t *testing.T) {

	// Given
	privateKey, csaPubKey, p2pId := createValidatorTestKeys()
	mockProvider := &mocks.NodeAuthProvider{}
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	jwtToken := createValidJWT(privateKey, csaPubKey, p2pId, types.EnvironmentNameProductionTestnet)

	// When - tampered request
	valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, "different-request")

	// Expect
	require.Error(t, err)
	assert.False(t, valid)
	assert.NotNil(t, claims)
}

func TestNodeJWTAuthenticator_AuthenticateJWT_ExpiredToken(t *testing.T) {

	privateKey, csaPubKey, p2pId := createValidatorTestKeys()
	mockProvider := &mocks.NodeAuthProvider{}
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Given: Expired JWT
	testRequest := testRequest{Field: "test-request"}
	digest := utils.CalculateRequestDigest(testRequest)

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
		P2PId:       p2pId.String(),
		PublicKey:   hex.EncodeToString(csaPubKey),
		Environment: string(types.EnvironmentNameProductionTestnet),
		Digest:      digest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p2pId.String(),
			Subject:   p2pId.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	})

	jwtToken, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// When: Authenticate JWT
	valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

	// Expect
	require.Error(t, err)
	assert.False(t, valid)
	assert.NotNil(t, claims)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestNodeJWTAuthenticator_AuthenticateJWT_InvalidPublicKeySignature(t *testing.T) {

	// Given - create two different key pairs
	privateKey1, _, _ := createValidatorTestKeys()
	_, csaPubKey2, p2pId2 := createValidatorTestKeys()

	mockProvider := &mocks.NodeAuthProvider{}
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Given: JWT signature mismatch public key
	testRequest := testRequest{Field: "test-request"}
	digest := utils.CalculateRequestDigest(testRequest)

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
		P2PId:       p2pId2.String(), // Claim to be from node 2 but signed with node 1's private key
		PublicKey:   hex.EncodeToString(csaPubKey2),
		Environment: string(types.EnvironmentNameProductionTestnet),
		Digest:      digest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p2pId2.String(),
			Subject:   p2pId2.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})
	jwtToken, err := token.SignedString(privateKey1)
	require.NoError(t, err)

	// When: Authenticate JWT
	valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

	// Expect - should fail due to signature mismatch
	require.Error(t, err)
	assert.False(t, valid)
	assert.NotNil(t, claims)
	assert.Contains(t, err.Error(), "JWT signature verification failed")
}

func TestNodeJWTAuthenticator_AuthenticateJWT_UntrustedPublicKey(t *testing.T) {

	// Given
	privateKey, csaPubKey, p2pId := createValidatorTestKeys()

	mockProvider := &mocks.NodeAuthProvider{}
	mockProvider.On("IsNodePubKeyTrusted", mock.Anything, p2pId, csaPubKey, string(types.EnvironmentNameProductionTestnet)).Return(false, nil)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Given: Valid JWT
	jwtToken := createValidJWT(privateKey, csaPubKey, p2pId, types.EnvironmentNameProductionTestnet)

	// Test
	testRequest := testRequest{Field: "test-request"}
	valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

	// Expect - should fail because node is not trusted
	require.Error(t, err)
	assert.False(t, valid)
	assert.NotNil(t, claims)
	assert.Contains(t, err.Error(), "unauthorized node")
	mockProvider.AssertExpectations(t)
}

func TestNodeJWTAuthenticator_parseJWTClaims_Success(t *testing.T) {
	mockProvider := mocks.NewNodeAuthProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	privateKey, csaPubKey, p2pId := createValidatorTestKeys()

	// Create valid JWT
	jwtToken := createValidJWT(privateKey, csaPubKey, p2pId, types.EnvironmentNameProductionTestnet)

	// When: Parse JWT claims
	claims, err := authenticator.parseJWTClaims(jwtToken)

	// Expect
	require.NoError(t, err)
	assert.Equal(t, p2pId.String(), claims.P2PId)
	assert.Equal(t, hex.EncodeToString(csaPubKey), claims.PublicKey)
}

func TestNodeJWTAuthenticator_parseJWTClaims_InvalidFormat(t *testing.T) {
	mockProvider := mocks.NewNodeAuthProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Test with invalid token format
	_, err := authenticator.parseJWTClaims("invalid.jwt")

	// Expect
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JWT format")
}

func TestNodeJWTAuthenticator_verifyJWTSignature_Success(t *testing.T) {
	mockProvider := mocks.NewNodeAuthProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	privateKey, csaPubKey, p2pId := createValidatorTestKeys()

	// Create valid JWT
	jwtToken := createValidJWT(privateKey, csaPubKey, p2pId, types.EnvironmentNameProductionTestnet)

	// Test
	err := authenticator.verifyJWTSignature(jwtToken, csaPubKey)

	// Assert
	require.NoError(t, err)
}

func TestNodeJWTAuthenticator_verifyRequestDigest_DigestMismatch(t *testing.T) {
	mockProvider := mocks.NewNodeAuthProvider(t)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	originalRequest := testRequest{Field: "original"}
	differentRequest := testRequest{Field: "different"}

	claims := &types.NodeJWTClaims{
		Digest: utils.CalculateRequestDigest(originalRequest),
	}

	// Test with different request
	err := authenticator.verifyRequestDigest(claims, differentRequest)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "digest mismatch")
}

func TestNewNodeJWTAuthenticator(t *testing.T) {
	mockProvider := mocks.NewNodeAuthProvider(t)
	logger := createTestLogger()

	authenticator := NewNodeJWTAuthenticator(mockProvider, logger)

	assert.NotNil(t, authenticator)
	assert.Equal(t, mockProvider, authenticator.nodeAuthProvider)
	assert.NotNil(t, authenticator.parser)
	assert.Equal(t, logger, authenticator.logger)
}

package jwt

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/jwt/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/types"
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
func createValidatorTestKeys() (ed25519.PrivateKey, ed25519.PublicKey) {
	// Generate a private key for signing
	csaPubKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Failed to generate Ed25519 key pair: " + err.Error())
	}

	return privateKey, csaPubKey
}

// Helper function to create a valid JWT token
func createValidJWT(privateKey ed25519.PrivateKey, csaPubKey ed25519.PublicKey) string {
	testRequest := testRequest{Field: "test-request"}
	digest := utils.CalculateRequestDigest(testRequest)

	// Create JWT claims
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
		PublicKey: hex.EncodeToString(csaPubKey),
		Digest:    digest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hex.EncodeToString(csaPubKey), // Issuer: Node's CSA PubKey
			Subject:   hex.EncodeToString(csaPubKey), // Subject: Node's CSA PubKey
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
	privateKey, csaPubKey := createValidatorTestKeys()
	mockProvider := &mocks.NodeAuthProvider{}
	mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(true, nil)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	jwtToken := createValidJWT(privateKey, csaPubKey)

	// Test
	testRequest := testRequest{Field: "test-request"}
	valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

	// Assertions
	require.NoError(t, err)
	assert.True(t, valid)
	assert.NotNil(t, claims)
	assert.Equal(t, hex.EncodeToString(csaPubKey), claims.PublicKey)
	mockProvider.AssertExpectations(t)
}

func TestNodeJWTAuthenticator_AuthenticateJWT_TamperedRequest(t *testing.T) {
	// Given
	privateKey, csaPubKey := createValidatorTestKeys()
	mockProvider := &mocks.NodeAuthProvider{}
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	jwtToken := createValidJWT(privateKey, csaPubKey)

	// When - tampered request
	valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, "different-request")

	// Expect
	require.Error(t, err)
	assert.False(t, valid)
	assert.NotNil(t, claims)
}

func TestNodeJWTAuthenticator_AuthenticateJWT_ExpiredToken(t *testing.T) {
	privateKey, csaPubKey := createValidatorTestKeys()
	mockProvider := &mocks.NodeAuthProvider{}
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Given: Expired JWT
	testRequest := testRequest{Field: "test-request"}
	digest := utils.CalculateRequestDigest(testRequest)

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
		PublicKey: hex.EncodeToString(csaPubKey),
		Digest:    digest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hex.EncodeToString(csaPubKey),           // Issuer: Node's CSA PubKey
			Subject:   hex.EncodeToString(csaPubKey),           // Subject: Node's CSA PubKey
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
	privateKey1, _ := createValidatorTestKeys()
	_, csaPubKey2 := createValidatorTestKeys()

	mockProvider := &mocks.NodeAuthProvider{}
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Given: JWT signature mismatch public key
	testRequest := testRequest{Field: "test-request"}
	digest := utils.CalculateRequestDigest(testRequest)

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
		PublicKey: hex.EncodeToString(csaPubKey2),
		Digest:    digest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hex.EncodeToString(csaPubKey2), // Issuer: Node's CSA PubKey
			Subject:   hex.EncodeToString(csaPubKey2), // Subject: Node's CSA PubKey
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
	privateKey, csaPubKey := createValidatorTestKeys()

	mockProvider := &mocks.NodeAuthProvider{}
	mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(false, nil)
	authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

	// Given: Valid JWT
	jwtToken := createValidJWT(privateKey, csaPubKey)

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

	privateKey, csaPubKey := createValidatorTestKeys()

	// Create valid JWT
	jwtToken := createValidJWT(privateKey, csaPubKey)

	// When: Parse JWT claims
	claims, err := authenticator.parseJWTClaims(jwtToken)

	// Expect
	require.NoError(t, err)
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

	privateKey, csaPubKey := createValidatorTestKeys()

	// Create valid JWT
	jwtToken := createValidJWT(privateKey, csaPubKey)

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

func TestNewNodeJWTAuthenticator_WithAndWithoutLeeway(t *testing.T) {
	t.Run("without config - backward compatible", func(t *testing.T) {
		mockProvider := &mocks.NodeAuthProvider{}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

		assert.NotNil(t, authenticator)
		assert.NotNil(t, authenticator.parser)
	})

	t.Run("with config struct with leeway", func(t *testing.T) {
		mockProvider := &mocks.NodeAuthProvider{}
		config := &NodeJWTAuthenticatorConfig{
			Leeway: 5 * time.Second,
		}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), config)

		assert.NotNil(t, authenticator)
		assert.NotNil(t, authenticator.parser)
	})

	t.Run("expired token within leeway - should pass", func(t *testing.T) {
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(true, nil)

		config := &NodeJWTAuthenticatorConfig{
			Leeway: 5 * time.Second,
		}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), config)

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Token expired 3 seconds ago (within 5s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(-3 * time.Second)),
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		require.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, claims)
		mockProvider.AssertExpectations(t)
	})

	t.Run("expired token beyond leeway - should fail", func(t *testing.T) {
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}

		config := &NodeJWTAuthenticatorConfig{
			Leeway: 5 * time.Second,
		}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), config)

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Token expired 10 seconds ago (beyond 5s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(-10 * time.Second)),
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		require.Error(t, err)
		assert.False(t, valid)
		assert.NotNil(t, claims)
		assert.Contains(t, err.Error(), "token is expired")
	})

	t.Run("future token within leeway - should pass", func(t *testing.T) {
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(true, nil)

		config := &NodeJWTAuthenticatorConfig{
			Leeway: 5 * time.Second,
		}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), config)

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Token issued 3 seconds in the future (within 5s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
				IssuedAt:  jwt.NewNumericDate(now.Add(3 * time.Second)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		require.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, claims)
		mockProvider.AssertExpectations(t)
	})

	t.Run("future token beyond leeway - should fail", func(t *testing.T) {
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}

		config := &NodeJWTAuthenticatorConfig{
			Leeway: 5 * time.Second,
		}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), config)

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Token issued 10 seconds in the future (beyond 5s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration + 10*time.Second)),
				IssuedAt:  jwt.NewNumericDate(now.Add(10 * time.Second)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		require.Error(t, err)
		assert.False(t, valid)
		assert.NotNil(t, claims)
		assert.Contains(t, err.Error(), "used before issued")
	})

	t.Run("with nil config - same as no config", func(t *testing.T) {
		mockProvider := &mocks.NodeAuthProvider{}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), nil)

		assert.NotNil(t, authenticator)
		assert.NotNil(t, authenticator.parser)
	})

	t.Run("with zero leeway - no leeway applied", func(t *testing.T) {
		mockProvider := &mocks.NodeAuthProvider{}
		config := &NodeJWTAuthenticatorConfig{
			Leeway: 0,
		}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), config)

		assert.NotNil(t, authenticator)
		assert.NotNil(t, authenticator.parser)
	})
}

// captureHandler is a minimal slog.Handler that captures log records for test assertions.
// It satisfies the full slog.Handler contract: WithAttrs and WithGroup return a new
// handler so that logger.With(...) / logger.WithGroup(...) calls don't silently discard
// attributes.
type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
	attrs   []slog.Attr
	groups  []string
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	// Prepend any inherited attrs so they appear in captured records.
	if len(h.attrs) > 0 {
		r2 := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		r2.AddAttrs(h.attrs...)
		r.Attrs(func(a slog.Attr) bool { r2.AddAttrs(a); return true })
		r = r2
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r.Clone())
	return nil
}
func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	merged := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(merged, h.attrs)
	copy(merged[len(h.attrs):], attrs)
	return &captureHandler{records: h.records, attrs: merged, groups: h.groups}
}
func (h *captureHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	groups := append(append([]string{}, h.groups...), name)
	return &captureHandler{records: h.records, attrs: h.attrs, groups: groups}
}

func TestNodeJWTAuthenticator_AuthenticateJWT_ProviderNonContextError(t *testing.T) {
	// Non-context provider errors must be logged at ERROR level.
	privateKey, csaPubKey := createValidatorTestKeys()
	providerErr := errors.New("database unavailable")
	mockProvider := &mocks.NodeAuthProvider{}
	mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(false, providerErr)

	h := &captureHandler{}
	authenticator := NewNodeJWTAuthenticator(mockProvider, slog.New(h))

	testRequest := testRequest{Field: "test-request"}
	valid, claims, err := authenticator.AuthenticateJWT(context.Background(), createValidJWT(privateKey, csaPubKey), testRequest)

	require.Error(t, err)
	assert.False(t, valid)
	assert.NotNil(t, claims)
	assert.Contains(t, err.Error(), "node validation failed")

	require.Len(t, h.records, 1)
	assert.Equal(t, slog.LevelError, h.records[0].Level, "non-context provider errors should log at ERROR")
	mockProvider.AssertExpectations(t)
}

func TestNodeJWTAuthenticator_AuthenticateJWT_ProviderContextCancelledError(t *testing.T) {
	// Context-cancellation errors from the provider must be logged at WARN, not ERROR,
	// because they are caused by the caller cancelling the request — not a system fault.
	privateKey, csaPubKey := createValidatorTestKeys()
	mockProvider := &mocks.NodeAuthProvider{}
	mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(false, context.Canceled)

	h := &captureHandler{}
	authenticator := NewNodeJWTAuthenticator(mockProvider, slog.New(h))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	testRequest := testRequest{Field: "test-request"}
	valid, claims, err := authenticator.AuthenticateJWT(ctx, createValidJWT(privateKey, csaPubKey), testRequest)

	require.Error(t, err)
	assert.False(t, valid)
	assert.NotNil(t, claims)
	assert.ErrorIs(t, err, context.Canceled)

	require.Len(t, h.records, 1)
	assert.Equal(t, slog.LevelWarn, h.records[0].Level, "context cancellation from provider should log at WARN not ERROR")
	mockProvider.AssertExpectations(t)
}

func TestNodeJWTAuthenticator_AuthenticateJWT_ProviderDeadlineExceededError(t *testing.T) {
	// context.DeadlineExceeded from the provider must also be logged at WARN, not ERROR,
	// because it is an expected transient condition (e.g. slow upstream), not a system fault.
	privateKey, csaPubKey := createValidatorTestKeys()
	mockProvider := &mocks.NodeAuthProvider{}
	mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(false, context.DeadlineExceeded)

	h := &captureHandler{}
	authenticator := NewNodeJWTAuthenticator(mockProvider, slog.New(h))

	ctx, cancel := context.WithTimeout(context.Background(), 0) // immediately expired
	defer cancel()

	testRequest := testRequest{Field: "test-request"}
	valid, claims, err := authenticator.AuthenticateJWT(ctx, createValidJWT(privateKey, csaPubKey), testRequest)

	require.Error(t, err)
	assert.False(t, valid)
	assert.NotNil(t, claims)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	require.Len(t, h.records, 1)
	assert.Equal(t, slog.LevelWarn, h.records[0].Level, "deadline exceeded from provider should log at WARN not ERROR")
	mockProvider.AssertExpectations(t)
}

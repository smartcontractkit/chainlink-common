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

func TestWithLeeway_OptionApplication(t *testing.T) {
	t.Run("WithLeeway option is applied correctly", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(true, nil)
		
		// Create authenticator with WithLeeway option
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), WithLeeway(5*time.Second))

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token that expired 3 seconds ago (within 5s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(-3 * time.Second)), // Expired 3s ago
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT - should succeed due to leeway
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Leeway was applied, token is accepted
		require.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, claims)
		mockProvider.AssertExpectations(t)
	})

	t.Run("multiple WithLeeway options can be applied", func(t *testing.T) {
		mockProvider := &mocks.NodeAuthProvider{}
		
		// Create authenticator with multiple leeway options
		authenticator := NewNodeJWTAuthenticator(
			mockProvider, 
			createTestLogger(), 
			WithLeeway(3*time.Second),
			WithLeeway(5*time.Second),
		)

		// Verify authenticator was created successfully with options applied
		assert.NotNil(t, authenticator)
		assert.NotNil(t, authenticator.parser)
	})

	t.Run("WithLeeway with different durations", func(t *testing.T) {
		durations := []time.Duration{
			0,
			1 * time.Second,
			5 * time.Second,
			30 * time.Second,
			1 * time.Minute,
		}

		for _, duration := range durations {
			t.Run(duration.String(), func(t *testing.T) {
				mockProvider := &mocks.NodeAuthProvider{}
				
				// Create option
				opt := WithLeeway(duration)
				
				// Apply option when creating authenticator
				authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), opt)
				
				assert.NotNil(t, authenticator)
				assert.NotNil(t, authenticator.parser)
			})
		}
	})
}

func TestNewNodeJWTAuthenticator_WithOptions(t *testing.T) {
	t.Run("no options creates default authenticator", func(t *testing.T) {
		mockProvider := &mocks.NodeAuthProvider{}
		logger := createTestLogger()

		authenticator := NewNodeJWTAuthenticator(mockProvider, logger)

		assert.NotNil(t, authenticator)
		assert.NotNil(t, authenticator.parser)
		assert.Equal(t, mockProvider, authenticator.nodeAuthProvider)
		assert.Equal(t, logger, authenticator.logger)
	})

	t.Run("options loop applies single option", func(t *testing.T) {
		mockProvider := &mocks.NodeAuthProvider{}
		logger := createTestLogger()

		// Single option
		authenticator := NewNodeJWTAuthenticator(mockProvider, logger, WithLeeway(10*time.Second))

		assert.NotNil(t, authenticator)
		assert.NotNil(t, authenticator.parser)
	})

	t.Run("options loop applies multiple options", func(t *testing.T) {
		mockProvider := &mocks.NodeAuthProvider{}
		logger := createTestLogger()

		// Multiple options applied via loop
		authenticator := NewNodeJWTAuthenticator(
			mockProvider, 
			logger, 
			WithLeeway(5*time.Second),
			WithLeeway(10*time.Second),
		)

		assert.NotNil(t, authenticator)
		assert.NotNil(t, authenticator.parser)
	})
}

func TestNodeJWTAuthenticator_AuthenticateJWT_LeewayHandlesClockSkew(t *testing.T) {
	t.Run("token expired within leeway window should be accepted", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(true, nil)
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), WithLeeway(5*time.Second))

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token that expired 3 seconds ago (within 5s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(-3 * time.Second)), // Expired 3s ago
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Should succeed due to leeway
		require.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, claims)
		mockProvider.AssertExpectations(t)
	})

	t.Run("token expired beyond leeway window should be rejected", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), WithLeeway(5*time.Second))

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token that expired 10 seconds ago (beyond 5s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(-10 * time.Second)), // Expired 10s ago
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Should fail as it's beyond leeway
		require.Error(t, err)
		assert.False(t, valid)
		assert.NotNil(t, claims)
		assert.Contains(t, err.Error(), "token is expired")
	})

	t.Run("token issued slightly in the future within leeway should be accepted", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(true, nil)
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), WithLeeway(5*time.Second))

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token issued 3 seconds in the future (within 5s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
				IssuedAt:  jwt.NewNumericDate(now.Add(3 * time.Second)), // Issued 3s in future
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Should succeed due to leeway
		require.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, claims)
		mockProvider.AssertExpectations(t)
	})

	t.Run("token issued far in the future beyond leeway should be rejected", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), WithLeeway(5*time.Second))

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token issued 10 seconds in the future (beyond 5s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration + 10*time.Second)),
				IssuedAt:  jwt.NewNumericDate(now.Add(10 * time.Second)), // Issued 10s in future
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Should fail as it's beyond leeway
		require.Error(t, err)
		assert.False(t, valid)
		assert.NotNil(t, claims)
		assert.Contains(t, err.Error(), "used before issued")
	})
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

func TestNewNodeJWTAuthenticator(t *testing.T) {
	t.Run("without options creates authenticator with default parser", func(t *testing.T) {
		mockProvider := mocks.NewNodeAuthProvider(t)
		logger := createTestLogger()

		authenticator := NewNodeJWTAuthenticator(mockProvider, logger)

		assert.NotNil(t, authenticator)
		assert.Equal(t, mockProvider, authenticator.nodeAuthProvider)
		assert.NotNil(t, authenticator.parser)
		assert.Equal(t, logger, authenticator.logger)
	})

	t.Run("with leeway option applies configuration correctly", func(t *testing.T) {
		mockProvider := mocks.NewNodeAuthProvider(t)
		logger := createTestLogger()

		// Create authenticator with leeway option
		authenticator := NewNodeJWTAuthenticator(mockProvider, logger, WithLeeway(5*time.Second))

		assert.NotNil(t, authenticator)
		assert.Equal(t, mockProvider, authenticator.nodeAuthProvider)
		assert.NotNil(t, authenticator.parser)
		assert.Equal(t, logger, authenticator.logger)
	})

	t.Run("with multiple option invocations", func(t *testing.T) {
		mockProvider := mocks.NewNodeAuthProvider(t)
		logger := createTestLogger()

		// Test that multiple calls to option functions work
		leewayOpt1 := WithLeeway(3 * time.Second)
		leewayOpt2 := WithLeeway(5 * time.Second)

		// Create authenticator with multiple options (last one should apply)
		authenticator := NewNodeJWTAuthenticator(mockProvider, logger, leewayOpt1, leewayOpt2)

		assert.NotNil(t, authenticator)
		assert.NotNil(t, authenticator.parser)
	})
}

func TestWithLeeway(t *testing.T) {
	t.Run("returns a valid option function", func(t *testing.T) {
		opt := WithLeeway(5 * time.Second)
		assert.NotNil(t, opt)
	})

	t.Run("option function modifies parser options slice", func(t *testing.T) {
		// Given: empty parser options slice
		parserOpts := []jwt.ParserOption{}

		// When: apply WithLeeway option
		opt := WithLeeway(10 * time.Second)
		opt(&parserOpts)

		// Expect: parser options slice is modified
		assert.Len(t, parserOpts, 1)
	})

	t.Run("option function can be applied multiple times", func(t *testing.T) {
		// Given: parser options slice
		parserOpts := []jwt.ParserOption{
			jwt.WithIssuedAt(),
		}
		initialLen := len(parserOpts)

		// When: apply multiple WithLeeway options
		opt1 := WithLeeway(5 * time.Second)
		opt2 := WithLeeway(10 * time.Second)
		opt1(&parserOpts)
		opt2(&parserOpts)

		// Expect: all options are added
		assert.Len(t, parserOpts, initialLen+2)
	})

	t.Run("different leeway durations can be configured", func(t *testing.T) {
		testCases := []struct {
			name     string
			duration time.Duration
		}{
			{"1 second", 1 * time.Second},
			{"5 seconds", 5 * time.Second},
			{"30 seconds", 30 * time.Second},
			{"1 minute", 1 * time.Minute},
			{"zero", 0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				opt := WithLeeway(tc.duration)
				assert.NotNil(t, opt)

				parserOpts := []jwt.ParserOption{}
				opt(&parserOpts)
				assert.Len(t, parserOpts, 1)
			})
		}
	})
}

func TestNodeJWTAuthenticator_WithoutLeeway_StrictValidation(t *testing.T) {
	t.Run("token expired by 1 second without leeway should be rejected", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		// No leeway option provided - strict validation
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token that expired 1 second ago
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Second)), // Expired 1s ago
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT without leeway
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Should fail as no leeway is configured
		require.Error(t, err)
		assert.False(t, valid)
		assert.NotNil(t, claims)
		assert.Contains(t, err.Error(), "token is expired")
	})

	t.Run("token issued 1 second in future without leeway should be rejected", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		// No leeway option provided - strict validation
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger())

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token issued 1 second in the future
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(workflowJWTExpiration)),
				IssuedAt:  jwt.NewNumericDate(now.Add(1 * time.Second)), // Issued 1s in future
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT without leeway
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Should fail as no leeway is configured
		require.Error(t, err)
		assert.False(t, valid)
		assert.NotNil(t, claims)
		assert.Contains(t, err.Error(), "used before issued")
	})
}

func TestNodeJWTAuthenticator_WithLeeway_CustomDurations(t *testing.T) {
	t.Run("custom leeway of 10 seconds allows token expired 8 seconds ago", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		mockProvider.On("IsNodePubKeyTrusted", mock.Anything, csaPubKey).Return(true, nil)
		// Custom 10 second leeway
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), WithLeeway(10*time.Second))

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token that expired 8 seconds ago (within 10s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(-8 * time.Second)), // Expired 8s ago
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Should succeed with 10s leeway
		require.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, claims)
		mockProvider.AssertExpectations(t)
	})

	t.Run("custom leeway of 2 seconds rejects token expired 3 seconds ago", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		// Small 2 second leeway
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), WithLeeway(2*time.Second))

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token that expired 3 seconds ago (beyond 2s leeway)
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(-3 * time.Second)), // Expired 3s ago
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Should fail as it's beyond 2s leeway
		require.Error(t, err)
		assert.False(t, valid)
		assert.NotNil(t, claims)
		assert.Contains(t, err.Error(), "token is expired")
	})

	t.Run("zero leeway behaves like no leeway option", func(t *testing.T) {
		// Given
		privateKey, csaPubKey := createValidatorTestKeys()
		mockProvider := &mocks.NodeAuthProvider{}
		// Explicitly set zero leeway
		authenticator := NewNodeJWTAuthenticator(mockProvider, createTestLogger(), WithLeeway(0*time.Second))

		testRequest := testRequest{Field: "test-request"}
		digest := utils.CalculateRequestDigest(testRequest)

		// Create a token that expired 1 second ago
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, types.NodeJWTClaims{
			PublicKey: hex.EncodeToString(csaPubKey),
			Digest:    digest,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    hex.EncodeToString(csaPubKey),
				Subject:   hex.EncodeToString(csaPubKey),
				ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Second)), // Expired 1s ago
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		})

		jwtToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// When: Authenticate JWT
		valid, claims, err := authenticator.AuthenticateJWT(context.Background(), jwtToken, testRequest)

		// Expect: Should fail with zero leeway
		require.Error(t, err)
		assert.False(t, valid)
		assert.NotNil(t, claims)
		assert.Contains(t, err.Error(), "token is expired")
	})
}

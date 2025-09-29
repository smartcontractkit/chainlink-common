package beholder_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func TestBuildAuthHeaders(t *testing.T) {
	csaPrivKeyHex := "1ac84741fa51c633845fa65c06f37a700303619135630a01f2d22fb98eb1c54ecab39509e63cfaa81c70e2c907391f96803aacb00db5619a5ace5588b4b08159"
	csaPrivKeyBytes, err := hex.DecodeString(csaPrivKeyHex)
	assert.NoError(t, err)
	csaPrivKey := ed25519.PrivateKey(csaPrivKeyBytes)

	expectedHeaders := map[string]string{
		"X-Beholder-Node-Auth-Token": "1:cab39509e63cfaa81c70e2c907391f96803aacb00db5619a5ace5588b4b08159:4403178e299e9acc5b48ae97de617d3975c5d431b794cfab1d23eda01c194119b2360f5f74cfb3e4f706237ab57a0ba88ffd3f8addbc1e5197b3d3e13a1fc409",
	}

	headers := beholder.BuildAuthHeaders(csaPrivKey)
	assert.Equal(t, expectedHeaders, headers)

	headers, err = beholder.NewAuthHeaders(csaPrivKey)
	require.NoError(t, err)
	assert.Equal(t, expectedHeaders, headers)
}

func TestStaticAuthHeaderProvider(t *testing.T) {
	// Create test headers
	testHeaders := map[string]string{
		"header1": "value1",
		"header2": "value2",
	}

	// Create new header provider
	provider := beholder.NewStaticAuth(testHeaders, false)

	// Get headers and verify they match
	headers, err := provider.Headers(t.Context())
	require.NoError(t, err)
	assert.Equal(t, testHeaders, headers)
}

// MockSigner implements the beholder.Signer interface for testing rotating auth
type MockSigner struct {
	mock.Mock
}

func (m *MockSigner) Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error) {
	args := m.Called(ctx, keyID, data)
	return args.Get(0).([]byte), args.Error(1)
}

func TestRotatingAuth(t *testing.T) {
	// Generate test key pair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	t.Run("creates valid rotating auth headers", func(t *testing.T) {

		mockSigner := &MockSigner{}

		dummySignature := ed25519.Sign(privKey, []byte("test data"))

		mockSigner.
			On("Sign", mock.Anything, mock.MatchedBy(func(keyID []byte) bool {
				return string(keyID) == string(pubKey) // Verify correct public key is passed
			}), mock.Anything).
			Return(dummySignature, nil)

		ttl := 5 * time.Minute
		auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false)

		headers, err := auth.Headers(t.Context())
		require.NoError(t, err)
		require.NotEmpty(t, headers)

		authHeader := headers["X-Beholder-Node-Auth-Token"]
		require.NotEmpty(t, authHeader)

		parts := strings.Split(authHeader, ":")
		require.Len(t, parts, 4, "Auth header should have format version:pubkey_hex:timestamp:signature_hex")

		assert.Equal(t, "2", parts[0], "Version should be 2")
		assert.Equal(t, hex.EncodeToString(pubKey), parts[1], "Public key should match")
		assert.NotEmpty(t, parts[2], "Timestamp should not be empty")

		// Verify signature is hex encoded
		_, err = hex.DecodeString(parts[3])
		assert.NoError(t, err, "Signature should be valid hex")

		mockSigner.AssertExpectations(t)
	})

	t.Run("reuses headers within TTL", func(t *testing.T) {

		mockSigner := &MockSigner{}

		dummySignature := ed25519.Sign(privKey, []byte("test data"))

		mockSigner.
			On("Sign", mock.Anything, mock.Anything, mock.Anything).
			Return(dummySignature, nil).
			Maybe()

		ttl := 5 * time.Minute
		auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false)

		headers1, err := auth.Headers(t.Context())
		require.NoError(t, err)

		headers2, err := auth.Headers(t.Context())
		require.NoError(t, err)

		assert.Equal(t, headers1, headers2, "Headers should be reused within TTL")

		mockSigner.AssertExpectations(t)
	})

	t.Run("handles signer errors", func(t *testing.T) {

		mockSigner := &MockSigner{}
		expectedErr := assert.AnError

		mockSigner.
			On("Sign", mock.Anything, mock.Anything, mock.Anything).
			Return([]byte{}, expectedErr)

		ttl := 5 * time.Minute
		auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false)

		headers, err := auth.Headers(t.Context())
		require.Error(t, err)
		assert.Nil(t, headers)
		assert.Contains(t, err.Error(), "beholder: failed to sign auth header")
		assert.Contains(t, err.Error(), expectedErr.Error())

		mockSigner.AssertExpectations(t)
	})

	t.Run("implements PerRPCCredentialsProvider interface", func(t *testing.T) {

		mockSigner := &MockSigner{}
		dummySignature := ed25519.Sign(privKey, []byte("test data"))

		mockSigner.
			On("Sign", mock.Anything, mock.Anything, mock.Anything).
			Return(dummySignature, nil).
			Maybe()

		ttl := 5 * time.Minute
		auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false)

		creds := auth.Credentials()
		require.NotNil(t, creds)

		assert.False(t, creds.RequireTransportSecurity())

		metadata, err := creds.GetRequestMetadata(t.Context())
		require.NoError(t, err)
		assert.NotEmpty(t, metadata)

		mockSigner.AssertExpectations(t)
	})

	t.Run("respects transport security requirement", func(t *testing.T) {

		mockSigner := &MockSigner{}
		dummySignature := ed25519.Sign(privKey, []byte("test data"))

		mockSigner.
			On("Sign", mock.Anything, mock.Anything, mock.Anything).
			Return(dummySignature, nil).
			Maybe()

		ttl := 5 * time.Minute
		// transport security required
		authSecure := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, true)
		credsSecure := authSecure.Credentials()
		assert.True(t, credsSecure.RequireTransportSecurity())
		// transport security not required
		authInsecure := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false)
		credsInsecure := authInsecure.Credentials()
		assert.False(t, credsInsecure.RequireTransportSecurity())

		mockSigner.AssertExpectations(t)
	})
}

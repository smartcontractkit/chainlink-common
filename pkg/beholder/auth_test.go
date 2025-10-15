package beholder_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
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

func (m *MockSigner) Sign(ctx context.Context, keyID string, data []byte) ([]byte, error) {
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
			On("Sign", mock.Anything, mock.MatchedBy(func(keyID string) bool {
				return keyID == hex.EncodeToString(pubKey) // Verify correct public key hex is passed
			}), mock.Anything).
			Return(dummySignature, nil)

		ttl := 5 * time.Minute
		auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, nil)

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
		auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, nil)

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
		auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, nil)

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
		auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, nil)

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
		authSecure := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, true, nil)
		credsSecure := authSecure.Credentials()
		assert.True(t, credsSecure.RequireTransportSecurity())
		// transport security not required
		authInsecure := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, nil)
		credsInsecure := authInsecure.Credentials()
		assert.False(t, credsInsecure.RequireTransportSecurity())

		mockSigner.AssertExpectations(t)
	})

	t.Run("uses initial headers until TTL expires", func(t *testing.T) {
		mockSigner := &MockSigner{}

		// Create initial headers with v2 format
		ts := time.Now()
		signature := ed25519.Sign(privKey, []byte("initial"))
		initialHeaders := map[string]string{
			"X-Beholder-Node-Auth-Token": "2:" + hex.EncodeToString(pubKey) + ":" + fmt.Sprintf("%d", ts.UnixNano()) + ":" + hex.EncodeToString(signature),
		}

		// Use a very short TTL so it expires quickly
		ttl := 1 * time.Millisecond
		auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, initialHeaders)

		// First call should return the initial headers without calling Sign
		headers1, err := auth.Headers(t.Context())
		require.NoError(t, err)
		assert.Equal(t, initialHeaders, headers1)

		// Wait for TTL to expire
		time.Sleep(5 * time.Millisecond)

		// Now the signer should be called to generate new headers
		newSignature := ed25519.Sign(privKey, []byte("new"))
		mockSigner.
			On("Sign", mock.Anything, mock.Anything, mock.Anything).
			Return(newSignature, nil).
			Once()

		headers2, err := auth.Headers(t.Context())
		require.NoError(t, err)
		assert.NotEqual(t, initialHeaders, headers2, "Should generate new headers after TTL expires")

		mockSigner.AssertExpectations(t)
	})
}

// BenchmarkRotatingAuth_Headers_CachedPath benchmarks the fast path where headers are cached and within TTL.
// This is the most common case in production.
func BenchmarkRotatingAuth_Headers_CachedPath(b *testing.B) {

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(b, err)

	mockSigner := &MockSigner{}
	dummySignature := ed25519.Sign(privKey, []byte("test data"))

	mockSigner.
		On("Sign", mock.Anything, mock.Anything, mock.Anything).
		Return(dummySignature, nil).
		Maybe()

	// Use a long TTL so headers don't expire during the benchmark
	ttl := 1 * time.Hour
	auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, nil)

	// Prime the cache by calling Headers once
	ctx := b.Context()
	_, err = auth.Headers(ctx)
	require.NoError(b, err)

	b.ReportAllocs()

	for b.Loop() {
		headers, err := auth.Headers(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if len(headers) == 0 {
			b.Fatal("expected non-empty headers")
		}
	}
}

// BenchmarkRotatingAuth_Headers_ExpiredPath benchmarks the slow path where headers need to be regenerated.
// This happens when TTL expires.
func BenchmarkRotatingAuth_Headers_ExpiredPath(b *testing.B) {

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(b, err)

	mockSigner := &MockSigner{}
	dummySignature := ed25519.Sign(privKey, []byte("test data"))

	mockSigner.
		On("Sign", mock.Anything, mock.Anything, mock.Anything).
		Return(dummySignature, nil).
		Maybe()

	// Use a TTL of 0 to force regeneration on every call
	ttl := 0 * time.Second
	auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, nil)

	ctx := b.Context()

	b.ReportAllocs()

	for b.Loop() {
		headers, err := auth.Headers(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if len(headers) == 0 {
			b.Fatal("expected non-empty headers")
		}
	}
}

// BenchmarkRotatingAuth_Headers_ParallelCached benchmarks concurrent access when headers are cached.
// This simulates multiple goroutines making concurrent requests with valid cached headers.
func BenchmarkRotatingAuth_Headers_ParallelCached(b *testing.B) {

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(b, err)

	mockSigner := &MockSigner{}
	dummySignature := ed25519.Sign(privKey, []byte("test data"))

	mockSigner.
		On("Sign", mock.Anything, mock.Anything, mock.Anything).
		Return(dummySignature, nil).
		Maybe()

	// Use a long TTL so headers don't expire during the benchmark
	ttl := 1 * time.Hour
	auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, nil)

	// Prime the cache
	ctx := b.Context()
	_, err = auth.Headers(ctx)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			headers, err := auth.Headers(ctx)
			if err != nil {
				b.Fatal(err)
			}
			if len(headers) == 0 {
				b.Fatal("expected non-empty headers")
			}
		}
	})
}

// BenchmarkRotatingAuth_Headers_ParallelExpired benchmarks concurrent access when headers expire.
// This tests contention on the mutex when multiple goroutines race to regenerate headers.
func BenchmarkRotatingAuth_Headers_ParallelExpired(b *testing.B) {

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(b, err)

	mockSigner := &MockSigner{}
	dummySignature := ed25519.Sign(privKey, []byte("test data"))

	mockSigner.
		On("Sign", mock.Anything, mock.Anything, mock.Anything).
		Return(dummySignature, nil).
		Maybe()

	// Use a short TTL to cause periodic regeneration
	ttl := 10 * time.Millisecond
	auth := beholder.NewRotatingAuth(pubKey, mockSigner, ttl, false, nil)

	ctx := b.Context()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			headers, err := auth.Headers(ctx)
			if err != nil {
				b.Fatal(err)
			}
			if len(headers) == 0 {
				b.Fatal("expected non-empty headers")
			}
		}
	})
}

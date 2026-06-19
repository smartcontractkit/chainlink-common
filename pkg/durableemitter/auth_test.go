package durableemitter_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/durableemitter"
)

// globalSigner tests share process-wide state via globalSigner; serialize them.
var globalSignerTestMu sync.Mutex

func withIsolatedGlobalSigner(t *testing.T) {
	t.Helper()
	globalSignerTestMu.Lock()
	durableemitter.ResetGlobalSignerForTest()
	t.Cleanup(func() {
		durableemitter.ResetGlobalSignerForTest()
		globalSignerTestMu.Unlock()
	})
}

type mockSigner struct{}

func (mockSigner) Sign(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return []byte("signature"), nil
}

func TestNewAuthHeaderProvider_Static(t *testing.T) {
	t.Parallel()

	headers := map[string]string{"X-Node-Auth-Token": "static"}
	provider, err := durableemitter.NewAuthHeaderProvider(durableemitter.AuthConfig{
		AuthHeaders: headers,
	})
	require.NoError(t, err)

	got, err := provider.Headers(t.Context())
	require.NoError(t, err)
	require.Equal(t, headers, got)
}

func TestNewAuthHeaderProvider_RotatingDeferredSigner(t *testing.T) {
	withIsolatedGlobalSigner(t)

	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	initial := map[string]string{"X-Node-Auth-Token": "initial"}
	provider, err := durableemitter.NewAuthHeaderProvider(durableemitter.AuthConfig{
		AuthHeaders:      initial,
		AuthHeadersTTL:   10 * time.Minute,
		AuthPublicKeyHex: hex.EncodeToString(pubKey),
	})
	require.NoError(t, err)
	require.False(t, durableemitter.IsGlobalSignerSet())

	got, err := provider.Headers(t.Context())
	require.NoError(t, err)
	require.Equal(t, initial, got)

	durableemitter.SetGlobalSigner(mockSigner{})
	require.True(t, durableemitter.IsGlobalSignerSet())
}

func TestNewAuthHeaderProvider_RotatingWithSigner(t *testing.T) {
	withIsolatedGlobalSigner(t)

	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	provider, err := durableemitter.NewAuthHeaderProvider(durableemitter.AuthConfig{
		AuthHeaders:      map[string]string{"X-Node-Auth-Token": "initial"},
		AuthHeadersTTL:   10 * time.Minute,
		AuthPublicKeyHex: hex.EncodeToString(pubKey),
		AuthKeySigner:    mockSigner{},
	})
	require.NoError(t, err)
	require.True(t, durableemitter.IsGlobalSignerSet())

	got, err := provider.Headers(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, got["X-Node-Auth-Token"])
}

func TestNewAuthHeaderProvider_Validation(t *testing.T) {
	t.Parallel()

	_, err := durableemitter.NewAuthHeaderProvider(durableemitter.AuthConfig{
		AuthHeadersTTL: 10 * time.Minute,
	})
	require.ErrorContains(t, err, "public key hex required")

	_, err = durableemitter.NewAuthHeaderProvider(durableemitter.AuthConfig{
		AuthHeadersTTL:   time.Minute,
		AuthPublicKeyHex: "abcd",
	})
	require.ErrorContains(t, err, "at least 10 minutes")
}

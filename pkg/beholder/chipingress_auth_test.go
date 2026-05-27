package beholder_test

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func TestNewChipIngressHeaderProvider(t *testing.T) {
	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	pubKeyHex := hex.EncodeToString(pubKey)

	t.Run("returns nil provider when no auth configured", func(t *testing.T) {
		cfg := beholder.Config{}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		assert.Nil(t, provider)
	})

	t.Run("returns nil provider when TTL is zero and no headers", func(t *testing.T) {
		cfg := beholder.Config{
			AuthHeadersTTL: 0,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		assert.Nil(t, provider)
	})

	t.Run("returns static auth when headers set but TTL is zero", func(t *testing.T) {
		cfg := beholder.Config{
			AuthHeaders: map[string]string{
				"Authorization": "Bearer token",
			},
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		headers, err := provider.Headers(t.Context())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"Authorization": "Bearer token"}, headers)
	})

	t.Run("static auth respects transport security", func(t *testing.T) {
		cfg := beholder.Config{
			AuthHeaders: map[string]string{
				"Authorization": "Bearer token",
			},
			InsecureConnection: false, // requires TLS
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		auth, ok := provider.(beholder.Auth)
		require.True(t, ok)
		assert.True(t, auth.Credentials().RequireTransportSecurity())
	})

	t.Run("static auth does not require transport security when insecure", func(t *testing.T) {
		cfg := beholder.Config{
			AuthHeaders: map[string]string{
				"Authorization": "Bearer token",
			},
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		auth, ok := provider.(beholder.Auth)
		require.True(t, ok)
		assert.False(t, auth.Credentials().RequireTransportSecurity())
	})

	t.Run("returns rotating auth when TTL > 0 with valid config", func(t *testing.T) {
		mockSigner := &MockSigner{}

		cfg := beholder.Config{
			AuthPublicKeyHex:   pubKeyHex,
			AuthKeySigner:      mockSigner,
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		auth, ok := provider.(beholder.Auth)
		require.True(t, ok)
		assert.False(t, auth.Credentials().RequireTransportSecurity())
	})

	t.Run("rotating auth requires transport security when not insecure", func(t *testing.T) {
		mockSigner := &MockSigner{}

		cfg := beholder.Config{
			AuthPublicKeyHex:   pubKeyHex,
			AuthKeySigner:      mockSigner,
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: false,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		auth, ok := provider.(beholder.Auth)
		require.True(t, ok)
		assert.True(t, auth.Credentials().RequireTransportSecurity())
	})

	t.Run("rotating auth without AuthKeySigner still succeeds", func(t *testing.T) {
		cfg := beholder.Config{
			AuthPublicKeyHex:   pubKeyHex,
			AuthKeySigner:      nil, // signer injected later
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("error when TTL > 0 but public key hex is empty", func(t *testing.T) {
		cfg := beholder.Config{
			AuthPublicKeyHex:   "",
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: public key hex required for rotating auth (TTL > 0)")
	})

	t.Run("error when TTL is below 10 minutes", func(t *testing.T) {
		cfg := beholder.Config{
			AuthPublicKeyHex:   pubKeyHex,
			AuthHeadersTTL:     5 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: headers TTL must be at least 10 minutes")
	})

	t.Run("error when TTL is exactly 1 minute", func(t *testing.T) {
		cfg := beholder.Config{
			AuthPublicKeyHex:   pubKeyHex,
			AuthHeadersTTL:     time.Minute,
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: headers TTL must be at least 10 minutes")
	})

	t.Run("succeeds when TTL is exactly 10 minutes", func(t *testing.T) {
		cfg := beholder.Config{
			AuthPublicKeyHex:   pubKeyHex,
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("error when public key hex is invalid", func(t *testing.T) {
		cfg := beholder.Config{
			AuthPublicKeyHex:   "not-valid-hex!",
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: failed to decode public key hex")
	})

	t.Run("error when public key hex has odd length", func(t *testing.T) {
		cfg := beholder.Config{
			AuthPublicKeyHex:   "abc", // odd-length hex
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: failed to decode public key hex")
	})

	t.Run("rotating auth takes precedence over static headers", func(t *testing.T) {
		// When both AuthHeadersTTL > 0 and AuthHeaders are set,
		// rotating auth is returned (AuthHeaders are passed as initial headers)
		mockSigner := &MockSigner{}

		cfg := beholder.Config{
			AuthPublicKeyHex: pubKeyHex,
			AuthKeySigner:    mockSigner,
			AuthHeadersTTL:   10 * time.Minute,
			AuthHeaders: map[string]string{
				"Authorization": "Bearer static-token",
			},
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		// Should return initial headers (the static ones passed as initial)
		headers, err := provider.Headers(t.Context())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"Authorization": "Bearer static-token"}, headers)
	})

	t.Run("negative TTL treated as no rotating auth", func(t *testing.T) {
		cfg := beholder.Config{
			AuthHeadersTTL:     -1 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := beholder.NewChipIngressHeaderProvider(cfg)
		require.NoError(t, err)
		assert.Nil(t, provider)
	})
}

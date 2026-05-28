package chipingress_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// fakeSigner is a minimal chipingress.Signer used by tests that need to
// construct a rotating provider. It is not invoked unless headers are actually
// rotated, so its return value rarely matters here.
type fakeSigner struct{}

func (fakeSigner) Sign(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return []byte("signature"), nil
}

func TestNewHeaderProvider(t *testing.T) {
	// tsr is an inline interface used to assert the TLS requirement on
	// providers returned by NewHeaderProvider without depending on an
	// exported type assertion helper.
	type tsr interface {
		RequireTransportSecurity() bool
	}

	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	pubKeyHex := hex.EncodeToString(pubKey)

	t.Run("returns nil provider when no auth configured", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		assert.Nil(t, provider)
	})

	t.Run("returns nil provider when TTL is zero and no headers", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthHeadersTTL: 0,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		assert.Nil(t, provider)
	})

	t.Run("returns static auth when headers set but TTL is zero", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthHeaders: map[string]string{
				"Authorization": "Bearer token",
			},
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		headers, err := provider.Headers(t.Context())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"Authorization": "Bearer token"}, headers)
	})

	t.Run("static auth respects transport security", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthHeaders: map[string]string{
				"Authorization": "Bearer token",
			},
			InsecureConnection: false, // requires TLS
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		requirer, ok := provider.(tsr)
		require.True(t, ok)
		assert.True(t, requirer.RequireTransportSecurity())
	})

	t.Run("static auth does not require transport security when insecure", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthHeaders: map[string]string{
				"Authorization": "Bearer token",
			},
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		requirer, ok := provider.(tsr)
		require.True(t, ok)
		assert.False(t, requirer.RequireTransportSecurity())
	})

	t.Run("returns rotating auth when TTL > 0 with valid config", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex:   pubKeyHex,
			AuthKeySigner:      fakeSigner{},
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		requirer, ok := provider.(tsr)
		require.True(t, ok)
		assert.False(t, requirer.RequireTransportSecurity())
	})

	t.Run("rotating auth requires transport security when not insecure", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex:   pubKeyHex,
			AuthKeySigner:      fakeSigner{},
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: false,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		requirer, ok := provider.(tsr)
		require.True(t, ok)
		assert.True(t, requirer.RequireTransportSecurity())
	})

	t.Run("rotating auth without AuthKeySigner still succeeds", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex:   pubKeyHex,
			AuthKeySigner:      nil, // signer injected later
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("error when TTL > 0 but public key hex is empty", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex:   "",
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: public key hex required for rotating auth (TTL > 0)")
	})

	t.Run("error when TTL is below 10 minutes", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex:   pubKeyHex,
			AuthHeadersTTL:     5 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: headers TTL must be at least 10 minutes")
	})

	t.Run("error when TTL is exactly 1 minute", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex:   pubKeyHex,
			AuthHeadersTTL:     time.Minute,
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: headers TTL must be at least 10 minutes")
	})

	t.Run("succeeds when TTL is exactly 10 minutes", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex:   pubKeyHex,
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	t.Run("error when public key hex is invalid", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex:   "not-valid-hex!",
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: failed to decode public key hex")
	})

	t.Run("error when public key hex has odd length", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex:   "abc", // odd-length hex
			AuthHeadersTTL:     10 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "auth: failed to decode public key hex")
	})

	t.Run("rotating auth takes precedence over static headers", func(t *testing.T) {
		// When both AuthHeadersTTL > 0 and AuthHeaders are set, rotating auth
		// is returned (AuthHeaders are passed as initial headers).
		cfg := chipingress.HeaderProviderConfig{
			AuthPublicKeyHex: pubKeyHex,
			AuthKeySigner:    fakeSigner{},
			AuthHeadersTTL:   10 * time.Minute,
			AuthHeaders: map[string]string{
				"Authorization": "Bearer static-token",
			},
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		headers, err := provider.Headers(t.Context())
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"Authorization": "Bearer static-token"}, headers)
	})

	t.Run("negative TTL treated as no rotating auth", func(t *testing.T) {
		cfg := chipingress.HeaderProviderConfig{
			AuthHeadersTTL:     -1 * time.Minute,
			InsecureConnection: true,
		}

		provider, err := chipingress.NewHeaderProvider(cfg)
		require.NoError(t, err)
		assert.Nil(t, provider)
	})
}

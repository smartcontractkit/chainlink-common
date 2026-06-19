package durableemitter

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// Signer signs auth header payloads using the node's CSA key. It is an alias of
// chipingress.Signer so DurableEmitter callers don't need to import chipingress
// directly.
type Signer = chipingress.Signer

// AuthConfig configures chip ingress auth headers for DurableEmitter clients.
type AuthConfig struct {
	AuthHeaders      map[string]string
	AuthHeadersTTL   time.Duration
	AuthPublicKeyHex string
	// AuthKeySigner may be nil at init time for LOOP plugins; call SetGlobalSigner
	// after the CSA keystore is available.
	AuthKeySigner Signer
}

// lazySigner is a thread-safe wrapper that lets the CSA keystore be injected
// after the header provider is built. LOOP plugins start with rotating auth
// configured but receive the keystore later via SetGlobalSigner. The zero value
// is usable.
type lazySigner struct {
	signer atomic.Pointer[Signer]
}

func (l *lazySigner) Sign(ctx context.Context, keyID string, data []byte) ([]byte, error) {
	s := l.signer.Load()
	if s == nil {
		return nil, errors.New("keystore not yet available for signing")
	}
	return (*s).Sign(ctx, keyID, data)
}

func (l *lazySigner) Set(signer Signer) { l.signer.Store(&signer) }

func (l *lazySigner) IsSet() bool { return l.signer.Load() != nil }

// globalSigner holds the lazy wrapper backing the process-wide rotating auth
// provider, set by NewAuthHeaderProvider when rotating auth is configured.
var globalSigner atomic.Pointer[lazySigner]

// NewAuthHeaderProvider builds a chip ingress HeaderProvider for DurableEmitter
// clients, delegating the static/rotating provider logic to
// chipingress.NewHeaderProvider.
//
// For rotating auth (AuthHeadersTTL > 0) the signer is wrapped in a lazy holder
// so the CSA keystore can be injected after startup via SetGlobalSigner.
func NewAuthHeaderProvider(cfg AuthConfig) (chipingress.HeaderProvider, error) {
	lazy := &lazySigner{}
	if cfg.AuthKeySigner != nil {
		lazy.Set(cfg.AuthKeySigner)
	}

	provider, err := chipingress.NewHeaderProvider(chipingress.HeaderProviderConfig{
		AuthHeaders:      cfg.AuthHeaders,
		AuthHeadersTTL:   cfg.AuthHeadersTTL,
		AuthPublicKeyHex: cfg.AuthPublicKeyHex,
		AuthKeySigner:    lazy,
	})
	if err != nil {
		return nil, err
	}

	// Only rotating providers consult the signer; publish the lazy wrapper so
	// SetGlobalSigner can inject the keystore once it becomes available.
	if cfg.AuthHeadersTTL > 0 {
		globalSigner.Store(lazy)
	}

	return provider, nil
}

// SetGlobalSigner injects the CSA keystore used to refresh rotating chip ingress
// auth headers. No-op when rotating auth is not configured.
func SetGlobalSigner(signer Signer) {
	if lazy := globalSigner.Load(); lazy != nil {
		lazy.Set(signer)
	}
}

// IsGlobalSignerSet reports whether rotating DurableEmitter auth has a signer configured.
func IsGlobalSignerSet() bool {
	lazy := globalSigner.Load()
	return lazy != nil && lazy.IsSet()
}

// ResetGlobalSignerForTest clears the process-wide rotating-auth signer holder.
// It is intended for tests that assert on global signer state.
func ResetGlobalSignerForTest() {
	globalSigner.Store(nil)
}

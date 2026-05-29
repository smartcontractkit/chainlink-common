package durableemitter

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

const (
	authHeaderKey     = "X-Beholder-Node-Auth-Token"
	authHeaderVersion = "2"
)

// Signer signs auth header payloads using the node's CSA key.
type Signer interface {
	Sign(ctx context.Context, keyID string, data []byte) ([]byte, error)
}

// AuthConfig configures chip ingress auth headers for DurableEmitter clients.
type AuthConfig struct {
	AuthHeaders      map[string]string
	AuthHeadersTTL   time.Duration
	AuthPublicKeyHex string
	// AuthKeySigner may be nil at init time for LOOP plugins; call SetGlobalSigner
	// after the CSA keystore is available.
	AuthKeySigner Signer
}

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

func (l *lazySigner) Set(signer Signer) {
	l.signer.Store(&signer)
}

func (l *lazySigner) IsSet() bool {
	return l.signer.Load() != nil
}

type staticHeaderProvider struct {
	headers map[string]string
}

func (p *staticHeaderProvider) Headers(_ context.Context) (map[string]string, error) {
	return p.headers, nil
}

type rotatingHeaderProvider struct {
	csaPubKey        ed25519.PublicKey
	signer           Signer
	signerTimeout    time.Duration
	headers          atomic.Value // map[string]string
	ttl              time.Duration
	lastUpdatedNanos atomic.Int64
	lazy             *lazySigner
	mu               sync.Mutex
}

func (p *rotatingHeaderProvider) SetSigner(signer Signer) {
	if p.lazy != nil {
		p.lazy.Set(signer)
	}
}

func (p *rotatingHeaderProvider) IsSignerSet() bool {
	return p.lazy != nil && p.lazy.IsSet()
}

func (p *rotatingHeaderProvider) Headers(ctx context.Context) (map[string]string, error) {
	returnHeader := make(map[string]string)
	lastUpdated := time.Unix(0, p.lastUpdatedNanos.Load())

	if time.Since(lastUpdated) > p.ttl {
		p.mu.Lock()
		defer p.mu.Unlock()

		lastUpdated = time.Unix(0, p.lastUpdatedNanos.Load())
		if time.Since(lastUpdated) < p.ttl {
			maps.Copy(returnHeader, p.headers.Load().(map[string]string))
			return returnHeader, nil
		}

		ctxWithTimeout, cancel := context.WithTimeout(ctx, p.signerTimeout)
		defer cancel()

		ts := time.Now()
		newHeaders, err := newAuthHeaderV2(ctxWithTimeout, p.csaPubKey, p.signer, ts)
		if err != nil {
			return nil, fmt.Errorf("durableemitter: failed to create auth header: %w", err)
		}

		p.headers.Store(newHeaders)
		p.lastUpdatedNanos.Store(ts.UnixNano())
	}

	maps.Copy(returnHeader, p.headers.Load().(map[string]string))
	return returnHeader, nil
}

var globalRotatingAuth atomic.Pointer[rotatingHeaderProvider]

// NewAuthHeaderProvider builds a HeaderProvider for DurableEmitter chip ingress clients.
//
// Static mode (AuthHeadersTTL == 0): returns fixed AuthHeaders.
// Rotating mode (AuthHeadersTTL > 0): uses initial AuthHeaders until TTL expires, then signs fresh headers.
func NewAuthHeaderProvider(cfg AuthConfig) (chipingress.HeaderProvider, error) {
	if cfg.AuthHeadersTTL > 0 {
		if cfg.AuthPublicKeyHex == "" {
			return nil, errors.New("auth: public key hex required for rotating auth (TTL > 0)")
		}
		if cfg.AuthHeadersTTL < 10*time.Minute {
			return nil, errors.New("auth: headers TTL must be at least 10 minutes")
		}

		pubKey, err := hex.DecodeString(cfg.AuthPublicKeyHex)
		if err != nil {
			return nil, fmt.Errorf("auth: failed to decode public key hex: %w", err)
		}

		lazy := &lazySigner{}
		if cfg.AuthKeySigner != nil {
			lazy.Set(cfg.AuthKeySigner)
		}

		provider := &rotatingHeaderProvider{
			csaPubKey:     ed25519.PublicKey(pubKey),
			signer:        lazy,
			signerTimeout: 5 * time.Second,
			ttl:           cfg.AuthHeadersTTL,
			lazy:          lazy,
		}

		headers := make(map[string]string)
		if len(cfg.AuthHeaders) > 0 {
			headers = cfg.AuthHeaders
			provider.lastUpdatedNanos.Store(time.Now().UnixNano())
		}
		provider.headers.Store(headers)

		globalRotatingAuth.Store(provider)
		return provider, nil
	}

	if len(cfg.AuthHeaders) == 0 {
		return nil, nil
	}
	return &staticHeaderProvider{headers: cfg.AuthHeaders}, nil
}

// SetGlobalSigner injects the CSA keystore used to refresh rotating chip ingress auth headers.
// No-op when rotating auth is not configured.
func SetGlobalSigner(signer Signer) {
	if provider := globalRotatingAuth.Load(); provider != nil {
		provider.SetSigner(signer)
	}
}

// IsGlobalSignerSet reports whether rotating DurableEmitter auth has a signer configured.
func IsGlobalSignerSet() bool {
	provider := globalRotatingAuth.Load()
	return provider != nil && provider.IsSignerSet()
}

func newAuthHeaderV2(ctx context.Context, pubKey ed25519.PublicKey, signer Signer, ts time.Time) (map[string]string, error) {
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(ts.UnixNano()))
	msgBytes := append(pubKey, tsBytes...)

	signature, err := signer.Sign(ctx, fmt.Sprintf("%x", pubKey), msgBytes)
	if err != nil {
		return nil, fmt.Errorf("durableemitter: failed to sign auth header: %w", err)
	}

	return map[string]string{
		authHeaderKey: fmt.Sprintf("%s:%x:%d:%x", authHeaderVersion, pubKey, ts.UnixNano(), signature),
	}, nil
}

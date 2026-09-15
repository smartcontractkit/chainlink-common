package chipingress

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// authHeaderKey is the header name used to carry the auth token. The value is
// preserved verbatim from the original beholder-side implementation to keep
// the wire protocol unchanged.
const (
	authHeaderKey = "X-Beholder-Node-Auth-Token"
	authHeaderV2  = "2"
)

// Signer is the minimal signing interface needed by the rotating header
// provider. It is structurally identical to beholder.Signer; defining it here
// avoids importing pkg/beholder (which lives in a different Go module and
// would invert the dependency edge).
type Signer interface {
	Sign(ctx context.Context, keyID string, data []byte) ([]byte, error)
}

// HeaderProviderConfig captures the inputs needed by NewHeaderProvider.
type HeaderProviderConfig struct {
	// AuthHeaders are returned as-is for static auth, or used as the initial
	// headers for rotating auth until the first rotation occurs.
	AuthHeaders map[string]string

	// AuthHeadersTTL > 0 selects the rotating provider. Must be >= 10 minutes
	// when set. TTL <= 0 selects the static (or nil) provider.
	AuthHeadersTTL time.Duration

	// AuthPublicKeyHex is the hex-encoded ed25519 public key. Required when
	// AuthHeadersTTL > 0.
	AuthPublicKeyHex string

	// AuthKeySigner is used by the rotating provider to sign refreshed
	// headers. May be nil at construction time if the signer will be injected
	// later (e.g. via a lazy wrapper held by the caller).
	AuthKeySigner Signer

	// InsecureConnection, when true, indicates the resulting provider does
	// not require TLS.
	InsecureConnection bool
}

// NewHeaderProvider creates a HeaderProvider from cfg.
//
// Selection rules — these match the inline switch in pkg/beholder/client.go
// that wires the chipingress emitter's auth, so a chipingress.HeaderProvider
// built here is observationally equivalent to the one beholder builds from
// the corresponding fields on beholder.Config:
//
//	beholder.Config field             chipingress.HeaderProviderConfig field
//	------------------------------    --------------------------------------
//	AuthHeaders                       AuthHeaders
//	AuthHeadersTTL                    AuthHeadersTTL
//	AuthPublicKeyHex                  AuthPublicKeyHex
//	AuthKeySigner                     AuthKeySigner
//	ChipIngressInsecureConnection     InsecureConnection
//
// Resulting provider:
//
//   - AuthHeadersTTL > 0: returns a rotating provider. Requires
//     AuthPublicKeyHex and AuthHeadersTTL >= 10 minutes.
//   - AuthHeadersTTL == 0 and len(AuthHeaders) > 0: returns a static provider.
//   - Otherwise: returns (nil, nil).
func NewHeaderProvider(cfg HeaderProviderConfig) (HeaderProvider, error) {
	if cfg.AuthHeadersTTL > 0 {
		if cfg.AuthPublicKeyHex == "" {
			return nil, errors.New("auth: public key hex required for rotating auth (TTL > 0)")
		}
		if cfg.AuthHeadersTTL < 10*time.Minute {
			return nil, errors.New("auth: headers TTL must be at least 10 minutes")
		}
		key, err := hex.DecodeString(cfg.AuthPublicKeyHex)
		if err != nil {
			return nil, fmt.Errorf("auth: failed to decode public key hex: %w", err)
		}
		return newRotatingHeaderProvider(
			key,
			cfg.AuthKeySigner,
			cfg.AuthHeadersTTL,
			!cfg.InsecureConnection,
			cfg.AuthHeaders,
		), nil
	}

	if len(cfg.AuthHeaders) > 0 {
		return newStaticHeaderProvider(cfg.AuthHeaders, !cfg.InsecureConnection), nil
	}

	return nil, nil
}

// newStaticHeaderProvider returns a HeaderProvider that always returns the
// given headers.
func newStaticHeaderProvider(headers map[string]string, requireTLS bool) HeaderProvider {
	return &staticHeaderProvider{headers: headers, requireTLS: requireTLS}
}

// NewStaticHeaderProvider returns a HeaderProvider that always returns the given headers,
// for use with WithHeaderProvider to attach fixed, non-auth gRPC metadata (e.g. resource
// attributes) to every request.
//
// This is for the non-auth interceptor path only. It reports RequireTransportSecurity() == false,
// which WithHeaderProvider never consults — the HeaderProvider interface declares only Headers,
// and grpc asks only credentials.PerRPCCredentials about transport security. Do not pass the
// result to WithTokenAuth: that path takes its TLS requirement from the client config
// (!c.insecureConnection), not from the provider, so the false here would be silently ignored
// rather than honoured. Use NewHeaderProvider for auth headers.
func NewStaticHeaderProvider(headers map[string]string) HeaderProvider {
	return newStaticHeaderProvider(headers, false)
}

// isPrintableASCII reports whether every byte of val is in the printable ASCII range [0x20, 0x7E].
// grpc-go hard-fails the entire RPC — auth header included — when an outgoing metadata value fails
// this check, so a value that does not pass is omitted rather than rewritten: a byte-mangled value
// is a worse outcome than a dropped attribute for an operator-facing observability field.
func isPrintableASCII(val string) bool {
	for i := 0; i < len(val); i++ {
		if c := val[i]; c < 0x20 || c > 0x7E {
			return false
		}
	}
	return true
}

// SanitizeMetadataHeaders projects a map of resource attributes onto the closed whitelist defined
// by ResourceAttributeHeaders, returning the gRPC metadata headers to attach to every request
// (e.g. via NewStaticHeaderProvider). An attribute named csa_public_key is emitted as
// chainlink-resource-csa-public-key; chip-ingress reads exactly those fixed header names and forwards them
// onto every Kafka record a request produces under resource_<original attribute key>.
//
// The whitelist is what makes this safe without validation machinery. Key matching is
// case-insensitive against the fixed set, so no operator-defined key can ever become a header name:
// an attribute named X-Beholder-Node-Auth-Token is simply not in the whitelist and is ignored,
// which is how the CSA auth token stays out of reach without a deny-list (the header interceptor
// appends to outgoing metadata, so an attribute that could land on the auth header's name would
// break authentication by sending a second value under it).
//
// An attribute is omitted, never rewritten, when its key is not whitelisted (regardless of case)
// or its value is not printable ASCII (isPrintableASCII). Keys are processed in sorted order so
// that if case variants of one whitelisted attribute collide on the same header name, the first in
// sorted order of the original keys wins, deterministically.
func SanitizeMetadataHeaders(in map[string]string) map[string]string {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make(map[string]string, len(ResourceAttributeHeaders))
	for _, k := range keys {
		header, ok := ResourceAttributeHeaders[strings.ToLower(k)]
		if !ok {
			continue
		}
		if _, dup := out[header]; dup {
			continue
		}
		if !isPrintableASCII(in[k]) {
			continue
		}
		out[header] = in[k]
	}
	return out
}

// newRotatingHeaderProvider returns a HeaderProvider that refreshes its
// headers every ttl using signer. initialHeaders, if non-empty, are served
// until the first rotation occurs.
func newRotatingHeaderProvider(
	pubKey ed25519.PublicKey,
	signer Signer,
	ttl time.Duration,
	requireTLS bool,
	initialHeaders map[string]string,
) HeaderProvider {
	r := &rotatingHeaderProvider{
		pubKey:        pubKey,
		signer:        signer,
		signerTimeout: 5 * time.Second,
		ttl:           ttl,
		requireTLS:    requireTLS,
	}

	headers := make(map[string]string)
	if len(initialHeaders) > 0 {
		headers = initialHeaders
		// Assume the initial headers were generated approximately "now".
		r.lastUpdatedNanos.Store(time.Now().UnixNano())
	}
	r.headers.Store(headers)

	return r
}

// staticHeaderProvider serves a fixed set of headers.
type staticHeaderProvider struct {
	headers    map[string]string
	requireTLS bool
}

func (s *staticHeaderProvider) Headers(_ context.Context) (map[string]string, error) {
	return s.headers, nil
}

func (s *staticHeaderProvider) RequireTransportSecurity() bool {
	return s.requireTLS
}

// rotatingHeaderProvider refreshes its headers when ttl has elapsed by
// invoking signer to produce a new V2 auth header.
type rotatingHeaderProvider struct {
	pubKey           ed25519.PublicKey
	signer           Signer
	signerTimeout    time.Duration
	headers          atomic.Value // map[string]string
	ttl              time.Duration
	lastUpdatedNanos atomic.Int64
	requireTLS       bool
	mu               sync.Mutex
}

func (r *rotatingHeaderProvider) Headers(ctx context.Context) (map[string]string, error) {
	returnHeader := make(map[string]string)
	lastUpdated := time.Unix(0, r.lastUpdatedNanos.Load())

	if time.Since(lastUpdated) > r.ttl {
		r.mu.Lock()
		defer r.mu.Unlock()

		// Double-check after acquiring the lock in case another goroutine
		// already refreshed.
		lastUpdated = time.Unix(0, r.lastUpdatedNanos.Load())
		if time.Since(lastUpdated) < r.ttl {
			maps.Copy(returnHeader, r.headers.Load().(map[string]string))
			return returnHeader, nil
		}

		ctxWithTimeout, cancel := context.WithTimeout(ctx, r.signerTimeout)
		defer cancel()

		ts := time.Now()
		newHeaders, err := newAuthHeaderV2(ctxWithTimeout, r.pubKey, r.signer, ts)
		if err != nil {
			return nil, fmt.Errorf("chipingress: failed to create auth header: %w", err)
		}

		r.headers.Store(newHeaders)
		r.lastUpdatedNanos.Store(ts.UnixNano())
	}

	maps.Copy(returnHeader, r.headers.Load().(map[string]string))
	return returnHeader, nil
}

func (r *rotatingHeaderProvider) RequireTransportSecurity() bool {
	return r.requireTLS
}

// newAuthHeaderV2 creates the V2 auth header value. The signed message is the
// concatenation of the public key bytes and the big-endian uint64 nanosecond
// timestamp. The header format is:
//
//	<version>:<public_key_hex>:<timestamp_nanos>:<signature_hex>
func newAuthHeaderV2(ctx context.Context, pubKey ed25519.PublicKey, signer Signer, ts time.Time) (map[string]string, error) {
	if signer == nil {
		return nil, errors.New("chipingress: signer is nil")
	}

	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(ts.UnixNano()))
	msgBytes := append(pubKey, tsBytes...)

	signature, err := signer.Sign(ctx, fmt.Sprintf("%x", pubKey), msgBytes)
	if err != nil {
		return nil, fmt.Errorf("chipingress: failed to sign auth header: %w", err)
	}

	return map[string]string{
		authHeaderKey: fmt.Sprintf("%s:%x:%d:%x", authHeaderV2, pubKey, ts.UnixNano(), signature),
	}, nil
}

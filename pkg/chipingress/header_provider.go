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
func NewStaticHeaderProvider(headers map[string]string) HeaderProvider {
	return newStaticHeaderProvider(headers, false)
}

// SanitizeMetadataValue replaces any byte outside the printable ASCII range [0x20-0x7E]
// with '?'. grpc-go hard-fails the entire RPC when an outgoing metadata value fails this
// check (unlike the CE-extension path, where an invalid entry is simply dropped), so
// values headed for gRPC metadata must be normalized before being sent.
func SanitizeMetadataValue(val string) string {
	b := []byte(val)
	out := make([]byte, len(b))
	for i, c := range b {
		if c >= 0x20 && c <= 0x7E {
			out[i] = c
		} else {
			out[i] = '?'
		}
	}
	return string(out)
}

// SanitizeMetadataHeaders sanitizes a map of resource-attribute headers for use as outgoing
// gRPC metadata (e.g. via NewStaticHeaderProvider). Keys are sanitized with
// SanitizeExtensionName — the same strict [a-z0-9] charset used for CloudEvent extensions —
// which is a subset of grpc's allowed metadata-key charset, so a sanitized key can never trip
// grpc's key validation or the reserved "-bin" suffix, and produces the same key stem as the
// corresponding CE extension (differing only by the CloudEvents Kafka binding's "ce_" prefix
// once on the wire). Values are sanitized via SanitizeMetadataValue, since grpc-go fails the
// whole RPC on a non-printable value. Entries that sanitize to an empty key, or that collide
// with a reserved extension name (see reservedExtensionNames) or a gRPC-reserved header name
// (see reservedMetadataKeys), are skipped. Keys are applied in sorted order so duplicate
// sanitized keys resolve deterministically (first in sorted order wins), matching
// WithResourceAttributeExtensions' collision handling.
//
// Note: unlike the CloudEvents Kafka binding, gRPC metadata keys are NOT prefixed with "ce_" —
// that prefix is a CloudEvents-binding concept, not a metadata one, and reusing it here would
// collide with the CE binding's own "ce_<name>" Kafka header if the server ever forwards gRPC
// metadata verbatim onto Kafka.
func SanitizeMetadataHeaders(in map[string]string) map[string]string {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make(map[string]string, len(in))
	for _, k := range keys {
		name := SanitizeExtensionName(k)
		if name == "" {
			continue
		}
		if _, reserved := reservedExtensionNames[name]; reserved {
			continue
		}
		if _, reserved := reservedMetadataKeys[name]; reserved {
			continue
		}
		if _, exists := out[name]; exists {
			continue
		}
		out[name] = SanitizeMetadataValue(in[k])
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

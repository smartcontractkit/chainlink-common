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

// Limits on resource attributes accepted by SanitizeMetadataHeaders. They reserve headroom in the
// gRPC HEADERS frame for authentication and normal gRPC metadata, and bound how much of every Kafka
// record's header space a producer's resource attributes can consume.
const (
	maxResourceAttributes          = 32
	maxResourceAttributeKeyBytes   = 128
	maxResourceAttributeValueBytes = 512
	maxResourceAttributeTotalBytes = 4096 // sum of accepted key + value bytes, prefix excluded
)

// isPrintableASCII reports whether every byte of val is in the printable ASCII range [0x20, 0x7E].
// grpc-go hard-fails the entire RPC — auth header included — when an outgoing metadata value fails
// this check, so a value that does not pass is omitted rather than rewritten: a byte-mangled value
// is a worse outcome than a dropped attribute for an operator-facing routing/observability field.
func isPrintableASCII(val string) bool {
	for i := 0; i < len(val); i++ {
		if c := val[i]; c < 0x20 || c > 0x7E {
			return false
		}
	}
	return true
}

// DroppedAttribute records a resource attribute SanitizeMetadataHeaders omitted, and why.
type DroppedAttribute struct {
	Key    string
	Reason string
}

// Reasons a resource attribute can be omitted by SanitizeMetadataHeaders. Exposed as strings (not
// an enum type) so callers can attach them to a log field or a metric attribute directly.
const (
	reasonInvalidKey    = "invalid_key"
	reasonInvalidValue  = "invalid_value"
	reasonDuplicateKey  = "duplicate_key"
	reasonLimitExceeded = "limit_exceeded"
)

// sanitizeMetadataKey validates a resource-attribute key as a valid outgoing gRPC metadata key,
// without the ResourceHeaderPrefix that SanitizeMetadataHeaders adds, and reports whether it is
// valid. It never rewrites: a key that fails validation is omitted by the caller rather than
// mutated, so two distinct configured keys can never collapse into one gRPC metadata key.
//
// Valid keys, once lower-cased, match [0-9a-z-_.]+ (grpc's own key charset — see
// internal/metadata.ValidateKey) and do not end in "-bin", which grpc treats as declaring a
// base64-encoded binary value. A valid key's structure survives untouched: "csa_public_key" stays
// "csa_public_key" and "service.name" stays "service.name", which is what lets chip-ingress emit the
// forwarded header verbatim.
func sanitizeMetadataKey(key string) (string, bool) {
	if key == "" {
		return "", false
	}
	lower := strings.ToLower(key)
	for _, r := range lower {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_') {
			return "", false
		}
	}
	if strings.HasSuffix(lower, "-bin") {
		return "", false
	}
	return lower, true
}

// SanitizeMetadataHeaders validates a map of resource attributes for use as outgoing gRPC metadata
// (e.g. via NewStaticHeaderProvider). Every emitted key is ResourceHeaderPrefix followed by the
// validated key, unchanged, so service.name becomes resource_service.name and csa_public_key becomes
// resource_csa_public_key. Chip-ingress forwards keys carrying that prefix onto every Kafka record a
// request produces, emitting the key unchanged.
//
// The prefix is what makes this safe without a deny-list. The header interceptor appends to outgoing
// metadata rather than replacing it, so an attribute landing on an existing header name would send
// two values under one key — an attribute named X-Beholder-Node-Auth-Token would have broken
// authentication that way. Because every emitted key is prefixed, no attribute can reach a reserved
// gRPC key: that one becomes resource_x-beholder-node-auth-token, which collides with nothing, and
// the same holds for authorization, te, content-type, the grpc- prefix and pseudo-headers.
//
// An attribute is omitted, rather than rewritten, when: its key is empty, exceeds
// maxResourceAttributeKeyBytes, fails sanitizeMetadataKey's charset/[-bin] validation, or duplicates
// an already-accepted key (first in sorted order of the original keys wins); its value exceeds
// maxResourceAttributeValueBytes or is not printable ASCII (isPrintableASCII); or accepting it would
// push the accepted count past maxResourceAttributes or the accepted key+value byte total past
// maxResourceAttributeTotalBytes. Keys are processed in sorted order so every omission is
// deterministic. dropped records each omission and why, for the caller to warn and meter.
func SanitizeMetadataHeaders(in map[string]string) (map[string]string, []DroppedAttribute) {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys) // deterministic: first in sorted order wins, and excess entries drop from the tail

	out := make(map[string]string, len(in))
	var dropped []DroppedAttribute
	totalBytes := 0
	for _, k := range keys {
		if len(out) >= maxResourceAttributes {
			dropped = append(dropped, DroppedAttribute{Key: k, Reason: reasonLimitExceeded})
			continue
		}
		if len(k) > maxResourceAttributeKeyBytes {
			dropped = append(dropped, DroppedAttribute{Key: k, Reason: reasonInvalidKey})
			continue
		}
		name, ok := sanitizeMetadataKey(k)
		if !ok {
			dropped = append(dropped, DroppedAttribute{Key: k, Reason: reasonInvalidKey})
			continue
		}
		name = ResourceHeaderPrefix + name
		if _, dup := out[name]; dup {
			dropped = append(dropped, DroppedAttribute{Key: k, Reason: reasonDuplicateKey})
			continue
		}
		val := in[k]
		if len(val) > maxResourceAttributeValueBytes {
			dropped = append(dropped, DroppedAttribute{Key: k, Reason: reasonInvalidValue})
			continue
		}
		if !isPrintableASCII(val) {
			dropped = append(dropped, DroppedAttribute{Key: k, Reason: reasonInvalidValue})
			continue
		}
		if totalBytes+len(name)-len(ResourceHeaderPrefix)+len(val) > maxResourceAttributeTotalBytes {
			dropped = append(dropped, DroppedAttribute{Key: k, Reason: reasonLimitExceeded})
			continue
		}
		totalBytes += len(name) - len(ResourceHeaderPrefix) + len(val)
		out[name] = val
	}
	return out, dropped
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

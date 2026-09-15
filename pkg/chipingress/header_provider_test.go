package chipingress_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
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

func TestNewStaticHeaderProvider(t *testing.T) {
	headers := map[string]string{"chain_id": "1", "environment": "prod"}
	provider := chipingress.NewStaticHeaderProvider(headers)
	require.NotNil(t, provider)

	got, err := provider.Headers(t.Context())
	require.NoError(t, err)
	assert.Equal(t, headers, got)

	type tsr interface {
		RequireTransportSecurity() bool
	}
	tlsReq, ok := provider.(tsr)
	require.True(t, ok)
	assert.False(t, tlsReq.RequireTransportSecurity())
}

func TestSanitizeMetadataHeaders(t *testing.T) {
	t.Run("every whitelisted attribute maps to its fixed header name", func(t *testing.T) {
		in := map[string]string{
			"deployed_by":      "ci",
			"host.name":        "ip-10-0-0-1",
			"internal_node_id": "42",
			"node_id":          "7",
			"donID":            "don-1",
			"platformEnv":      "staging",
			"zone":             "us-east-1a",
			"csa_public_key":   "abc123",
			"service.name":     "chainlink",
			"service.sha":      "deadbeef",
		}
		got := chipingress.SanitizeMetadataHeaders(in)
		assert.Equal(t, map[string]string{
			"chainlink-resource-deployed-by":      "ci",
			"chainlink-resource-host-name":        "ip-10-0-0-1",
			"chainlink-resource-internal-node-id": "42",
			"chainlink-resource-node-id":          "7",
			"chainlink-resource-don-id":           "don-1",
			"chainlink-resource-platform-env":     "staging",
			"chainlink-resource-zone":             "us-east-1a",
			"chainlink-resource-csa-public-key":   "abc123",
			"chainlink-resource-service-name":     "chainlink",
			"chainlink-resource-service-sha":      "deadbeef",
		}, got)
		// The whitelist mapping itself is the wire contract with chip-ingress; pinning the input
		// spellings guards both sides drifting apart.
		assert.Len(t, chipingress.ResourceAttributeHeaders, len(in))
	})

	t.Run("key matching is case-insensitive", func(t *testing.T) {
		got := chipingress.SanitizeMetadataHeaders(map[string]string{
			"DonID":        "don-1",
			"PLATFORMENV":  "prod",
			"Service.Name": "chainlink",
		})
		assert.Equal(t, map[string]string{
			"chainlink-resource-don-id":       "don-1",
			"chainlink-resource-platform-env": "prod",
			"chainlink-resource-service-name": "chainlink",
		}, got)
	})

	t.Run("non-whitelisted keys are omitted", func(t *testing.T) {
		got := chipingress.SanitizeMetadataHeaders(map[string]string{
			"service.version": "1.2.3", // not whitelisted
			"k8s.pod-name_1":  "pod",   // valid gRPC key, still not whitelisted
			"":                "empty",
		})
		assert.Empty(t, got)
	})

	// This is the property that makes the whitelist a substitute for a deny-list. The header
	// interceptor appends to outgoing metadata rather than replacing, so an attribute landing on an
	// existing header name would send two values under one key — for the CSA auth token that breaks
	// authentication. Because the emitted header names are a fixed set under chainlink-, no
	// configured attribute can reach any reserved gRPC key.
	t.Run("no attribute can collide with a reserved gRPC metadata key", func(t *testing.T) {
		for _, key := range []string{
			"X-Beholder-Node-Auth-Token", // CSA auth
			"x-include-nop-info",         // WithNOPLookup
			"authorization",              // WithBasicAuth
			"te", "content-type", "cookie", "host", "user-agent",
			"grpc-timeout", "grpc-encoding",
		} {
			got := chipingress.SanitizeMetadataHeaders(map[string]string{key: "forged"})
			assert.Empty(t, got, "key %q must not be emittable as gRPC metadata", key)
		}
	})

	t.Run("non-printable values omit the whole attribute", func(t *testing.T) {
		got := chipingress.SanitizeMetadataHeaders(map[string]string{"csa_public_key": "abc\x01"})
		assert.Empty(t, got)
	})

	t.Run("case variants of one attribute resolve deterministically to sorted-first key", func(t *testing.T) {
		// Both map to chainlink-resource-don-id; sorted order of the ORIGINAL keys is "DonID" < "donid"
		// (upper-case sorts first in ASCII), so "DonID" wins.
		got := chipingress.SanitizeMetadataHeaders(map[string]string{"DonID": "upper", "donid": "lower"})
		assert.Equal(t, map[string]string{"chainlink-resource-don-id": "upper"}, got)
	})
}

// pingServer is a minimal ChipIngressServer that always answers Ping successfully.
type pingServer struct {
	pb.UnimplementedChipIngressServer
}

func (pingServer) Ping(context.Context, *pb.EmptyRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{}, nil
}

// TestSanitizeMetadataHeaders_AvoidsRPCFailure is a regression/guard test for the core reason
// SanitizeMetadataHeaders exists: grpc-go hard-fails an entire RPC (codes.Internal) when an
// outgoing metadata pair fails its charset validation. An unsanitized resource-attribute key or
// value (dots, non-printable characters) reproduces that failure; running it through
// SanitizeMetadataHeaders first must not.
func TestSanitizeMetadataHeaders_AvoidsRPCFailure(t *testing.T) {
	lis, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer lis.Close()

	srv := grpc.NewServer()
	pb.RegisterChipIngressServer(srv, pingServer{})
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	dirty := map[string]string{"csa_public_key": "abc\x01def"}

	t.Run("unsanitized headers fail the RPC", func(t *testing.T) {
		client, err := chipingress.NewClient(lis.Addr().String(),
			chipingress.WithInsecureConnection(),
			chipingress.WithHeaderProvider(chipingress.NewStaticHeaderProvider(dirty)),
		)
		require.NoError(t, err)
		defer client.Close() //nolint:errcheck

		_, err = client.Ping(t.Context(), &chipingress.EmptyRequest{})
		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("sanitized headers succeed", func(t *testing.T) {
		sanitized := chipingress.SanitizeMetadataHeaders(dirty)
		assert.Empty(t, sanitized, "the non-printable value must be omitted, not rewritten")
		client, err := chipingress.NewClient(lis.Addr().String(),
			chipingress.WithInsecureConnection(),
			chipingress.WithHeaderProvider(chipingress.NewStaticHeaderProvider(sanitized)),
		)
		require.NoError(t, err)
		defer client.Close() //nolint:errcheck

		_, err = client.Ping(t.Context(), &chipingress.EmptyRequest{})
		require.NoError(t, err)
	})
}

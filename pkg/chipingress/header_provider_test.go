package chipingress_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"net"
	"strings"
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

func TestSanitizeMetadataValue(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "printable ASCII is unchanged", in: "chain-1_prod.v2", want: "chain-1_prod.v2"},
		{name: "empty", in: "", want: ""},
		{name: "control character replaced", in: "value\nwith\tcontrol", want: "value?with?control"},
		{name: "non-ASCII UTF-8 replaced byte-wise", in: "café", want: "caf??"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, chipingress.SanitizeMetadataValue(tt.in))
		})
	}
}

const rp = chipingress.ResourceHeaderPrefix

func TestSanitizeMetadataHeaders(t *testing.T) {
	t.Run("keys are prefixed and keep their structure", func(t *testing.T) {
		got := chipingress.SanitizeMetadataHeaders(map[string]string{
			"service.name":   "beholder",
			"csa_public_key": "abc123",
			"node-operator":  "acme",
			"DonID":          "don-1",
		})
		assert.Equal(t, map[string]string{
			rp + "service.name":   "beholder",
			rp + "csa_public_key": "abc123",
			rp + "node-operator":  "acme",
			rp + "donid":          "don-1",
		}, got)
	})

	t.Run("structure-preserving normalization", func(t *testing.T) {
		tests := []struct {
			name string
			in   string
			want string
		}{
			// grpc accepts [0-9a-z-_.], so structure survives and chip-ingress can emit the
			// forwarded header verbatim.
			{"snake case preserved", "csa_public_key", rp + "csa_public_key"},
			{"dotted preserved", "service.name", rp + "service.name"},
			{"upper-cased is lowered", "DonID", rp + "donid"},
			{"mixed separators preserved", "k8s.pod-name_1", rp + "k8s.pod-name_1"},
			{"illegal characters become underscores", "chain id/2:x", rp + "chain_id_2_x"},
			{"non-ascii becomes underscores", "héllo", rp + "h_llo"},
			// A "-bin" suffix tells grpc the value is base64-encoded binary; rewrite it so grpc
			// does not try to decode a plain-text resource attribute.
			{"bin suffix is rewritten", "payload-bin", rp + "payload_bin"},
			{"bin substring is untouched", "payload-binary", rp + "payload-binary"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, map[string]string{tt.want: "v"},
					chipingress.SanitizeMetadataHeaders(map[string]string{tt.in: "v"}))
			})
		}
	})

	t.Run("keys with nothing left after normalization are dropped", func(t *testing.T) {
		// A bare prefix carries no information.
		for _, key := range []string{"", "---", "__"} {
			assert.Empty(t, chipingress.SanitizeMetadataHeaders(map[string]string{key: "value"}),
				"key %q must be dropped", key)
		}
	})

	// This is the property that replaces the reserved-key set the prefix made redundant. The header
	// interceptor appends to outgoing metadata rather than replacing, so an attribute landing on an
	// existing header name would send two values under one key — for the CSA auth token that breaks
	// authentication. Prefixing puts every attribute out of reach of every reserved gRPC key.
	t.Run("no attribute can collide with a reserved gRPC metadata key", func(t *testing.T) {
		for _, key := range []string{
			"X-Beholder-Node-Auth-Token", // CSA auth token, via WithTokenAuth
			"x-include-nop-info",         // WithNOPLookup
			"authorization",              // WithBasicAuth
			"te", "content-type", "cookie", "host", "user-agent",
			"grpc-timeout", "grpc-encoding",
		} {
			got := chipingress.SanitizeMetadataHeaders(map[string]string{key: "forged"})
			require.Len(t, got, 1, "key %q should still be sent, just namespaced", key)
			for name := range got {
				assert.True(t, strings.HasPrefix(name, rp), "key %q must be prefixed, got %q", key, name)
				assert.NotEqual(t, strings.ToLower(key), name, "key %q must not reach the reserved name", key)
			}
		}
	})

	t.Run("CloudEvents context attribute names are kept, they mean nothing as gRPC metadata", func(t *testing.T) {
		got := chipingress.SanitizeMetadataHeaders(map[string]string{"subject": "keep-me", "source": "keep-me-too"})
		assert.Equal(t, map[string]string{rp + "subject": "keep-me", rp + "source": "keep-me-too"}, got)
	})

	t.Run("non-printable values are sanitized", func(t *testing.T) {
		got := chipingress.SanitizeMetadataHeaders(map[string]string{"chain_id": "1\n2"})
		assert.Equal(t, "1?2", got[rp+"chain_id"])
	})

	t.Run("duplicate normalized keys resolve deterministically to sorted-first key", func(t *testing.T) {
		// Both normalize to chain_id; sorted order is "chain id" < "chain_id" (' ' < '_'), so the
		// space-separated key wins.
		got := chipingress.SanitizeMetadataHeaders(map[string]string{"chain id": "from-space", "chain_id": "from-snake"})
		assert.Equal(t, map[string]string{rp + "chain_id": "from-space"}, got)
	})

	t.Run("keys that differ only in case collapse deterministically", func(t *testing.T) {
		got := chipingress.SanitizeMetadataHeaders(map[string]string{"DonID": "upper", "donid": "lower"})
		// sorted order: "DonID" < "donid" (upper-case sorts first in ASCII).
		assert.Equal(t, map[string]string{rp + "donid": "upper"}, got)
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

	dirty := map[string]string{"k8s.pod.name": "pod-\x01abc"}

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
		client, err := chipingress.NewClient(lis.Addr().String(),
			chipingress.WithInsecureConnection(),
			chipingress.WithHeaderProvider(chipingress.NewStaticHeaderProvider(chipingress.SanitizeMetadataHeaders(dirty))),
		)
		require.NoError(t, err)
		defer client.Close() //nolint:errcheck

		_, err = client.Ping(t.Context(), &chipingress.EmptyRequest{})
		require.NoError(t, err)
	})
}

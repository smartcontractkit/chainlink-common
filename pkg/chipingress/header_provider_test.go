package chipingress_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
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

const rp = chipingress.ResourceHeaderPrefix

func TestSanitizeMetadataHeaders(t *testing.T) {
	t.Run("valid keys are prefixed and kept verbatim", func(t *testing.T) {
		got, dropped := chipingress.SanitizeMetadataHeaders(map[string]string{
			"service.name":   "beholder",
			"csa_public_key": "abc123",
			"node-operator":  "acme",
			"donid":          "don-1",
		})
		assert.Empty(t, dropped)
		assert.Equal(t, map[string]string{
			rp + "service.name":   "beholder",
			rp + "csa_public_key": "abc123",
			rp + "node-operator":  "acme",
			rp + "donid":          "don-1",
		}, got)
	})

	t.Run("validate, don't rewrite", func(t *testing.T) {
		tests := []struct {
			name    string
			in      string
			want    string // "" if the key must be omitted
			omitted bool
		}{
			// grpc accepts [0-9a-z-_.], so a valid key's structure survives untouched and
			// chip-ingress can emit the forwarded header verbatim.
			{name: "snake case preserved", in: "csa_public_key", want: rp + "csa_public_key"},
			{name: "dotted preserved", in: "service.name", want: rp + "service.name"},
			{name: "upper-cased is lowered", in: "DonID", want: rp + "donid"},
			{name: "mixed separators preserved", in: "k8s.pod-name_1", want: rp + "k8s.pod-name_1"},
			// Invalid keys are OMITTED, never rewritten: silently collapsing two distinct
			// configured keys into one gRPC metadata key is worse than dropping one.
			{name: "illegal characters omit the attribute", in: "chain id/2:x", omitted: true},
			{name: "non-ascii omits the attribute", in: "héllo", omitted: true},
			{name: "empty key omits the attribute", in: "", omitted: true},
			// A "-bin" suffix tells grpc the value is base64-encoded binary; omit rather than
			// rewrite so a plain-text resource attribute can't silently start being decoded.
			{name: "bin suffix omits the attribute", in: "payload-bin", omitted: true},
			{name: "bin substring is untouched", in: "payload-binary", want: rp + "payload-binary"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, dropped := chipingress.SanitizeMetadataHeaders(map[string]string{tt.in: "v"})
				if tt.omitted {
					assert.Empty(t, got)
					require.Len(t, dropped, 1)
					assert.Equal(t, tt.in, dropped[0].Key)
					return
				}
				assert.Empty(t, dropped)
				assert.Equal(t, map[string]string{tt.want: "v"}, got)
			})
		}
	})

	// This is the property that replaces the reserved-key set the prefix made redundant. The header
	// interceptor appends to outgoing metadata rather than replacing, so an attribute landing on an
	// existing header name would send two values under one key — for the CSA auth token that breaks
	// authentication. Prefixing puts every valid attribute out of reach of every reserved gRPC key.
	t.Run("no attribute can collide with a reserved gRPC metadata key", func(t *testing.T) {
		for _, key := range []string{
			"x-include-nop-info", // WithNOPLookup
			"authorization",      // WithBasicAuth
			"te", "content-type", "cookie", "host", "user-agent",
			"grpc-timeout", "grpc-encoding",
		} {
			got, dropped := chipingress.SanitizeMetadataHeaders(map[string]string{key: "forged"})
			assert.Empty(t, dropped, "key %q should be valid and namespaced, not dropped", key)
			require.Len(t, got, 1, "key %q should still be sent, just namespaced", key)
			for name := range got {
				assert.True(t, strings.HasPrefix(name, rp), "key %q must be prefixed, got %q", key, name)
				assert.NotEqual(t, strings.ToLower(key), name, "key %q must not reach the reserved name", key)
			}
		}
		// The CSA auth token header contains uppercase letters and hyphens; hyphens are a valid
		// gRPC metadata char, so this key is namespaced rather than omitted, same as the others.
		got, dropped := chipingress.SanitizeMetadataHeaders(map[string]string{"X-Beholder-Node-Auth-Token": "forged"})
		assert.Empty(t, dropped)
		assert.Equal(t, map[string]string{rp + "x-beholder-node-auth-token": "forged"}, got)
	})

	t.Run("CloudEvents context attribute names are kept, they mean nothing as gRPC metadata", func(t *testing.T) {
		got, dropped := chipingress.SanitizeMetadataHeaders(map[string]string{"subject": "keep-me", "source": "keep-me-too"})
		assert.Empty(t, dropped)
		assert.Equal(t, map[string]string{rp + "subject": "keep-me", rp + "source": "keep-me-too"}, got)
	})

	t.Run("non-printable values omit the whole attribute", func(t *testing.T) {
		got, dropped := chipingress.SanitizeMetadataHeaders(map[string]string{"chain_id": "1\n2"})
		assert.Empty(t, got)
		require.Len(t, dropped, 1)
		assert.Equal(t, "chain_id", dropped[0].Key)
		assert.Equal(t, "invalid_value", dropped[0].Reason)
	})

	t.Run("duplicate keys resolve deterministically to sorted-first key", func(t *testing.T) {
		// "DonID" and "donid" both validate to "donid"; sorted order of the ORIGINAL keys is
		// "DonID" < "donid" (upper-case sorts first in ASCII), so "DonID" wins.
		got, dropped := chipingress.SanitizeMetadataHeaders(map[string]string{"DonID": "upper", "donid": "lower"})
		assert.Equal(t, map[string]string{rp + "donid": "upper"}, got)
		require.Len(t, dropped, 1)
		assert.Equal(t, "donid", dropped[0].Key)
		assert.Equal(t, "duplicate_key", dropped[0].Reason)
	})

	t.Run("oversized key is omitted, not truncated", func(t *testing.T) {
		longKey := strings.Repeat("a", 129)
		got, dropped := chipingress.SanitizeMetadataHeaders(map[string]string{longKey: "v"})
		assert.Empty(t, got)
		require.Len(t, dropped, 1)
		assert.Equal(t, "invalid_key", dropped[0].Reason)
	})

	t.Run("oversized value is omitted, not truncated", func(t *testing.T) {
		longVal := strings.Repeat("v", 513)
		got, dropped := chipingress.SanitizeMetadataHeaders(map[string]string{"chain_id": longVal})
		assert.Empty(t, got)
		require.Len(t, dropped, 1)
		assert.Equal(t, "invalid_value", dropped[0].Reason)
	})

	t.Run("attribute count is capped at 32, excess dropped deterministically", func(t *testing.T) {
		in := make(map[string]string, 33)
		for i := 0; i < 33; i++ {
			in[fmt.Sprintf("attr_%02d", i)] = "v"
		}
		got, dropped := chipingress.SanitizeMetadataHeaders(in)
		assert.Len(t, got, 32)
		require.Len(t, dropped, 1)
		// Sorted order: "attr_32" sorts last among "attr_00".."attr_32".
		assert.Equal(t, "attr_32", dropped[0].Key)
		assert.Equal(t, "limit_exceeded", dropped[0].Reason)
	})

	t.Run("total key+value bytes are capped at 4096, tail dropped deterministically", func(t *testing.T) {
		// Each accepted attribute contributes len(key)+len(value) bytes (prefix excluded). 9
		// attributes of 500 bytes each would total 4500, over the 4096 cap, so the last one or
		// two (in sorted order) must be dropped.
		in := make(map[string]string, 9)
		val := strings.Repeat("v", 490)
		for i := 0; i < 9; i++ {
			in[fmt.Sprintf("attr_%d", i)] = val // key is 6 bytes, so each entry is 496 bytes
		}
		got, dropped := chipingress.SanitizeMetadataHeaders(in)
		assert.NotEmpty(t, dropped)
		for _, d := range dropped {
			assert.Equal(t, "limit_exceeded", d.Reason)
		}
		total := 0
		for name, v := range got {
			total += len(name) - len(rp) + len(v)
		}
		assert.LessOrEqual(t, total, 4096)
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
		sanitized, dropped := chipingress.SanitizeMetadataHeaders(dirty)
		require.Len(t, dropped, 1, "the non-printable value must be omitted, not rewritten")
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

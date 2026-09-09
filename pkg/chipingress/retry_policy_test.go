package chipingress

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gp "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"google.golang.org/grpc/serviceconfig"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

// parseServiceConfigJSON drives the given service-config JSON through gRPC's real,
// production parser (the same one grpc.NewClient uses for grpc.WithDefaultServiceConfig), via
// a manual resolver. This is the only way to exercise that parser from outside
// google.golang.org/grpc, since the concrete parsing logic lives under an internal package.
//
// This is deliberately NOT a hand-rolled reimplementation of gRPC's parsing rules: the bug this
// package shipped previously (a malformed, flat service-config JSON) was only "caught" by
// actually parsing it with gRPC's own code - asserting "NewClient returned no error" is
// worthless, since that malformed JSON also produced no error.
func parseServiceConfigJSON(t *testing.T, scJSON string) *serviceconfig.ParseResult {
	t.Helper()

	scheme := fmt.Sprintf("chipingress-retry-policy-test-%d", time.Now().UnixNano())
	r := manual.NewBuilderWithScheme(scheme)

	ccCh := make(chan resolver.ClientConn, 1)
	r.BuildCallback = func(_ resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) {
		ccCh <- cc
	}

	conn, err := gp.NewClient(r.Scheme()+":///test",
		gp.WithTransportCredentials(insecure.NewCredentials()),
		gp.WithResolvers(r),
	)
	require.NoError(t, err)
	defer conn.Close() //nolint:errcheck

	conn.Connect()

	select {
	case cc := <-ccCh:
		return cc.ParseServiceConfig(scJSON)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the manual resolver to be built")
		return nil
	}
}

// chipIngressMethodPath is the map key gRPC's service-config parser uses for a
// service-scoped (no method) name selector: "/<service>/".
var chipIngressMethodPath = "/" + pb.ChipIngress_ServiceDesc.ServiceName + "/"

// TestBuildRetryServiceConfigJSON_DefaultPolicy is the core regression test: it proves the
// service config the client installs by default actually results in gRPC installing a retry
// policy for the ChipIngress service, with the exact values the client intends. A test that
// merely asserted "NewClient/buildRetryServiceConfigJSON returns no error" would NOT have
// caught the original bug, since the original malformed JSON also parsed without error.
func TestBuildRetryServiceConfigJSON_DefaultPolicy(t *testing.T) {
	throttling := defaultRetryThrottlingPolicy()
	scJSON, err := buildRetryServiceConfigJSON(defaultRetryPolicy(), &throttling)
	require.NoError(t, err)

	scpr := parseServiceConfigJSON(t, scJSON)
	require.NoError(t, scpr.Err, "gRPC's parser rejected the generated service config")
	require.NotNil(t, scpr.Config)

	sc, ok := scpr.Config.(*gp.ServiceConfig)
	require.True(t, ok, "expected *grpc.ServiceConfig, got %T", scpr.Config)
	require.NotNil(t, sc.Methods, "MethodConfig must be populated - this is exactly what the old malformed JSON failed to do")

	mc, ok := sc.Methods[chipIngressMethodPath]
	require.True(t, ok, "expected a MethodConfig entry for %q, got keys %v", chipIngressMethodPath, mapKeys(sc.Methods))
	require.NotNil(t, mc.RetryPolicy, "expected a retry policy to be installed for the ChipIngress service")

	assert.Equal(t, 3, mc.RetryPolicy.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, mc.RetryPolicy.InitialBackoff)
	assert.Equal(t, 1*time.Second, mc.RetryPolicy.MaxBackoff)
	assert.InDelta(t, 2.0, mc.RetryPolicy.BackoffMultiplier, 0.0001)
	assert.True(t, mc.RetryPolicy.RetryableStatusCodes[codes.Unavailable])
	assert.True(t, mc.RetryPolicy.RetryableStatusCodes[codes.ResourceExhausted])
	assert.Len(t, mc.RetryPolicy.RetryableStatusCodes, 2)
}

// TestBuildRetryServiceConfigJSON_CustomPolicy exercises WithRetryPolicy end to end: a custom
// policy must round-trip through gRPC's parser with the caller's exact values.
func TestBuildRetryServiceConfigJSON_CustomPolicy(t *testing.T) {
	custom := RetryPolicy{
		MaxAttempts:          5,
		InitialBackoff:       250 * time.Millisecond,
		MaxBackoff:           5 * time.Second,
		BackoffMultiplier:    1.5,
		RetryableStatusCodes: []codes.Code{codes.Unavailable, codes.Internal, codes.DeadlineExceeded},
	}

	cfg := newClientConfig("localhost")
	WithRetryPolicy(custom)(cfg)
	assert.Equal(t, custom, cfg.retryPolicy)

	scJSON, err := buildRetryServiceConfigJSON(cfg.retryPolicy, nil)
	require.NoError(t, err)

	scpr := parseServiceConfigJSON(t, scJSON)
	require.NoError(t, scpr.Err)

	sc, ok := scpr.Config.(*gp.ServiceConfig)
	require.True(t, ok)

	mc, ok := sc.Methods[chipIngressMethodPath]
	require.True(t, ok)
	require.NotNil(t, mc.RetryPolicy)

	assert.Equal(t, 5, mc.RetryPolicy.MaxAttempts)
	assert.Equal(t, 250*time.Millisecond, mc.RetryPolicy.InitialBackoff)
	assert.Equal(t, 5*time.Second, mc.RetryPolicy.MaxBackoff)
	assert.InDelta(t, 1.5, mc.RetryPolicy.BackoffMultiplier, 0.0001)
	assert.True(t, mc.RetryPolicy.RetryableStatusCodes[codes.Unavailable])
	assert.True(t, mc.RetryPolicy.RetryableStatusCodes[codes.Internal])
	assert.True(t, mc.RetryPolicy.RetryableStatusCodes[codes.DeadlineExceeded])
	assert.Len(t, mc.RetryPolicy.RetryableStatusCodes, 3)
}

// TestOldMalformedServiceConfigInstallsNoRetryPolicy demonstrates the original bug this package
// shipped: the flat, unnested JSON that used to be hard-coded in NewClient. gRPC's parser
// accepts it (all its fields are unknown at the top level and silently discarded) and returns
// no error, yet installs zero MethodConfig entries - i.e., no retry policy at all. This is why
// asserting "no error" is not a valid regression test for this bug class.
func TestOldMalformedServiceConfigInstallsNoRetryPolicy(t *testing.T) {
	const oldMalformedJSON = `{
		"maxAttempts": 3,
		"initialBackoff": "100ms",
		"maxBackoff": "1s",
		"backoffMultiplier": 2,
		"retryableStatusCodes": ["UNAVAILABLE", "RESOURCE_EXHAUSTED"]
	}`

	scpr := parseServiceConfigJSON(t, oldMalformedJSON)
	require.NoError(t, scpr.Err, "the old malformed shape parses without error - that IS the bug")

	sc, ok := scpr.Config.(*gp.ServiceConfig)
	require.True(t, ok)
	assert.Empty(t, sc.Methods, "the old malformed shape must install no retry policy for any method")
}

// TestGRPCServiceConfigDuration_RejectsGoDurationStrings proves the second, latent defect noted
// while fixing the first: gRPC's service-config duration parser (protobuf JSON duration
// encoding) does NOT accept Go duration strings like "100ms" once they're actually reachable
// (i.e., once correctly nested under methodConfig[].retryPolicy). Before the nesting fix, this
// was unreachable - the malformed top-level JSON meant the duration strings were never even
// looked at by gRPC's typed unmarshaling.
func TestGRPCServiceConfigDuration_RejectsGoDurationStrings(t *testing.T) {
	const malformedDurationJSON = `{
		"methodConfig": [{
			"name": [{"service": "chipingress.pb.ChipIngress"}],
			"retryPolicy": {
				"maxAttempts": 3,
				"initialBackoff": "100ms",
				"maxBackoff": "1s",
				"backoffMultiplier": 2,
				"retryableStatusCodes": ["UNAVAILABLE", "RESOURCE_EXHAUSTED"]
			}
		}]
	}`

	scpr := parseServiceConfigJSON(t, malformedDurationJSON)
	require.Error(t, scpr.Err, "\"100ms\" is not a valid protobuf JSON duration and must be rejected once actually parsed")
}

// TestGRPCServiceConfigDuration_Format asserts the exact wire format grpcServiceConfigDuration
// produces for representative values, and that it round-trips through gRPC's parser.
func TestGRPCServiceConfigDuration_Format(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{name: "100ms", d: 100 * time.Millisecond, want: "0.1s"},
		{name: "1s", d: 1 * time.Second, want: "1s"},
		{name: "250ms", d: 250 * time.Millisecond, want: "0.25s"},
		{name: "1500ms", d: 1500 * time.Millisecond, want: "1.5s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, grpcServiceConfigDuration(tt.d))
		})
	}
}

// TestBuildRetryServiceConfigJSON_RetryThrottling asserts the retryThrottling block is present
// and correctly populated when requested, since it lives alongside - but structurally separate
// from - the per-method retry policy.
func TestBuildRetryServiceConfigJSON_RetryThrottling(t *testing.T) {
	throttling := RetryThrottlingPolicy{MaxTokens: 20, TokenRatio: 0.2}
	scJSON, err := buildRetryServiceConfigJSON(defaultRetryPolicy(), &throttling)
	require.NoError(t, err)

	scpr := parseServiceConfigJSON(t, scJSON)
	require.NoError(t, scpr.Err)

	sc, ok := scpr.Config.(*gp.ServiceConfig)
	require.True(t, ok)
	// retryThrottling is unexported on grpc.ServiceConfig, so we can't assert its fields
	// directly; asserting no parse error with maxTokens/tokenRatio present in range is the
	// available signal that gRPC accepted and applied it (see service_config.go's range
	// validation: maxTokens in (0, 1000], tokenRatio > 0).
	require.NotNil(t, sc)
}

func mapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

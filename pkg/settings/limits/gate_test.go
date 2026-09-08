package limits

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func ExampleGateLimiter_AllowErr() {
	ctx := context.Background()
	gl := NewGateLimiter(true)

	open, err := gl.Limit(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("open:", open)

	err = gl.AllowErr(ctx)
	fmt.Println("allow:", err)

	gl = NewGateLimiter(false)

	open, err = gl.Limit(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("open:", open)

	err = gl.AllowErr(ctx)
	fmt.Println("allow:", err)

	// Output:
	// open: true
	// allow: <nil>
	// open: false
	// allow: limited: not allowed
}

// TestGateLimiter_Open covers the property fail-closed callers depend on: a closed gate
// must not look like an evaluation failure, and vice versa.
func TestGateLimiter_Open(t *testing.T) {
	t.Parallel()

	t.Run("open gate", func(t *testing.T) {
		t.Parallel()
		open, err := NewGateLimiter(true).Open(t.Context())
		require.NoError(t, err)
		assert.True(t, open)
	})

	t.Run("closed gate is not an error", func(t *testing.T) {
		t.Parallel()
		open, err := NewGateLimiter(false).Open(t.Context())
		require.NoError(t, err, "a closed gate is a normal outcome, not a failure")
		assert.False(t, open)
	})

	t.Run("read failure is distinguishable from closed", func(t *testing.T) {
		t.Parallel()
		setting := settings.Bool(true)
		setting.Key, setting.Scope = "test.gate.open", settings.ScopeGlobal
		gl, err := MakeGateLimiter(Factory{Settings: failingGetter{}}, setting)
		require.NoError(t, err)
		t.Cleanup(func() { assert.NoError(t, gl.Close()) })

		open, err := gl.Open(t.Context())
		require.ErrorIs(t, err, errGetterUnavailable, "an unevaluatable gate must surface the error")
		assert.False(t, open)
	})
}

// TestGateLimiter_NoTenantFailsOpen pins the missing-tenant behaviour for a scope that
// does not require one (ScopeOrg). get() returns before resolving a value, so AllowErr
// used to read the unset `open` as a denial and report ErrorNotAllowed, contradicting its
// own "failing open" log. It now fails open, matching boundLimiter/rangeLimiter Check.
func TestGateLimiter_NoTenantFailsOpen(t *testing.T) {
	t.Parallel()

	setting := settings.Bool(true)
	setting.Key, setting.Scope = "test.gate.no-tenant", settings.ScopeOrg
	gl, err := MakeGateLimiter(Factory{Logger: logger.Test(t)}, setting)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, gl.Close()) })

	ctx := t.Context() // no contexts.WithCRE, so no org tenant

	require.NoError(t, gl.AllowErr(ctx), "an org gate with no org in context must not deny")

	open, err := gl.Open(ctx)
	require.NoError(t, err)
	assert.True(t, open, "Open must not report a closed gate when the gate was never evaluated")
}

func TestMakeGateLimiter(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		scope settings.Scope
		cre   contexts.CRE
	}{
		{settings.ScopeGlobal, contexts.CRE{}},
		{settings.ScopeOwner, contexts.CRE{Owner: "ow-id"}},
	} {
		t.Run(tt.scope.String(), func(t *testing.T) {
			t.Parallel()
			mc := newMetricsChecker(t)
			f := Factory{Meter: mc.Meter(t.Name())}
			limit := settings.PerChainSelector(settings.Bool(false),
				map[string]bool{
					"42": true,
				})

			limit.Default.Key = "foo.bar"
			limit.Default.Scope = tt.scope
			gl, err := MakeGateLimiter(f, limit)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, gl.Close()) })

			ctx := t.Context()
			ctx = contexts.WithCRE(ctx, tt.cre)

			assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 42)))
			var errGate ErrorNotAllowed
			if assert.ErrorAs(t, gl.AllowErr(contexts.WithChainSelector(ctx, 100)), &errGate) {
				assert.Equal(t, "foo.bar", errGate.Key)
				assert.Equal(t, tt.scope, errGate.Scope)
			}

			ms := mc.lastResourceFirstScopeMetric(t)

			attrs := attribute.NewSet(kvsFromScope(ctx, tt.scope)...)

			require.Equal(t, metrics{
				{
					Name: "gate.foo.bar.limit",
					Data: metricdata.Gauge[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: int64(0)},
						},
					},
				},
				{
					Name: "gate.foo.bar.usage",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attrs,
								Value:      int64(1),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
				{
					Name: "gate.foo.bar.denied",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attrs,
								Value:      int64(1),
							},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
			}, ms)
		})
	}
}

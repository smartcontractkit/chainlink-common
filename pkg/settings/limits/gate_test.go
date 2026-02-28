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
	// allow: limited: operation not allowed. This action is restricted by current configuration and gate settings
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

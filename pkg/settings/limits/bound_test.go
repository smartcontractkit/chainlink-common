package limits

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func ExampleBoundLimiter_Check_upper() {
	bl := NewUpperBoundLimiter(10)
	fn := func(n int) {
		if err := bl.Check(context.Background(), n); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("used", n)
	}
	fn(11)
	fn(4)
	fn(10)
	// Output:
	// limited: cannot use 11, limit is 10
	// used 4
	// used 10
}

func ExampleBoundLimiter_Check_lower() {
	bl := NewLowerBoundLimiter(10)
	fn := func(n int) {
		if err := bl.Check(context.Background(), n); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("used", n)
	}
	fn(4)
	fn(11)
	fn(10)
	// Output:
	// limited: cannot use 4, limit is 10
	// used 11
	// used 10
}

func TestMakeBoundLimiter(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		scope settings.Scope
		cre   contexts.CRE
	}{
		{settings.ScopeGlobal, contexts.CRE{}},
		{settings.ScopeOwner, contexts.CRE{Owner: "ow-id"}},
	} {
		for _, dir := range []struct {
			name    string
			makeFn  func(Factory, settings.IsSetting[config.Size]) (BoundLimiter[config.Size], error)
			denied  config.Size
			allowed []config.Size
		}{
			{
				name:    "upper",
				makeFn:  MakeUpperBoundLimiter[config.Size],
				denied:  11 * config.GByte,
				allowed: []config.Size{4 * config.GByte, 10 * config.GByte},
			},
			{
				name:    "lower",
				makeFn:  MakeLowerBoundLimiter[config.Size],
				denied:  4 * config.GByte,
				allowed: []config.Size{11 * config.GByte, 10 * config.GByte},
			},
		} {
			t.Run(tt.scope.String()+"/"+dir.name, func(t *testing.T) {
				t.Parallel()
				mc := newMetricsChecker(t)
				f := Factory{Meter: mc.Meter(t.Name())}
				limit := settings.Size(10 * config.GByte)
				limit.Key = "foo.bar"
				limit.Scope = tt.scope
				bl, err := dir.makeFn(f, limit)
				require.NoError(t, err)
				t.Cleanup(func() { assert.NoError(t, bl.Close()) })

				ctx := t.Context()
				ctx = contexts.WithCRE(ctx, tt.cre)

				var errBound ErrorBoundLimited[config.Size]
				if assert.ErrorAs(t, bl.Check(ctx, dir.denied), &errBound) {
					assert.Equal(t, "foo.bar", errBound.Key)
					assert.Equal(t, tt.scope, errBound.Scope)
					assert.Equal(t, 10*config.GByte, errBound.Limit)
					assert.Equal(t, dir.denied, errBound.Amount)
				}
				for _, v := range dir.allowed {
					assert.NoError(t, bl.Check(ctx, v))
				}

				ms := mc.lastResourceFirstScopeMetric(t)
				redactHistogramVals[int64](t, ms, "bound.foo.bar.usage")
				redactHistogramVals[int64](t, ms, "bound.foo.bar.denied")

				attrs := attribute.NewSet(kvsFromScope(ctx, tt.scope)...)

				require.Equal(t, metrics{
					{
						Name: "bound.foo.bar.limit",
						Unit: "By",
						Data: metricdata.Gauge[int64]{
							DataPoints: []metricdata.DataPoint[int64]{
								{Attributes: attrs, Value: int64(10 * config.GByte)},
							},
						},
					},
					{
						Name: "bound.foo.bar.usage",
						Unit: "By",
						Data: metricdata.Histogram[int64]{
							DataPoints: []metricdata.HistogramDataPoint[int64]{
								{
									Attributes:   attrs,
									Count:        2,
									Bounds:       []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
									BucketCounts: []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
								},
							},
							Temporality: metricdata.CumulativeTemporality,
						},
					},
					{
						Name: "bound.foo.bar.denied",
						Unit: "By",
						Data: metricdata.Histogram[int64]{
							DataPoints: []metricdata.HistogramDataPoint[int64]{
								{
									Attributes:   attrs,
									Count:        1,
									Bounds:       []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
									BucketCounts: []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
								},
							},
							Temporality: metricdata.CumulativeTemporality,
						},
					},
				}, ms)
			})
		}
	}
}

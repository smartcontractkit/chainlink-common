package limits

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func ExampleTimeLimiter_WithTimeout() {
	tl := NewTimeLimiter(time.Second)
	fn := func(ctx context.Context) {
		if ctx.Err() != nil {
			fmt.Println(ctx.Err())
			return
		}
		fmt.Println("done")
	}
	ctx, cancel, err := tl.WithTimeout(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cancel()
	fn(ctx)
	cancel()
	fn(ctx)
	// Output:
	// done
	// context canceled
}

func TestFactory_NewTimeLimiter(t *testing.T) {
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
			s := settings.Duration(time.Second)
			s.Key = "foo.bar"
			s.Scope = tt.scope
			s.Unit = "{action}"
			tl, err := f.MakeTimeLimiter(s)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, tl.Close()) })

			ctx := t.Context()
			ctx = contexts.WithCRE(ctx, tt.cre)

			func(ctx context.Context) {
				ctx, done, err := tl.WithTimeout(ctx)
				require.NoError(t, err)
				defer done()
				time.Sleep(10 * time.Millisecond)
			}(ctx)
			func(ctx context.Context) {
				ctx, done, err := tl.WithTimeout(ctx)
				require.NoError(t, err)
				defer done()
				time.Sleep(2 * time.Second)
			}(ctx)

			ms := mc.lastResourceFirstScopeMetric(t)
			redactHistogramVals[float64](t, ms, "time.foo.bar.runtime")

			attrs := attribute.NewSet(kvsFromScope(ctx, tt.scope)...)

			require.Equal(t, metrics{
				{
					Name: "time.foo.bar.limit",
					Unit: "s",
					Data: metricdata.Gauge[float64]{
						DataPoints: []metricdata.DataPoint[float64]{
							{Attributes: attrs, Value: 1},
						},
					},
				},
				{
					Name: "time.foo.bar.runtime",
					Unit: "s",
					Data: metricdata.Histogram[float64]{
						DataPoints: []metricdata.HistogramDataPoint[float64]{
							{
								Attributes:   attrs,
								Count:        2,
								Bounds:       []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
								BucketCounts: []uint64{0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
				{
					Name: "time.foo.bar.timeout",
					Unit: "{action}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: 1},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
				{
					Name: "time.foo.bar.success",
					Unit: "{action}",
					Data: metricdata.Sum[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: 1},
						},
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
					},
				},
			}, ms)
		})
	}
}

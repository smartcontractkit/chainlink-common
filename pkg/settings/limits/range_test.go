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

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func ExampleRangeLimiter_Check() {
	bl := NewRangeLimiter(settings.Range[int]{Lower: 1, Upper: 10})
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
	// limited: cannot use 11, limited to range [1,10]
	// used 4
	// used 10
}

func TestMakeRangeLimiter(t *testing.T) {
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
			startTime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
			limit := settings.TimeRange(startTime, startTime.Add(10*time.Hour))
			limit.Key = "foo.bar"
			limit.Scope = tt.scope
			bl, err := MakeRangeLimiter[config.Timestamp](f, limit)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, bl.Close()) })

			ctx := t.Context()
			ctx = contexts.WithCRE(ctx, tt.cre)

			denied := config.Timestamp(startTime.Add(11 * time.Hour).Unix())
			var errBound ErrorRangeLimited[config.Timestamp]
			if assert.ErrorAs(t, bl.Check(ctx, denied), &errBound) {
				assert.Equal(t, "foo.bar", errBound.Key)
				assert.Equal(t, tt.scope, errBound.Scope)
				assert.Equal(t, limit.DefaultValue, errBound.Limit)
				assert.Equal(t, denied, errBound.Amount)
			}
			for _, v := range []time.Time{startTime, startTime.Add(5 * time.Hour)} {
				assert.NoError(t, bl.Check(ctx, config.Timestamp(v.Unix())))
			}

			ms := mc.lastResourceFirstScopeMetric(t)
			redactHistogramVals[int64](t, ms, "range.foo.bar.usage")
			redactHistogramVals[int64](t, ms, "range.foo.bar.denied")

			attrs := attribute.NewSet(kvsFromScope(ctx, tt.scope)...)

			require.Equal(t, metrics{
				{
					Name: "range.foo.bar.lower.limit",
					Unit: "s",
					Data: metricdata.Gauge[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: int64(limit.DefaultValue.Lower)},
						},
					},
				},
				{
					Name: "range.foo.bar.upper.limit",
					Unit: "s",
					Data: metricdata.Gauge[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: int64(limit.DefaultValue.Upper)},
						},
					},
				},
				{
					Name: "range.foo.bar.usage",
					Unit: "s",
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
					Name: "range.foo.bar.denied",
					Unit: "s",
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

package limits

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func ExampleQueueLimiter() {
	ctx := context.Background()
	ql := NewQueueLimiter[string](2)

	if err := ql.Put(ctx, "foo"); err != nil {
		log.Fatalf("Failed to put foo: %v", err)
	}
	fmt.Println("Queued foo")
	// [foo]
	if err := ql.Put(ctx, "bar"); err != nil {
		log.Fatalf("Failed to put bar: %v", err)
	}
	fmt.Println("Queued bar")
	// [foo, bar]
	if err := ql.Put(ctx, "baz"); err == nil {
		log.Fatalf("Put baz when queue should have been full")
	}
	fmt.Println("Queued too full for baz")

	if v, err := ql.Get(ctx); err != nil {
		log.Fatalf("Failed to get foo: %v", err)
	} else if v != "foo" {
		log.Fatalf("Got %s, but expected foo", v)
	}
	fmt.Println("Got foo")
	// [bar]
	if err := ql.Put(ctx, "baz"); err != nil {
		log.Fatalf("Failed to put baz: %v", err)
	}
	fmt.Println("Queued baz")
	// [bar baz]

	if v, err := ql.Get(ctx); err != nil {
		log.Fatalf("Failed to get bar: %v", err)
	} else if v != "bar" {
		log.Fatalf("Got %s, but expected bar", v)
	}
	fmt.Println("Got bar")
	if v, err := ql.Get(ctx); err != nil {
		log.Fatalf("Failed to get baz: %v", err)
	} else if v != "baz" {
		log.Fatalf("Got %s, but expected baz", v)
	}
	fmt.Println("Got baz")
	// []
	l, err := ql.Len(ctx)
	if err != nil {
		log.Fatalf("Failed to get length: %v", err)
	}
	fmt.Println("Queue length", l)

	// Output:
	// Queued foo
	// Queued bar
	// Queued too full for baz
	// Got foo
	// Queued baz
	// Got bar
	// Got baz
	// Queue length 0
}

func TestMakeQueueLimiter(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		scope settings.Scope
		cre   contexts.CRE
	}{
		{settings.ScopeGlobal, contexts.CRE{}},
		{settings.ScopeWorkflow, contexts.CRE{Workflow: "wf-id"}},
	} {
		t.Run(tt.scope.String(), func(t *testing.T) {
			t.Parallel()
			mc := newMetricsChecker(t)
			f := Factory{Logger: logger.Test(t), Meter: mc.Meter(t.Name())}

			limit := settings.Int(2)
			limit.Key = "foo.bar"
			limit.Scope = tt.scope
			limit.Unit = "{task}"
			ql, err := MakeQueueLimiter[string](f, limit)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, ql.Close()) })

			ctx := t.Context()
			ctx = contexts.WithCRE(ctx, tt.cre)

			require.NoError(t, ql.Put(ctx, "foo"))
			require.NoError(t, ql.Put(ctx, "bar"))
			require.Error(t, ql.Put(ctx, "baz"))
			v, err := ql.Get(ctx)
			require.NoError(t, err)
			require.Equal(t, "foo", v)

			ms := mc.lastResourceFirstScopeMetric(t)
			redactHistogramVals[int64](t, ms, "queue.foo.bar.denied")
			attrs := attribute.NewSet(kvsFromScope(ctx, tt.scope)...)
			require.Equal(t, metrics{
				metricdata.Metrics{
					Name: "queue.foo.bar.limit",
					Unit: "{task}",
					Data: metricdata.Gauge[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: 2}},
					},
				},
				metricdata.Metrics{
					Name: "queue.foo.bar.usage",
					Unit: "{task}",
					Data: metricdata.Gauge[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: 1}},
					},
				},
				{
					Name: "queue.foo.bar.denied",
					Unit: "{task}",
					Data: metricdata.Histogram[int64]{
						DataPoints: []metricdata.HistogramDataPoint[int64]{
							{
								Attributes:   attrs,
								Count:        1,
								Bounds:       []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
								BucketCounts: []uint64{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
			}, ms)
		})
	}
}

package limits

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func ExampleRateLimiter_Allow() {
	ctx := context.Background()

	rl := GlobalRateLimiter(rate.Every(time.Second), 4)

	// Try 5
	var g errgroup.Group
	for range 5 {
		g.Go(func() error {
			if !rl.Allow(ctx) {
				return fmt.Errorf("rate limit exceeded")
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("5:", err)
	} else {
		fmt.Println("5: success")
	}

	rl = GlobalRateLimiter(rate.Every(time.Second), 4) // reset

	// Try 4
	g = errgroup.Group{}
	for range 4 {
		g.Go(func() error {
			if !rl.Allow(ctx) {
				return fmt.Errorf("rate limit exceeded")
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("4:", err)
	} else {
		fmt.Println("4: success")
	}

	// Output:
	// 5: rate limit exceeded
	// 4: success
}

func ExampleRateLimiter_AllowErr() {
	ctx := context.Background()

	rl := GlobalRateLimiter(rate.Every(time.Second), 4)

	// Try 5
	var g errgroup.Group
	for range 5 {
		g.Go(func() error {
			return rl.AllowErr(ctx)
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("5:", err)
	} else {
		fmt.Println("5: success")
	}

	rl = GlobalRateLimiter(rate.Every(time.Second), 4) // reset

	// Try 4
	g = errgroup.Group{}
	for range 4 {
		g.Go(func() error {
			return rl.AllowErr(ctx)
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("4:", err)
	} else {
		fmt.Println("4: success")
	}

	// Output:
	// 5: rate limited: request rate has exceeded the allowed limit. Please reduce request frequency or wait before retrying
	// 4: success
}

func ExampleRateLimiter_Wait() {
	ctx := context.Background()

	rl := GlobalRateLimiter(rate.Every(time.Second), 1)
	var wg sync.WaitGroup
	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			start := time.Now()
			err := rl.Wait(ctx)
			if err != nil {
				fmt.Println("rate limited:", err)
				return
			}
			fmt.Println("waited:", time.Since(start).Round(time.Second))
		}()
	}
	wg.Wait()

	// Output:
	// waited: 0s
	// waited: 1s
}

func ExampleMultiRateLimiter() {
	ctx := context.Background()
	ctxA := contexts.WithCRE(ctx, contexts.CRE{Owner: "0xabcd"})
	ctxB := contexts.WithCRE(ctx, contexts.CRE{Workflow: "ABCD"})

	global := GlobalRateLimiter(rate.Every(time.Second), 4)
	multiA := MultiRateLimiter{global, OwnerRateLimiter(rate.Every(time.Second), 4)}
	multiB := MultiRateLimiter{global, WorkflowRateLimiter(rate.Every(time.Second), 4)}

	// Try burst limit of 4 from A
	var g errgroup.Group
	for range 4 {
		g.Go(func() error {
			return multiA.AllowErr(ctxA)
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("A:", err)
	} else {
		fmt.Println("A: success")
	}

	// Try burst limit of 4 from A & B at the same time

	g = errgroup.Group{}
	for range 4 {
		g.Go(func() error {
			return multiA.AllowErr(ctxA)
		})
	}
	for range 4 {
		g.Go(func() error {
			return multiB.AllowErr(ctxB)
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("A&B:", err)
	} else {
		fmt.Println("A&B: success")
	}

	// Output:
	// A: success
	// A&B: rate limited: request rate has exceeded the allowed limit. Please reduce request frequency or wait before retrying
}

func TestFactory_NewRateLimiter(t *testing.T) {
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
			f := Factory{Meter: mc.Meter(t.Name())}
			s := settings.Rate(rate.Every(30*time.Second), 100)
			s.Key = "foo.bar"
			s.Scope = tt.scope
			s.Unit = "{action}"
			rl, err := f.MakeRateLimiter(s)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, rl.Close()) })

			ctx := t.Context()
			ctx = contexts.WithCRE(ctx, tt.cre)

			assert.NoError(t, rl.AllowErr(ctx))

			assert.Error(t, rl.AllowNErr(ctx, time.Now(), 101))

			time.Sleep(2 * pollPeriod) // ensure at least one update

			attrs := attribute.NewSet(kvsFromScope(ctx, tt.scope)...)
			ms := mc.lastResourceFirstScopeMetric(t)
			redactHistogramVals[int64](t, ms, "rate.foo.bar.denied")
			require.Equal(t, metrics{
				{
					Name: "rate.foo.bar.limit",
					Unit: "rps",
					Data: metricdata.Gauge[float64]{
						DataPoints: []metricdata.DataPoint[float64]{
							{Attributes: attrs, Value: 0.03333333333333333},
						},
					},
				},
				{
					Name: "rate.foo.bar.burst",
					Unit: "{action}",
					Data: metricdata.Gauge[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: 100},
						},
					},
				},
				{
					Name: "rate.foo.bar.usage",
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
					Name: "rate.foo.bar.denied",
					Unit: "{action}",
					Data: metricdata.Histogram[int64]{
						DataPoints: []metricdata.HistogramDataPoint[int64]{
							{
								Attributes:   attrs,
								Count:        2,
								Bounds:       []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
								BucketCounts: []uint64{0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
			}, ms)
		})
	}
}

func TestMultiRateLimiter(t *testing.T) {
	t.Parallel()

	global := GlobalRateLimiter(rate.Every(time.Second), 11)
	scoped := WorkflowRateLimiter(rate.Every(2*time.Second), 5)
	ml := MultiRateLimiter{global, scoped}
	t.Cleanup(func() { assert.NoError(t, ml.Close()) })

	ctx := t.Context()
	assert.Error(t, ml.AllowErr(ctx))

	ctxA := contexts.WithCRE(ctx, contexts.CRE{Workflow: "A"})
	ctxB := contexts.WithCRE(ctx, contexts.CRE{Workflow: "B"})

	assert.True(t, ml.Allow(ctxA))
	assert.NoError(t, ml.AllowErr(ctxB))
	assert.True(t, ml.AllowN(ctxA, time.Now(), 4))
	assert.NoError(t, ml.AllowNErr(ctxB, time.Now(), 4))

	time.Sleep(time.Second)

	ra, err := ml.Reserve(ctxA)
	if assert.NoError(t, err) {
		assert.True(t, ra.OK())
		time.Sleep(ra.Delay())
	}
	rb, err := ml.ReserveN(ctxA, time.Now(), 1)
	if assert.NoError(t, err) {
		assert.True(t, rb.OK())
		assert.True(t, ra.Allow())
		assert.NoError(t, ra.AllowErr())
		rb.CancelAt(time.Now())
	}
	require.NoError(t, ml.Wait(ctxB))
	require.NoError(t, ml.WaitN(ctxB, 2))
}

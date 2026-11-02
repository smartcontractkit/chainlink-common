package limits

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
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
				return errors.New("rate limit exceeded")
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
				return errors.New("rate limit exceeded")
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
	// 5: rate limited
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
	// A&B: rate limited
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

// TestRateLimiter_ColdStartBurst covers the first burst against a newly created bucket, which must be sized by the
// resolved setting rather than the default.
func TestRateLimiter_ColdStartBurst(t *testing.T) {
	t.Parallel()

	const (
		key      = "PerWorkflow.HTTPTrigger.RateLimit"
		org      = "org_u89K7cnwJzCek5m1"
		owner    = "0x6a4b771f7a439b38c7165bbf148542934247174c"
		workflow = "0x00aa"
		attempts = 40
	)

	for _, tt := range []struct {
		name  string
		scope settings.Scope
		toml  string
		cre   contexts.CRE
		want  int
	}{
		{
			name:  "org override above default",
			scope: settings.ScopeWorkflow,
			toml: `[org.org_u89K7cnwJzCek5m1.PerWorkflow.HTTPTrigger]
RateLimit = 'every30s:30'`,
			cre:  contexts.CRE{Org: org, Owner: owner, Workflow: workflow},
			want: 30,
		},
		{
			name:  "org override below default",
			scope: settings.ScopeWorkflow,
			toml: `[org.org_u89K7cnwJzCek5m1.PerWorkflow.HTTPTrigger]
RateLimit = 'every30s:2'`,
			cre:  contexts.CRE{Org: org, Owner: owner, Workflow: workflow},
			want: 2,
		},
		{
			name:  "no override",
			scope: settings.ScopeWorkflow,
			toml:  ``,
			cre:   contexts.CRE{Org: org, Owner: owner, Workflow: workflow},
			want:  3,
		},
		{
			name:  "org missing from context falls back to global",
			scope: settings.ScopeWorkflow,
			toml: `[global.PerWorkflow.HTTPTrigger]
RateLimit = 'every30s:1'
[org.org_u89K7cnwJzCek5m1.PerWorkflow.HTTPTrigger]
RateLimit = 'every30s:30'`,
			cre:  contexts.CRE{Owner: owner, Workflow: workflow},
			want: 1,
		},
		{
			name:  "global scope override above default",
			scope: settings.ScopeGlobal,
			toml: `[global.PerWorkflow.HTTPTrigger]
RateLimit = 'every30s:30'`,
			want: 30,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			getter, err := settings.NewTOMLGetter([]byte(tt.toml))
			require.NoError(t, err)

			s := settings.Rate(rate.Every(30*time.Second), 3)
			s.Key = key
			s.Scope = tt.scope

			rl, err := Factory{Settings: getter, Logger: logger.Test(t)}.MakeRateLimiter(s)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, rl.Close()) })

			ctx := contexts.WithCRE(t.Context(), tt.cre)
			allowed := 0
			for range attempts {
				if rl.Allow(ctx) {
					allowed++
				}
			}
			assert.Equal(t, tt.want, allowed)
		})
	}
}

// TestScopedRateLimiter_ConcurrentFirstUse checks that racing creators of one tenant's bucket end up sharing the
// stored bucket.
func TestScopedRateLimiter_ConcurrentFirstUse(t *testing.T) {
	t.Parallel()

	getter, err := settings.NewTOMLGetter([]byte(`[org.org_a.PerWorkflow.HTTPTrigger]
RateLimit = 'every30s:30'`))
	require.NoError(t, err)

	s := settings.Rate(rate.Every(30*time.Second), 3)
	s.Key = "PerWorkflow.HTTPTrigger.RateLimit"
	s.Scope = settings.ScopeWorkflow

	rl, err := Factory{Settings: getter, Logger: logger.Test(t)}.MakeRateLimiter(s)
	require.NoError(t, err)

	ctx := contexts.WithCRE(t.Context(), contexts.CRE{Org: "org_a", Owner: "0x01", Workflow: "0x02"})
	var allowed atomic.Int32
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 60 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if rl.Allow(ctx) {
				allowed.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()

	assert.Equal(t, int32(30), allowed.Load())
	assert.NoError(t, rl.Close())
}

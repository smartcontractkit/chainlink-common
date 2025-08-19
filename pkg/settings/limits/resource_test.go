package limits

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func ExampleResourcePoolLimiter_Wait() {
	ctx := context.Background()
	limiter := GlobalResourcePoolLimiter[int](5)
	ch := make(chan struct{})
	go func() { // Do 2s of work with all 5 resources
		free, err := limiter.Wait(ctx, 5)
		if err != nil {
			log.Fatalf("Failed to get resources: %v", err)
		}
		defer free()
		close(ch)
		time.Sleep(2 * time.Second)
	}()
	<-ch
	start := time.Now()
	// Blocks until goroutine frees resources
	free, err := limiter.Wait(ctx, 1)
	defer free()
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to get resources: %v", err)
	}
	fmt.Printf("Got resources after waiting: ~%s\n", elapsed.Round(time.Second))

	// Output:
	// Got resources after waiting: ~2s
}

func ExampleResourceLimiter_Use() {
	ctx := context.Background()
	limiter := GlobalResourcePoolLimiter[int](5)
	free, err := limiter.Wait(ctx, 5)
	if err != nil {
		log.Fatalf("Failed to get resources: %v", err)
	}
	defer free()

	// Returns immediately
	err = limiter.Use(ctx, 1)
	if err != nil {
		if errors.Is(err, ErrorResourceLimited[int]{}) {
			fmt.Printf("Try failed: %v\n", err)
			return
		}
		log.Fatalf("Failed to get resources: %v", err)
	}
	defer limiter.Free(ctx, 1)

	// Output:
	// Try failed: resource limited: cannot use 1, already using 5/5
}

func ExampleMultiResourcePoolLimiter() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = contexts.WithCRE(ctx, contexts.CRE{Org: "orgID", Owner: "owner-id", Workflow: "workflowID"})
	global := GlobalResourcePoolLimiter[int](100)
	freeGlobal, err := global.Wait(ctx, 95)
	if err != nil {
		log.Fatal(err)
	}
	org := OrgResourcePoolLimiter[int](50)
	freeOrg, err := org.Wait(ctx, 45)
	if err != nil {
		log.Fatal(err)
	}
	user := OwnerResourcePoolLimiter[int](20)
	freeUser, err := user.Wait(ctx, 15)
	if err != nil {
		log.Fatal(err)
	}
	workflow := WorkflowResourcePoolLimiter[int](10)
	freeWorkflow, err := workflow.Wait(ctx, 5)
	if err != nil {
		log.Fatal(err)
	}
	multi := MultiResourcePoolLimiter[int]{global, org, user, workflow}
	tryWork := func() error {
		err := multi.Use(ctx, 10)
		if err != nil {
			return err
		}
		return multi.Free(ctx, 10)
	}

	fmt.Println(tryWork())
	freeGlobal()
	fmt.Println(tryWork())
	freeOrg()
	fmt.Println(tryWork())
	freeUser()
	fmt.Println(tryWork())
	freeWorkflow()
	fmt.Println(tryWork())
	free, err := multi.Wait(ctx, 10)
	if err != nil {
		log.Fatal(err)
	}
	free()
	// Output:
	// resource limited: cannot use 10, already using 95/100
	// resource limited for org[orgID]: cannot use 10, already using 45/50
	// resource limited for owner[owner-id]: cannot use 10, already using 15/20
	// resource limited for workflow[workflowID]: cannot use 10, already using 5/10
	// <nil>
}

func TestMakeResourcePoolLimiter(t *testing.T) {
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

			limit := settings.Size(42)
			limit.Key = "foo.bar"
			limit.Scope = tt.scope
			rl, err := MakeResourcePoolLimiter(f, limit)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, rl.Close()) })

			ctx := t.Context()
			ctx = contexts.WithCRE(ctx, tt.cre)

			require.NoError(t, rl.Use(ctx, 1))
			require.NoError(t, rl.Use(ctx, 40))
			require.NoError(t, rl.Use(ctx, 1))
			require.Error(t, rl.Use(ctx, 1))
			require.NoError(t, rl.Free(ctx, 42))

			require.NoError(t, rl.Use(ctx, 42))
			require.NoError(t, rl.Free(ctx, 42))

			require.NoError(t, func(ctx context.Context) (err error) {
				ctx, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()
				_, err = rl.Wait(ctx, 42)
				return err
			}(ctx))

			ms := mc.lastResourceFirstScopeMetric(t)
			redactHistogramVals[int64](t, ms, "resource.foo.bar.amount")
			redactHistogramVals[int64](t, ms, "resource.foo.bar.denied")
			redactHistogramVals[float64](t, ms, "resource.foo.bar.block_time")
			attrs := attribute.NewSet(kvsFromScope(ctx, tt.scope)...)
			require.Equal(t, metrics{
				metricdata.Metrics{
					Name: "resource.foo.bar.limit",
					Unit: "By",
					Data: metricdata.Gauge[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: 42}},
					},
				},
				metricdata.Metrics{
					Name: "resource.foo.bar.usage",
					Unit: "By",
					Data: metricdata.Gauge[int64]{
						DataPoints: []metricdata.DataPoint[int64]{
							{Attributes: attrs, Value: 42}},
					},
				},
				metricdata.Metrics{
					Name: "resource.foo.bar.block_time",
					Unit: "By",
					Data: metricdata.Histogram[float64]{
						DataPoints: []metricdata.HistogramDataPoint[float64]{
							{
								Attributes:   attrs,
								Count:        0x5,
								Bounds:       []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
								BucketCounts: []uint64{0, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							},
						},
						Temporality: metricdata.CumulativeTemporality},
				},
				{
					Name: "resource.foo.bar.amount",
					Unit: "By",
					Data: metricdata.Histogram[int64]{
						DataPoints: []metricdata.HistogramDataPoint[int64]{
							{
								Attributes:   attrs,
								Count:        5,
								Bounds:       []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
								BucketCounts: []uint64{0, 2, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							},
						},
						Temporality: metricdata.CumulativeTemporality,
					},
				},
				{
					Name: "resource.foo.bar.denied",
					Unit: "By",
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

func TestOwnerResourcePoolLimiter(t *testing.T) {
	ctx1 := contexts.WithCRE(t.Context(), contexts.CRE{Owner: "foo"})
	ctx2 := contexts.WithCRE(t.Context(), contexts.CRE{Owner: "bar"})
	l := OwnerResourcePoolLimiter(1)
	require.NoError(t, l.Use(ctx1, 1))
	t.Cleanup(func() {
		assert.NoError(t, l.Free(ctx1, 1))
	})
	var err ErrorResourceLimited[int]
	if assert.ErrorAs(t, l.Use(ctx1, 1), &err) {
		assert.Equal(t, "", err.Key)
		assert.Equal(t, settings.ScopeOwner, err.Scope)
		assert.Equal(t, "foo", err.Tenant)
		assert.Equal(t, 1, err.Used)
		assert.Equal(t, 1, err.Limit)
		assert.Equal(t, 1, err.Amount)
	}
	require.NoError(t, l.Use(ctx2, 1))
	t.Cleanup(func() {
		assert.NoError(t, l.Free(ctx2, 1))
	})
	require.Error(t, l.Use(ctx2, 1))
}

func Test_newScopedResourcePoolLimiterFromFactory(t *testing.T) {
	ctx1 := contexts.WithCRE(t.Context(), contexts.CRE{Owner: "foo"})
	ctx2 := contexts.WithCRE(t.Context(), contexts.CRE{Owner: "bar"})
	limit := settings.Int(1)
	limit.Scope = settings.ScopeOwner
	l, err := newScopedResourcePoolLimiterFromFactory(Factory{}, limit)
	require.NoError(t, err)
	require.NoError(t, l.Use(ctx1, 1))
	t.Cleanup(func() {
		assert.NoError(t, l.Free(ctx1, 1))
	})
	var errLimited ErrorResourceLimited[int]
	if assert.ErrorAs(t, l.Use(ctx1, 1), &errLimited) {
		t.Log(errLimited)
		assert.Equal(t, "", errLimited.Key)
		assert.Equal(t, settings.ScopeOwner, errLimited.Scope)
		assert.Equal(t, "foo", errLimited.Tenant)
		assert.Equal(t, 1, errLimited.Used)
		assert.Equal(t, 1, errLimited.Limit)
		assert.Equal(t, 1, errLimited.Amount)
	}
	require.NoError(t, l.Use(ctx2, 1))
	t.Cleanup(func() {
		assert.NoError(t, l.Free(ctx2, 1))
	})
	require.Error(t, l.Use(ctx2, 1))
}

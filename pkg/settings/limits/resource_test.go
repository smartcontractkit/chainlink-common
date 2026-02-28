package limits

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
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
	// Try failed: resource limited: cannot allocate 1, already using 5 of 5 maximum. Free existing resources or request a limit increase
}

func ExampleMultiResourcePoolLimiter() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = contexts.WithCRE(ctx, contexts.CRE{Org: "org-id", Owner: "owner-id", Workflow: "workflow-id"})
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
	// resource limited: cannot allocate 10, already using 95 of 100 maximum. Free existing resources or request a limit increase
	// resource limited for org[org-id]: cannot allocate 10, already using 45 of 50 maximum. Free existing resources or request a limit increase
	// resource limited for owner[owner-id]: cannot allocate 10, already using 15 of 20 maximum. Free existing resources or request a limit increase
	// resource limited for workflow[workflow-id]: cannot allocate 10, already using 5 of 10 maximum. Free existing resources or request a limit increase
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

// TestResourcePoolLimiter_WaitOrderPreserved confirms that ResourcePoolLimiter
// preserves FIFO ordering when multiple goroutines are waiting.
func TestResourcePoolLimiter_WaitOrderPreserved(t *testing.T) {
	const numWaiters = 10

	ctx := context.Background()
	limiter := newUnscopedResourcePoolLimiter(1)

	// Channel to signal when each waiter has been enqueued
	enqueued := make(chan struct{}, numWaiters)
	limiter.resourcePoolUsage.setOnEnqueue(func() {
		enqueued <- struct{}{}
	})

	// Acquire the single resource first
	free, err := limiter.Wait(ctx, 1)
	require.NoError(t, err)

	// Track the order in which waiters acquired resources
	acquiredOrder := make(chan int, numWaiters)

	// Start multiple waiters sequentially, waiting for each to be enqueued before starting the next
	for i := range numWaiters {
		waiterID := i
		go func() {
			f, err := limiter.Wait(ctx, 1)
			if err != nil {
				return
			}
			acquiredOrder <- waiterID
			f()
		}()
		// Wait for this waiter to be enqueued before starting the next
		<-enqueued
	}

	// All waiters are now in the queue in order 0, 1, 2, ... numWaiters-1
	// Release the resource - this should wake up waiters in FIFO order
	free()

	// Collect the order in which waiters acquired the resource
	acquired := make([]int, 0, numWaiters)
	for range numWaiters {
		select {
		case id := <-acquiredOrder:
			acquired = append(acquired, id)
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for waiter to acquire resource")
		}
	}

	// Verify FIFO order is preserved
	for i, id := range acquired {
		require.Equalf(t, i, id, "expected waiter %d at position %d, got %d (acquired order: %v)", i, i, id, acquired)
	}
}

// TestResourcePoolLimiter_ContextCancellation tests that context cancellation
// properly removes waiters from the queue without breaking FIFO order.
func TestResourcePoolLimiter_ContextCancellation(t *testing.T) {
	ctx := context.Background()
	limiter := newUnscopedResourcePoolLimiter(1)

	// Channel to signal when each waiter has been enqueued
	enqueued := make(chan struct{}, 5)
	limiter.resourcePoolUsage.setOnEnqueue(func() {
		enqueued <- struct{}{}
	})

	// Acquire the single resource
	free, err := limiter.Wait(ctx, 1)
	require.NoError(t, err)

	// Start 5 waiters, but cancel the middle one
	results := make(chan struct {
		id  int
		err error
	}, 5)

	var ctxs []context.Context
	var cancels []context.CancelFunc
	for range 5 {
		c, cancel := context.WithCancel(ctx)
		ctxs = append(ctxs, c)
		cancels = append(cancels, cancel)
	}

	// Start waiters sequentially, waiting for each to be enqueued
	for i := range 5 {
		waiterID := i
		go func() {
			f, err := limiter.Wait(ctxs[waiterID], 1)
			if err != nil {
				results <- struct {
					id  int
					err error
				}{waiterID, err}
				return
			}
			results <- struct {
				id  int
				err error
			}{waiterID, nil}
			f()
		}()
		// Wait for this waiter to be enqueued
		<-enqueued
	}

	// All 5 waiters are now in the queue in order 0, 1, 2, 3, 4
	// Cancel waiter 2 (middle of the queue) and wait for the cancellation result
	cancels[2]()
	cancelResult := <-results
	require.Equal(t, 2, cancelResult.id)
	require.Error(t, cancelResult.err)

	// Release the resource - remaining waiters should acquire in FIFO order
	free()

	// Collect remaining results
	var acquiredIDs []int
	for range 4 {
		select {
		case r := <-results:
			require.NoError(t, r.err)
			acquiredIDs = append(acquiredIDs, r.id)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for results")
		}
	}

	// Remaining waiters should acquire in order: 0, 1, 3, 4
	assert.Equal(t, []int{0, 1, 3, 4}, acquiredIDs, "waiters should acquire in FIFO order, skipping cancelled")
}

// TestResourcePoolLimiter_BasicUsage tests basic Use/Free functionality.
func TestResourcePoolLimiter_BasicUsage(t *testing.T) {
	ctx := context.Background()
	limiter := GlobalResourcePoolLimiter(5)

	// Use should work
	require.NoError(t, limiter.Use(ctx, 3))

	// Available should report 2
	avail, err := limiter.Available(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, avail)

	// Using more than available should fail
	err = limiter.Use(ctx, 3)
	require.Error(t, err)
	var limitErr ErrorResourceLimited[int]
	require.ErrorAs(t, err, &limitErr)
	assert.Equal(t, 3, limitErr.Used)
	assert.Equal(t, 5, limitErr.Limit)
	assert.Equal(t, 3, limitErr.Amount)

	// Free should work
	require.NoError(t, limiter.Free(ctx, 3))

	// Now should have 5 available
	avail, err = limiter.Available(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, avail)
}

// TestResourcePoolLimiter_LimitFlapToZeroDoesNotDeadlock verifies that a waiter
// is woken up when the limit is reduced to zero and then increased again.
func TestResourcePoolLimiter_LimitFlapToZeroDoesNotDeadlock(t *testing.T) {
	t.Parallel()

	var limit atomic.Int64
	limit.Store(1)

	limiter := newUnscopedResourcePoolLimiter(1)
	limiter.getLimitFn = func(context.Context) (int, error) {
		return int(limit.Load()), nil
	}
	go limiter.updateLoop(t.Context())
	t.Cleanup(func() { assert.NoError(t, limiter.Close()) })

	ctx := t.Context()

	// Consume the single available resource to force the next waiter to enqueue.
	freeFirst, err := limiter.Wait(ctx, 1)
	require.NoError(t, err)

	enqueued := make(chan struct{}, 1)
	limiter.resourcePoolUsage.setOnEnqueue(func() { enqueued <- struct{}{} })

	waitErr := make(chan error, 1)
	go func() {
		_, err := limiter.Wait(t.Context(), 1)
		waitErr <- err
	}()

	// Ensure the waiter is queued before mutating the limit.
	<-enqueued

	// Drop the limit to zero, then free the first resource. The queued waiter
	// remains blocked because tryWakeWaiters sees a zero limit.
	limit.Store(0)
	freeFirst()

	// Raise the limit again; the queued waiter should be woken by the update.
	limit.Store(1)

	select {
	case err := <-waitErr:
		require.NoError(t, err)
		// release to avoid affecting subsequent waits
		_ = limiter.Free(ctx, 1)
	case <-time.After(pollPeriod * 3):
		t.Fatal("waiter did not return after limit flap")
	}
}

// setOnEnqueue sets a callback that is invoked each time a waiter is added to the queue.
// The callback is called with the mutex held. Used for testing to synchronize without sleeps.
func (u *resourcePoolUsage[N]) setOnEnqueue(fn func()) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.onEnqueue = fn
}

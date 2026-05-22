package host

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	wfpb "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

// slowCapStub delays CallCapability by a configurable duration and counts in-flight calls.
type slowCapStub struct {
	delay     time.Duration
	inflight  atomic.Int32
	peakLoad  atomic.Int32
	callCount atomic.Int32
}

func (s *slowCapStub) CallCapability(_ context.Context, _ *sdkpb.CapabilityRequest) (*sdkpb.CapabilityResponse, error) {
	s.callCount.Add(1)
	cur := s.inflight.Add(1)
	for {
		peak := s.peakLoad.Load()
		if cur <= peak || s.peakLoad.CompareAndSwap(peak, cur) {
			break
		}
	}
	time.Sleep(s.delay)
	s.inflight.Add(-1)

	payload, _ := anypb.New(&emptypb.Empty{})
	return &sdkpb.CapabilityResponse{
		Response: &sdkpb.CapabilityResponse_Payload{Payload: payload},
	}, nil
}

func (s *slowCapStub) GetSecrets(context.Context, *sdkpb.GetSecretsRequest) ([]*sdkpb.SecretResponse, error) {
	return nil, nil
}
func (s *slowCapStub) GetWorkflowExecutionID() string { return "test-exec" }
func (s *slowCapStub) GetNodeTime() time.Time         { return time.Now() }
func (s *slowCapStub) GetDONTime() (time.Time, error) { return time.Now(), nil }
func (s *slowCapStub) EmitUserLog(string) error       { return nil }
func (s *slowCapStub) EmitUserMetric(context.Context, *wfpb.WorkflowUserMetric) error {
	return nil
}

var _ ExecutionHelper = (*slowCapStub)(nil)

func newTestExec(maxPending int, stub ExecutionHelper) *execution[*sdkpb.ExecutionResult] {
	return &execution[*sdkpb.ExecutionResult]{
		ctx:                 context.Background(),
		capabilityResponses: make(map[int32]<-chan *sdkpb.CapabilityResponse),
		secretsResponses:    make(map[int32]<-chan *secretsResponse),
		pendingCallsLimiter: limits.GlobalResourcePoolLimiter[int](maxPending),
		executor:            stub,
	}
}

// TestSemaphore_BackpressureBlocksCallN proves that call N+1 blocks when
// N == MaxPendingCalls and nothing has been awaited yet.
func TestSemaphore_BackpressureBlocksCallN(t *testing.T) {
	t.Parallel()
	const max = 5

	// Use a delay longer than the check window so goroutines hold their slots.
	stub := &slowCapStub{delay: 5 * time.Second}
	exec := newTestExec(max, stub)

	ctx := t.Context()

	// Fill semaphore.
	for i := int32(0); i < max; i++ {
		require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: i}))
	}

	// Next call should block.
	blocked := make(chan struct{})
	go func() {
		_ = exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: max})
		close(blocked)
	}()

	select {
	case <-blocked:
		t.Fatal("call max+1 did not block; semaphore backpressure broken")
	case <-time.After(200 * time.Millisecond):
		// expected — still blocked
	}

	// Await the first call to free a slot.
	resp, err := exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{0}})
	require.NoError(t, err)
	require.Len(t, resp.Responses, 1)

	// Now the blocked call should proceed.
	select {
	case <-blocked:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("call max+1 did not unblock after await freed a slot")
	}
}

// TestSemaphore_HighThroughputBounded issues many calls in batches,
// awaiting each batch before the next. Peak in-flight goroutines must never
// exceed MaxPendingCalls.
func TestSemaphore_HighThroughputBounded(t *testing.T) {
	t.Parallel()
	const max = 10
	const batches = 50
	const callsPerBatch = max

	stub := &slowCapStub{delay: 1 * time.Millisecond}
	exec := newTestExec(max, stub)

	ctx := t.Context()
	var callId int32

	for b := 0; b < batches; b++ {
		ids := make([]int32, callsPerBatch)
		for i := 0; i < callsPerBatch; i++ {
			ids[i] = callId
			require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: callId}))
			callId++
		}
		resp, err := exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{Ids: ids})
		require.NoError(t, err)
		require.Len(t, resp.Responses, callsPerBatch)
	}

	assert.LessOrEqual(t, int(stub.peakLoad.Load()), max,
		"peak in-flight goroutines exceeded MaxPendingCalls")
	assert.Equal(t, int32(batches*callsPerBatch), stub.callCount.Load())
}

// TestSemaphore_ContextCancelUnblocksCall proves that a blocked callCapAsync
// returns ctx.Err() when the context is cancelled.
func TestSemaphore_ContextCancelUnblocksCall(t *testing.T) {
	t.Parallel()
	const max = 2

	stub := &slowCapStub{delay: 5 * time.Second} // very slow, won't finish
	exec := newTestExec(max, stub)

	ctx, cancel := context.WithCancel(t.Context())

	// Fill semaphore.
	for i := int32(0); i < max; i++ {
		require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: i}))
	}

	// Next call will block on semaphore.
	var callErr error
	done := make(chan struct{})
	go func() {
		callErr = exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: max})
		close(done)
	}()

	// Cancel context.
	cancel()

	select {
	case <-done:
		require.ErrorIs(t, callErr, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("callCapAsync did not unblock after context cancel")
	}
}

// TestSemaphore_SlotsRecycledCorrectly ensures that after many await cycles,
// the semaphore is back to its full capacity and new calls can proceed.
func TestSemaphore_SlotsRecycledCorrectly(t *testing.T) {
	t.Parallel()
	const max = 5
	const rounds = 100

	stub := &slowCapStub{delay: 0}
	exec := newTestExec(max, stub)

	ctx := t.Context()

	for r := 0; r < rounds; r++ {
		ids := make([]int32, max)
		for i := int32(0); i < max; i++ {
			id := int32(r*max) + i
			ids[i] = id
			require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: id}))
		}
		_, err := exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{Ids: ids})
		require.NoError(t, err)
	}

	// After all rounds, all slots should be available again.
	// Goroutines release slots via defer after the channel send, so allow a
	// brief window for the last batch of defers to execute.
	assert.Eventually(t, func() bool {
		avail, err := exec.pendingCallsLimiter.Available(ctx)
		return err == nil && avail == max
	}, time.Second, 5*time.Millisecond,
		"limiter still has occupied slots after all awaits completed")
}

// TestSemaphore_MapCleanedOnAwait verifies the capabilityResponses map
// doesn't leak entries.
func TestSemaphore_MapCleanedOnAwait(t *testing.T) {
	t.Parallel()
	const max = 10
	const total = 200

	stub := &slowCapStub{delay: 0}
	exec := newTestExec(max, stub)

	ctx := t.Context()

	for i := int32(0); i < total; i += max {
		ids := make([]int32, max)
		for j := int32(0); j < max; j++ {
			id := i + j
			ids[j] = id
			require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: id}))
		}
		_, err := exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{Ids: ids})
		require.NoError(t, err)
	}

	exec.lock.RLock()
	mapLen := len(exec.capabilityResponses)
	exec.lock.RUnlock()

	assert.Equal(t, 0, mapLen, "capabilityResponses map leaked %d entries", mapLen)
}

// TestSemaphore_ConcurrentCallAndAwait exercises concurrent callers issuing
// callCapAsync from multiple goroutines while others await, simulating the
// real engine dispatching multiple workflow executions.
func TestSemaphore_ConcurrentCallAndAwait(t *testing.T) {
	t.Parallel()
	const max = 10
	const workers = 20
	const callsPerWorker = 50

	stub := &slowCapStub{delay: 10 * time.Microsecond}
	// Each worker gets its own execution (like real CRE — one per WASM invocation).
	// We want to prove that WITHIN a single execution, concurrent isn't needed because
	// WASM is single-threaded. But let's stress the shared semaphore anyway.
	exec := newTestExec(max, stub)

	ctx := t.Context()
	var wg sync.WaitGroup

	// Simulate sequential call-then-await pattern from a single WASM thread
	// (the real case). We run it in parallel workers to stress-test the lock.
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < callsPerWorker; i++ {
				id := int32(workerID*callsPerWorker + i)
				err := exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: id})
				if err != nil {
					return
				}
				_, err = exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{id}})
				if err != nil {
					return
				}
			}
		}(w)
	}

	wg.Wait()

	assert.LessOrEqual(t, int(stub.peakLoad.Load()), max)
	assert.Equal(t, int32(workers*callsPerWorker), stub.callCount.Load())
	avail, err := exec.pendingCallsLimiter.Available(context.Background())
	require.NoError(t, err)
	assert.Equal(t, max, avail)
}

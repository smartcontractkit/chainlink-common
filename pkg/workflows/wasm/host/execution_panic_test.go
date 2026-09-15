package host

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	wfpb "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

// panicCapStub panics from the two ExecutionHelper methods the host dispatches
// on goroutines of its own.
type panicCapStub struct{ msg string }

var _ ExecutionHelper = (*panicCapStub)(nil)

func (p *panicCapStub) CallCapability(context.Context, *sdkpb.CapabilityRequest) (*sdkpb.CapabilityResponse, error) {
	panic(p.msg)
}

func (p *panicCapStub) GetSecrets(context.Context, *sdkpb.GetSecretsRequest) ([]*sdkpb.SecretResponse, error) {
	panic(p.msg)
}

func (p *panicCapStub) GetWorkflowExecutionID() string { return "test-exec" }
func (p *panicCapStub) GetNodeTime() time.Time         { return time.Now() }
func (p *panicCapStub) GetDONTime() (time.Time, error) { return time.Now(), nil }
func (p *panicCapStub) EmitUserLog(string) error       { return nil }
func (p *panicCapStub) EmitUserMetric(context.Context, *wfpb.WorkflowUserMetric) error {
	return nil
}

// TestCallCapAsync_PanicBecomesCapabilityError proves a panicking
// ExecutionHelper surfaces as a capability error rather than reaching the
// runtime. The goroutine callCapAsync spawns carries no wasmtime frame, so
// neither wasmtime-go's host-callback recover nor callStart's recover can see a
// panic there; unrecovered it would terminate the process.
func TestCallCapAsync_PanicBecomesCapabilityError(t *testing.T) {
	t.Parallel()

	exec := newTestExec(2, &panicCapStub{msg: "boom"})
	ctx := t.Context()

	require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: 1}))

	resp, err := exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{1}})
	require.NoError(t, err)
	require.Contains(t, resp.Responses, int32(1))
	assert.Contains(t, resp.Responses[1].GetError(), "panic in capability call")
	assert.Contains(t, resp.Responses[1].GetError(), "boom")
}

// A panic must still put a value on the response channel, or the guest blocks in
// awaitCapabilities until the execution deadline instead of seeing a failed call.
func TestCallCapAsync_PanicDoesNotStallAwait(t *testing.T) {
	t.Parallel()

	exec := newTestExec(2, &panicCapStub{msg: "boom"})
	ctx := t.Context()
	require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: 1}))

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{1}})
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("awaitCapabilities stalled after a panicking capability call")
	}
}

// The limiter slot is released by `defer free()`, which only runs if the panic is
// recovered inside the goroutine rather than unwinding it.
func TestCallCapAsync_PanicReleasesLimiterSlot(t *testing.T) {
	t.Parallel()

	exec := newTestExec(1, &panicCapStub{msg: "boom"})
	ctx := t.Context()

	require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: 1}))
	_, err := exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{1}})
	require.NoError(t, err)

	// With maxPending=1 this blocks forever if the slot leaked.
	second := make(chan error, 1)
	go func() { second <- exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: 2}) }()

	select {
	case err := <-second:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("limiter slot was not released after a panicking capability call")
	}
}

func TestGetSecretsAsync_PanicBecomesError(t *testing.T) {
	t.Parallel()

	exec := newTestExec(2, &panicCapStub{msg: "boom"})
	ctx := t.Context()

	require.NoError(t, exec.getSecretsAsync(ctx, &sdkpb.GetSecretsRequest{CallbackId: 1}))

	_, err := exec.awaitSecrets(ctx, &sdkpb.AwaitSecretsRequest{Ids: []int32{1}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic in get secrets")
	assert.Contains(t, err.Error(), "boom")
}

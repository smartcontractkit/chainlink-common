package host

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	wfpb "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

// panicCapStub panics from CallCapability, the method the host dispatches on a
// goroutine of its own.
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

const (
	// asyncPanicChildEnv marks the re-executed copy of this test binary as the
	// subprocess that actually runs the panicking capability call.
	asyncPanicChildEnv = "CRE_ASYNC_PANIC_CHILD"

	// asyncPanicDispatchedMarker is printed once the guest's call_capability has
	// returned, i.e. once wasmtime is off the stack.
	asyncPanicDispatchedMarker = "ASYNC_PANIC_DISPATCHED"

	// asyncPanicSurvivedMarker is printed only if the process is still alive
	// after the spawned goroutine panicked.
	asyncPanicSurvivedMarker = "ASYNC_PANIC_SURVIVED"

	asyncPanicMessage = "boom-from-capability"
)

func TestCallCapability_AsyncPanicEscapesWasmtimeRecover(t *testing.T) {
	if os.Getenv(asyncPanicChildEnv) == "1" {
		runAsyncPanicChild(t)
		return
	}

	cmd := osexec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v=true")
	cmd.Env = append(os.Environ(), asyncPanicChildEnv+"=1")
	out, runErr := cmd.CombinedOutput()
	output := string(out)
	t.Logf("subprocess output:\n%s", output)

	require.Contains(t, output, asyncPanicDispatchedMarker,
		"call_capability must have returned cleanly before the panic; without that "+
			"the test is not exercising the async path at all")

	// The child must have died, and died by panic rather than by a failed
	// assertion: Go exits 2 on an unrecovered panic.
	var exitErr *osexec.ExitError
	require.ErrorAs(t, runErr, &exitErr,
		"the panic is expected to kill the child process; a clean exit means "+
			"something recovered it and this test needs inverting")
	assert.Equal(t, 2, exitErr.ExitCode(), "Go exits 2 on an unrecovered panic")

	assert.Contains(t, output, "panic: "+asyncPanicMessage,
		"the child should die with the capability's panic, not some other failure")

	// The frame that settles the argument: the panicking goroutine was created by
	// callCapAsync, so no wasmtime frame is on its stack and neither wasmtime-go's
	// callback recover nor callStart can see it.
	assert.Contains(t, output, "created by",
		"the traceback should show the panic came from a spawned goroutine")
	assert.Contains(t, output, "callCapAsync",
		"the panicking goroutine should be the one callCapAsync spawned")

	assert.NotContains(t, output, asyncPanicSurvivedMarker,
		"the child should not have reached the point after awaitCapabilities")
}

// runAsyncPanicChild drives a real wasmtime instance through the production
// call_capability host function with an ExecutionHelper that panics.
func runAsyncPanicChild(t *testing.T) {
	lggr := logger.Test(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	e := &execution[*sdkpb.ExecutionResult]{
		ctx:                 ctx,
		capabilityResponses: make(map[int32]<-chan *sdkpb.CapabilityResponse),
		secretsResponses:    make(map[int32]<-chan *secretsResponse),
		usedCallbackIDs:     make(map[string]bool),
		pendingCallsLimiter: limits.GlobalResourcePoolLimiter[int](4),
		executor:            &panicCapStub{msg: asyncPanicMessage},
	}

	store, inst, mem := instantiateCallCapModule(t, watCallCapV2Test, e, lggr)

	req := &sdkpb.CapabilityRequest{Id: "test-cap@1.0.0", CallbackId: 1}
	reqBytes, err := proto.Marshal(req)
	require.NoError(t, err)

	const reqOffset, respOffset, respSize = int32(0), int32(256), int32(256)
	memWrite(t, mem, store, reqOffset, reqBytes)

	callCap := inst.GetExport(store, "call_cap").Func()
	result, err := callCap.Call(store, reqOffset, int32(len(reqBytes)), respOffset, respSize)

	// The host function only dispatches, so it succeeds even though the
	// capability call it started is about to panic.
	require.NoError(t, err, "call_capability should not trap: it dispatches and returns")
	require.Equal(t, int64(0), result.(int64), "async dispatch should report success")
	fmt.Println(asyncPanicDispatchedMarker)

	// Func.Call has returned, so wasmtime is no longer on any stack. Blocking here
	// is what makes the test deterministic rather than a sleep race: the goroutine
	// can only deliver a response if its panic was recovered. If it was not, the
	// process is already gone and this line never returns.
	resp, err := e.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{1}})
	require.NoError(t, err)
	require.Contains(t, resp.Responses, int32(1))
	require.Contains(t, resp.Responses[1].GetError(), asyncPanicMessage,
		"the recovered panic should surface as the capability's error")

	fmt.Println(asyncPanicSurvivedMarker, resp.Responses[1].GetError())
}

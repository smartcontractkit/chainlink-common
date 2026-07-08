package host

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

// TestStandardNonDeterministicInput verifies that an execution which produces a
// different capability request on resume than on its original run (i.e. a
// non-deterministic workflow) is rejected by the host's integrity check.
//
// The standard test guest builds its capability request from the DON random
// seed. We run the real Execute loop but wrap module.callWasm so that, just
// before the resume, we flip exec.donSeed. The replayed run then issues a
// different request for the same callback id, which fails the proto.Equal
// integrity check in callCapAsync, surfaces to the guest as a failed capability
// call, and ends the execution in error.
func TestStandardNonDeterministicInput(t *testing.T) {
	t.Parallel()

	lggr, observed := logger.TestObserved(t, zapcore.ErrorLevel)
	cfg := &ModuleConfig{
		Logger:              lggr,
		IsUncompressed:      true,
		PendingCallsLimiter: limits.GlobalResourcePoolLimiter(cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit.DefaultValue),
	}

	m := makeTestModuleByName(t, testPath, "non_deterministic_input", cfg, false)
	m.Start()
	defer m.Close()

	helper := mocks.NewMockExecutionHelper(t)
	helper.EXPECT().GetWorkflowExecutionID().Return("id")
	helper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()
	helper.EXPECT().GetDONTime().RunAndReturn(func() (time.Time, error) {
		return time.Now(), nil
	}).Maybe()
	helper.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ *sdkpb.CapabilityRequest) (*sdkpb.CapabilityResponse, error) {
		payload, perr := anypb.New(&basicaction.Outputs{AdaptedThing: "ok"})
		require.NoError(t, perr)
		return &sdkpb.CapabilityResponse{
			Response: &sdkpb.CapabilityResponse_Payload{Payload: payload},
		}, nil
	}).Maybe()

	// Wrap callWasm so the workflow becomes non-deterministic across the resume:
	// flip the DON seed before the second (resume) invocation, so the replayed
	// run builds a different capability request than the original. Flipping by one
	// guarantees the seed's parity - and therefore the request's input - changes.
	original := m.callWasm
	var runs int
	m.callWasm = func(timeout time.Duration, req *sdkpb.ExecuteRequest, linkWasm linkFn[*sdkpb.ExecutionResult], exec *execution[*sdkpb.ExecutionResult]) (time.Duration, error) {
		runs++
		if runs == 2 {
			exec.donSeed++
		}
		return original(timeout, req, linkWasm, exec)
	}

	req := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})
	// Suspension is opted into per-execution via the request, not the module config.
	req.SuspendOnAwait = true
	result, err := m.Execute(t.Context(), req, helper)
	require.NoError(t, err)

	// The execution ends in error. The exact message is SDK-specific (the rawsdk
	// reports "callCapability returned an error", cre-sdk-go "cannot find
	// capability ..."), so we only assert that it is an error result; the
	// determinism violation itself is asserted via the host log below.
	errResult, ok := result.Result.(*sdkpb.ExecutionResult_Error)
	require.True(t, ok, "expected an error result, got %T", result.Result)
	require.NotEmpty(t, errResult.Error)

	// The execution suspended once and resumed once, and the host logged the
	// underlying determinism violation.
	require.Equal(t, 2, runs, "expected the guest to run twice: the suspended run and the resumed run")
	require.NotEmpty(t, observed.FilterMessageSnippet("non-determinism error").All(),
		"expected the host to log a non-determinism error")
}

package host

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

// TestStandardSuspendedExecutions runs a real rawsdk workflow that dispatches an
// async capability call and awaits it, and asserts the host's suspend/resume
// behaviour end-to-end:
//
//   - With suspension disabled the await blocks in the host, so the guest runs
//     exactly once (one log line).
//   - With suspension enabled the host has no response at the await, so it
//     suspends the guest and resumes it once the response is available. The
//     guest therefore runs twice (two log lines), while the capability itself is
//     dispatched only once - the second run replays the recorded call from the
//     store rather than re-invoking the capability.
func TestStandardSuspendedExecutions(t *testing.T) {
	t.Parallel()

	const adaptedThing = "adapted-thing"
	const runLogMessage = "suspended_executions:run"

	run := func(t *testing.T, suspensionEnabled bool) (result string, capabilityCalls, runs int32) {
		var capabilityCallCount, runCount atomic.Int32

		helper := mocks.NewMockExecutionHelper(t)
		helper.EXPECT().GetWorkflowExecutionID().Return("id")
		helper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			return time.Now()
		}).Maybe()
		helper.EXPECT().GetDONTime().RunAndReturn(func() (time.Time, error) {
			return time.Now(), nil
		}).Maybe()
		helper.EXPECT().EmitUserLog(mock.Anything).RunAndReturn(func(s string) error {
			// Match on Contains rather than equality: the rawsdk emits the raw
			// message, while other SDKs (e.g. cre-sdk-go) emit a formatted log line
			// that contains it.
			if strings.Contains(s, runLogMessage) {
				runCount.Add(1)
			}
			return nil
		})
		helper.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, req *sdk.CapabilityRequest) (*sdk.CapabilityResponse, error) {
			capabilityCallCount.Add(1)
			assert.Equal(t, "basic-test-action@1.0.0", req.Id)
			assert.Equal(t, "PerformAction", req.Method)

			payload, err := anypb.New(&basicaction.Outputs{AdaptedThing: adaptedThing})
			require.NoError(t, err)
			return &sdk.CapabilityResponse{
				Response: &sdk.CapabilityResponse_Payload{Payload: payload},
			}, nil
		})

		cfg := defaultNoDAGModCfg(t)
		m := makeTestModuleByName(t, testPath, "suspended_executions", cfg, false)
		m.Start()
		defer m.Close()

		req := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})
		// Suspension is opted into per-execution via the request, not the module config.
		req.SuspendOnAwait = suspensionEnabled
		result = executeWithResult[string](t, m, req, helper)
		return result, capabilityCallCount.Load(), runCount.Load()
	}

	t.Run("suspension disabled blocks in the host and runs once", func(t *testing.T) {
		result, capabilityCalls, runs := run(t, false)
		assert.Equal(t, adaptedThing, result)
		assert.Equal(t, int32(1), capabilityCalls, "capability should be dispatched once")
		assert.Equal(t, int32(1), runs, "guest should run exactly once when suspension is disabled")
	})

	t.Run("suspension enabled suspends on await and resumes with the response", func(t *testing.T) {
		result, capabilityCalls, runs := run(t, true)
		assert.Equal(t, adaptedThing, result)
		assert.Equal(t, int32(1), capabilityCalls, "capability should be dispatched once and replayed from the store on resume")
		assert.Equal(t, int32(2), runs, "guest should run twice: the suspended run and the resumed run")
	})
}

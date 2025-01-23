package host

import (
	_ "embed"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

const (
	nodagBinaryLocation = "test/nodag/testmodule.wasm"
	nodagBinaryCmd      = "test/nodag"
)

func Test_V2_Run(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	mc := &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
	}
	capResponsesSoFar := map[string]*pb.CapabilityResponse{}
	triggerRef := "trigger"
	binary := createTestBinary(nodagBinaryCmd, nodagBinaryLocation, true, t)

	// (1) Engine calls GetWorkflowSpec() first.
	spec, err := GetWorkflowSpec(ctx, mc, binary, []byte(""))
	require.NoError(t, err)

	// (2) Engine expects only triggers to be included in the returned spec. Trigger subscriptions are performed.
	//
	//  [trigger]
	//      |
	//  (promise)
	require.Len(t, spec.Triggers, 1)
	require.Equal(t, triggerRef, spec.Triggers[0].Ref)
	require.Len(t, spec.Actions, 0)
	require.Len(t, spec.Consensus, 0)
	require.Len(t, spec.Targets, 0)

	m, err := NewModule(mc, binary)
	require.NoError(t, err)
	m.Start()

	// (3) When a TriggerEvent occurs, Engine calls Run() with that Event.
	triggerEvent := &pb.TriggerEvent{
		TriggerType: "my_trigger@1.0.0",
		Outputs:     values.ProtoMap(values.EmptyMap()),
	}

	req := newRunRequest(triggerRef, triggerEvent, capResponsesSoFar)
	resp, err := m.Run(ctx, req)
	require.NoError(t, err)
	runResp := resp.GetRunResponse()
	require.NotNil(t, runResp)

	// (4) In the first response, the Workflow requests two action capability calls.
	//
	//     [trigger]
	//         |
	//     [compute1]
	//       /    \
	// [action1] [action2]
	//      \     /
	//     (promise)
	require.Len(t, runResp.RefToCapCall, 2)
	require.Contains(t, runResp.RefToCapCall, "ref_action1")
	require.Contains(t, runResp.RefToCapCall, "ref_action2")

	// (5) Engine now makes capability calls and when they are ready, it invokes Run() again with both responses.
	capResponsesSoFar["ref_action1"] = pb.CapabilityResponseToProto(capabilities.CapabilityResponse{Value: &values.Map{}})
	capResponsesSoFar["ref_action2"] = pb.CapabilityResponseToProto(capabilities.CapabilityResponse{Value: &values.Map{}})
	req = newRunRequest(triggerRef, triggerEvent, capResponsesSoFar)
	resp, err = m.Run(ctx, req)
	require.NoError(t, err)
	runResp = resp.GetRunResponse()
	require.NotNil(t, runResp)

	// (6) Workflow now requests a target capability call.
	//
	//     [trigger]
	//         |
	//     [compute1]
	//       /    \
	// [action1] [action2]
	//      \     /
	//     [compute2]
	//         |
	//     [target1]
	//         |
	//     (promise)
	require.Len(t, runResp.RefToCapCall, 1)
	require.Contains(t, runResp.RefToCapCall, "ref_target1")

	// (7) After calling the target, Engine makes one last Run() call and expects the workflow to complete without errors.
	capResponsesSoFar["ref_target1"] = pb.CapabilityResponseToProto(capabilities.CapabilityResponse{Value: &values.Map{}})
	req = newRunRequest(triggerRef, triggerEvent, capResponsesSoFar)
	resp, err = m.Run(ctx, req)
	require.NoError(t, err)
	runResp = resp.GetRunResponse()
	require.Nil(t, runResp)
}

func newRunRequest(triggerRef string, triggerEvent *pb.TriggerEvent, capResponsesSoFar map[string]*pb.CapabilityResponse) *wasmpb.Request {
	return &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_RunRequest{
			RunRequest: &wasmpb.RunRequest{
				TriggerRef:    triggerRef,
				TriggerEvent:  triggerEvent,
				RefToResponse: capResponsesSoFar,
			},
		},
	}
}

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
	calls := make(chan *wasmpb.CapabilityCall, 10)
	awaitReq := make(chan *wasmpb.AwaitRequest, 1)
	awaitResp := make(chan *wasmpb.AwaitResponse, 1)
	// TODO this shouldn't live here
	on := int32(0)
	mc := &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
		CallCapAsync: func(req *wasmpb.CapabilityCall) (int32, error) {
			calls <- req
			on++
			return on, nil
		},
		AwaitCaps: func(req *wasmpb.AwaitRequest) (*wasmpb.AwaitResponse, error) {
			awaitReq <- req
			return <-awaitResp, nil
		},
	}
	capResponses := map[int32]*pb.CapabilityResponse{}
	triggerID := "basic-trigger@1.0.0"
	triggerRef := "trigger-0"
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
	require.Equal(t, triggerID, spec.Triggers[0].ID)
	require.Equal(t, triggerRef, spec.Triggers[0].Ref)
	require.Len(t, spec.Actions, 0)
	require.Len(t, spec.Consensus, 0)
	require.Len(t, spec.Targets, 0)

	m, err := NewModule(mc, binary)
	require.NoError(t, err)
	m.Start()

	// (3) When a TriggerEvent occurs, Engine calls Run() with that Event.
	triggerEvent := &pb.TriggerEvent{
		TriggerType: triggerID,
		Outputs:     values.ProtoMap(values.EmptyMap()),
	}

	doneCh := make(chan struct{})
	go func() {
		req := newRunRequest(triggerRef, triggerEvent, capResponses)
		_, err := m.Run(ctx, req)
		require.NoError(t, err)
		close(doneCh)
	}()

	// (4) The workflow makes two capability calls and then awaits them.
	//
	//     [trigger]
	//         |
	//     [compute1]
	//       /    \
	// [action1] [action2]
	//      \     /
	//     (promise)
	call1 := <-calls
	require.Equal(t, "basicaction@1.0.0", call1.CapabilityId)
	call2 := <-calls
	require.Equal(t, "basicaction@1.0.0", call2.CapabilityId)

	// (5) Engine performs async capability calls.
	capResponses[1] = pb.CapabilityResponseToProto(capabilities.CapabilityResponse{Value: &values.Map{}})
	capResponses[2] = pb.CapabilityResponseToProto(capabilities.CapabilityResponse{Value: &values.Map{}})
	awaitReqMsg := <-awaitReq
	require.Len(t, awaitReqMsg.Refs, 2)
	awaitResp <- &wasmpb.AwaitResponse{RefToResponse: capResponses}

	// (6) Workflow now makes a target capability call.
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
	call3 := <-calls
	require.Equal(t, "basictarget@1.0.0", call3.CapabilityId)

	// (7) Engine performs the call.
	capResponses[3] = pb.CapabilityResponseToProto(capabilities.CapabilityResponse{Value: &values.Map{}})
	awaitReqMsg = <-awaitReq
	require.Len(t, awaitReqMsg.Refs, 1)
	awaitResp <- &wasmpb.AwaitResponse{RefToResponse: capResponses}

	<-doneCh
}

func newRunRequest(triggerRef string, triggerEvent *pb.TriggerEvent, capResponsesSoFar map[int32]*pb.CapabilityResponse) *wasmpb.Request {
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

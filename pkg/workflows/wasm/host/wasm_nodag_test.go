package host

import (
	"context"
	_ "embed"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/testhelpers"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

// TODO this should go generate the capability clients

const (
	nodagBinaryLocation = "test/nodag/cmd/testmodule.wasm"
	nodagBinaryCmd      = "test/nodag/cmd"
)

const anyNoDagExecId = "executionId"

func Test_NoDag_Run(t *testing.T) {
	t.Parallel()
	mc := createNoDagMc(t)

	triggerID := "basic-test-trigger@1.0.0"
	binary := createTestBinary(nodagBinaryCmd, nodagBinaryLocation, true, t)

	ctx := t.Context()
	triggers, err := GetTriggersSpec(ctx, mc, binary, []byte(""))
	require.NoError(t, err)

	require.Len(t, triggers.Subscriptions, 1)
	require.Equal(t, triggerID, triggers.Subscriptions[0].Id)
	configProto := triggers.Subscriptions[0].Payload
	config := &basictrigger.Config{}
	require.NoError(t, configProto.UnmarshalTo(config))
	require.Equal(t, "name", config.Name)
	require.Equal(t, int32(100), config.Number)

	m, err := NewModule(mc, binary)
	require.NoError(t, err)
	m.Start()

	// When a TriggerEvent occurs, Engine calls Run() with that Event.
	wrapped, err := anypb.New(testhelpers.TestWorkflowTrigger())
	require.NoError(t, err)

	// TODO test config
	req := &pb.ExecuteRequest{
		Id: anyNoDagExecId,
		Request: &pb.ExecuteRequest_Trigger{
			Trigger: &pb.Trigger{
				Id:      triggerID,
				Payload: wrapped,
			},
		},
	}
	response, err := m.Execute(ctx, req)
	require.NoError(t, err)

	require.Equal(t, anyNoDagExecId, response.Id)
	switch output := response.Result.(type) {
	case *pb.ExecutionResult_Value:
		valuePb := output.Value
		value, err := values.FromProto(valuePb)
		require.NoError(t, err)
		unwrapped, err := value.Unwrap()
		require.NoError(t, err)
		require.Equal(t, testhelpers.TestWorkflowExpectedResult(), unwrapped)
	case *pb.ExecutionResult_Error:
		require.Fail(t, "unexpected error response", output.Error)
	default:
		t.Fatalf("unexpected response type %T", output)
	}
}

func createNoDagMc(t *testing.T) *ModuleConfig {
	testhelpers.SetupExpectedCalls(t)
	registry := testutils.GetRegistry(t)
	lock := sync.Mutex{}
	responses := [2]*sdkpb.CapabilityResponse{}
	numRequests := 0
	numAwaits := 0

	mc := &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
		CallCapability: func(ctx context.Context, req *sdkpb.CapabilityRequest) ([sdk.IdLen]byte, error) {
			id := [sdk.IdLen]byte{}
			assert.Equal(t, anyNoDagExecId, req.ExecutionId)

			capability, err := registry.GetCapability(req.Id)
			if err != nil {
				return id, err
			}

			lock.Lock()
			defer lock.Unlock()

			responses[numRequests] = capability.Invoke(t.Context(), req)

			for i := 0; i < sdk.IdLen; i++ {
				id[i] = 'a' + byte(i+numRequests)%25
			}

			numRequests++

			return id, nil
		},
		AwaitCapabilities: func(ctx context.Context, req *pb.AwaitCapabilitiesRequest) (*pb.AwaitCapabilitiesResponse, error) {
			lock.Lock()
			defer lock.Unlock()

			assert.Equal(t, numRequests, numAwaits+1)
			require.Len(t, req.Ids, 1)
			assert.Equal(t, anyNoDagExecId, req.ExecId)

			for i := 0; i < sdk.IdLen; i++ {
				assert.Equal(t, 'a'+byte(numAwaits+i)%25, req.Ids[0][i])
			}

			resp := responses[numAwaits]

			numAwaits++
			return &pb.AwaitCapabilitiesResponse{
				Responses: map[string]*sdkpb.CapabilityResponse{req.Ids[0][:]: resp},
			}, nil
		},
	}
	return mc
}

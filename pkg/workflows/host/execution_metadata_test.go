package host

import (
	"bytes"
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

// Synthetic identifiers use production encodings; execution IDs follow the
// engine hash for the workflow ID, a cron event timestamp, and trigger index 0.
const (
	metadataWorkflowIDFixture       = "0020350fb4f89856a031268826d6e9b9340996637aa27859baa7b07db58bb020"
	metadataExecutionIDFixture      = "9f38bdece8892b3452c401a7a2d5a170f078ab9b2c98c591e24028ce94b33ed2"
	metadataOwnerFixture            = "5366890c26e8737b89d9084b4c632f7de73d1674"
	metadataOtherExecutionIDFixture = "e48ddcb9c9fdca45b68f7d37d1dc8bcb29df3777b3ece2f01c14d54872d44f01"
)

func TestRequirementSelectingModule_PreHookClonesExecutionMetadata(t *testing.T) {
	request := metadataTriggerRequest()
	original := proto.Clone(request).(*sdk.ExecuteRequest)

	received := executeMetadataPreHook(t, request)

	require.NotNil(t, received[0].GetPreHook())
	require.True(t, proto.Equal(original.ExecutionMetadata, received[0].ExecutionMetadata))
	require.Same(t, request, received[1])
	received[0].ExecutionMetadata.ExecutionId = metadataOtherExecutionIDFixture
	require.True(t, proto.Equal(original, request))
}

func TestRequirementSelectingModule_PreHookPreservesAbsentMetadata(t *testing.T) {
	received := executeMetadataPreHook(t, triggerRequest(0))

	require.NotNil(t, received[0].GetPreHook())
	require.Nil(t, received[0].ExecutionMetadata)
	require.Nil(t, received[1].ExecutionMetadata)
}

func TestRequirementSelectingModule_PreHookPreservesEmptyMetadata(t *testing.T) {
	request := triggerRequest(0)
	request.ExecutionMetadata = &sdk.WorkflowExecutionMetadata{}

	received := executeMetadataPreHook(t, request)

	require.NotNil(t, received[0].ExecutionMetadata)
	require.True(t, proto.Equal(request.ExecutionMetadata, received[0].ExecutionMetadata))
}

func TestRequirementSelectingModule_PreHookPreservesZeroFaultTolerance(t *testing.T) {
	request := metadataTriggerRequest()
	request.ExecutionMetadata.DonF = 0
	want := proto.Clone(request.ExecutionMetadata)

	received := executeMetadataPreHook(t, request)

	require.True(t, proto.Equal(want, received[0].ExecutionMetadata))
}

func TestRequirementSelectingModule_PreHookClonesUnknownMetadataFields(t *testing.T) {
	request := metadataTriggerRequest()
	unknown := protowire.AppendTag(nil, 100, protowire.VarintType)
	unknown = protowire.AppendVarint(unknown, 42)
	request.ExecutionMetadata.ProtoReflect().SetUnknown(unknown)
	want := bytes.Clone(unknown)

	received := executeMetadataPreHook(t, request)

	got := received[0].ExecutionMetadata.ProtoReflect().GetUnknown()
	require.Equal(t, want, []byte(got))
	got[len(got)-1]++
	require.Equal(t, want, []byte(request.ExecutionMetadata.ProtoReflect().GetUnknown()))
}

func metadataTriggerRequest() *sdk.ExecuteRequest {
	request := triggerRequest(0)
	request.ExecutionMetadata = &sdk.WorkflowExecutionMetadata{
		WorkflowId:    metadataWorkflowIDFixture,
		ExecutionId:   metadataExecutionIDFixture,
		WorkflowOwner: metadataOwnerFixture,
		DonId:         17,
		DonF:          3,
	}
	return request
}

func executeMetadataPreHook(t *testing.T, request *sdk.ExecuteRequest) []*sdk.ExecuteRequest {
	t.Helper()
	var received []*sdk.ExecuteRequest
	main := ModuleAndHandler{Module: &stubModule{
		executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
			received = append(received, req)
			return restrictionsResult(&sdk.Restrictions{}), nil
		},
	}}
	module := NewRequirementSelectingModule(main, nil)
	require.NoError(t, module.(TriggerCachePrimer).PrimeTriggerCache(t.Context(), []*sdk.TriggerSubscription{{PreHook: true}}))
	_, err := module.Execute(t.Context(), request, nil)
	require.NoError(t, err)
	require.Len(t, received, 2)
	return received
}

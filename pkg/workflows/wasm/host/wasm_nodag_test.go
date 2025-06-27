package host

import (
	_ "embed"
	"errors"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Test_NoDag_Run(t *testing.T) {
	t.Parallel()

	// Any of the test binaries that do subscription can be used here.
	m := makeTestModuleByName(t, "multiple_triggers")
	m.Start()
	defer m.Close()

	t.Run("NOK fails with unset ExecutionHelper for trigger", func(t *testing.T) {
		ctx := t.Context()
		req := &pb.ExecuteRequest{
			Request: &pb.ExecuteRequest_Trigger{},
		}

		_, err := m.Execute(ctx, req, nil)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid capability executor")
	})

	t.Run("OK can subscribe without setting ExecutionHelper", func(t *testing.T) {
		triggers, err := getTriggersSpec(t, m, []byte(""))
		require.NoError(t, err)
		require.Equal(t, len(triggers.Subscriptions), 3)
	})
}

func getTriggersSpec(t *testing.T, m ModuleV2, config []byte) (*pb.TriggerSubscriptionRequest, error) {
	helper := NewMockExecutionHelper(t)
	helper.EXPECT().GetWorkflowExecutionID().Return("Id")
	execResult, err := m.Execute(t.Context(), &pb.ExecuteRequest{
		Config:  config,
		Request: &pb.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}},
	}, helper)

	if err != nil {
		return nil, err
	}

	switch r := execResult.Result.(type) {
	case *pb.ExecutionResult_TriggerSubscriptions:
		return r.TriggerSubscriptions, nil
	case *pb.ExecutionResult_Error:
		return nil, errors.New(r.Error)
	default:
		return nil, errors.New("unexpected response from WASM binary: got nil spec response")
	}
}

package host

import (
	_ "embed"
	"errors"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	nodagBinaryCmd                  = "test/nodag/singlehandler/cmd"
	nodagBinaryLocation             = nodagBinaryCmd + "/testmodule.wasm"
	nodagMultiTriggerBinaryCmd      = "test/nodag/multihandler/cmd"
	nodagMultiTriggerBinaryLocation = nodagMultiTriggerBinaryCmd + "/testmodule.wasm"
	nodagRandomBinaryCmd            = "test/nodag/randoms/cmd"
	nodagRandomBinaryLocation       = nodagRandomBinaryCmd + "/testmodule.wasm"
)

var wordList = []string{"Hello, ", "world", "!"}

func Test_NoDag_Run(t *testing.T) {
	t.Parallel()

	binary := createTestBinary(nodagBinaryCmd, nodagBinaryLocation, true, t)

	t.Run("NOK fails with unset ExecutionHelper for trigger", func(t *testing.T) {
		mc := defaultNoDAGModCfg(t)
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		ctx := t.Context()
		req := &pb.ExecuteRequest{
			Request: &pb.ExecuteRequest_Trigger{},
		}

		_, err = m.Execute(ctx, req, nil)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid capability executor")
	})

	t.Run("OK can subscribe without setting ExecutionHelper", func(t *testing.T) {
		mc := defaultNoDAGModCfg(t)
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		triggers, err := getTriggersSpec(t, m, []byte(""))
		require.NoError(t, err)
		require.Equal(t, len(triggers.Subscriptions), 1)
	})
}

func defaultNoDAGModCfg(t testing.TB) *ModuleConfig {
	return &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
	}
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

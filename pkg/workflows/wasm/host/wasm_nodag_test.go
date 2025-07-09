package host

import (
	_ "embed"
	"errors"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/stretchr/testify/require"
)

const (
	nodagRandomBinaryCmd            = "standard_tests/multiple_triggers"
	nodagRandomBinaryLocation       = nodagRandomBinaryCmd + "/testmodule.wasm"
	nodagSleepTimeoutBinaryCmd      = "standard_tests/sleep_timeout"
	nodagSleepTimeoutBinaryLocation = nodagSleepTimeoutBinaryCmd + "/testmodule.wasm"
)

func Test_Sleep_Timeout(t *testing.T) {
	t.Parallel()

	binary := createTestBinary(nodagSleepTimeoutBinaryCmd, nodagSleepTimeoutBinaryLocation, true, t)

	mc := defaultNoDAGModCfg(t)
	timeout := 1 * time.Second
	mc.Timeout = &timeout
	m, err := NewModule(mc, binary)
	require.NoError(t, err)

	m.v2ImportName = "test"
	m.Start()
	defer m.Close()

	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	})

	req := &pb.ExecuteRequest{
		Request: &pb.ExecuteRequest_Trigger{},
	}

	start := time.Now()
	_, err = m.Execute(t.Context(), req, mockExecutionHelper)
	duration := time.Since(start)
	require.ErrorContains(t, err, "wasm trap: interrupt")
	require.Less(t, duration.Seconds(), 3.0, "execution should be interrupted quickly")
}

func Test_NoDag_Run(t *testing.T) {
	t.Parallel()

	binary := createTestBinary(nodagRandomBinaryCmd, nodagRandomBinaryLocation, true, t)

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
		require.Equal(t, len(triggers.Subscriptions), 3)
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
	helper.EXPECT().GetNodeTime().Return(time.Now()).Maybe()
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

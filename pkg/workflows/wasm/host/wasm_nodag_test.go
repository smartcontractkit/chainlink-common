package host

import (
	"context"
	_ "embed"
	"errors"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/engine"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testRuntimeOverride struct {
	engine.Runtime
	v2ImportNameOverride string
}

func (t *testRuntimeOverride) V2ImportName() string {
	return t.v2ImportNameOverride
}

const (
	nodagRandomBinaryCmd        = "standard_tests/multiple_triggers"
	nodagRandomBinaryLocation   = nodagRandomBinaryCmd + "/testmodule.wasm"
	loggingLimitsBinaryCmd      = "test/logging_limits/cmd"
	loggingLimitsBinaryLocation = loggingLimitsBinaryCmd + "/testmodule.wasm"
)

func Test_Sleep_Timeout(t *testing.T) {
	t.Parallel()

	binary := createTestBinary(sleepBinaryCmd, sleepBinaryLocation, true, t)

	mc := defaultNoDAGModCfg(t)
	timeout := 1 * time.Second
	mc.Timeout = &timeout
	m, err := NewModule(t.Context(), mc, binary)
	require.NoError(t, err)

	m.runtime = &testRuntimeOverride{Runtime: m.runtime, v2ImportNameOverride: "test"}
	m.Start()
	defer m.Close()

	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	})

	req := &sdk.ExecuteRequest{
		Request: &sdk.ExecuteRequest_Trigger{},
	}

	start := time.Now()
	_, err = m.Execute(t.Context(), req, mockExecutionHelper)
	duration := time.Since(start)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, duration.Seconds(), 3.0, "execution should be interrupted quickly")
}

func Test_NoDag_Run(t *testing.T) {
	t.Parallel()

	binary := createTestBinary(nodagRandomBinaryCmd, nodagRandomBinaryLocation, true, t)

	t.Run("NOK fails with unset ExecutionHelper for trigger", func(t *testing.T) {
		mc := defaultNoDAGModCfg(t)
		m, err := NewModule(t.Context(), mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		ctx := t.Context()
		req := &sdk.ExecuteRequest{
			Request: &sdk.ExecuteRequest_Trigger{},
		}

		_, err = m.Execute(ctx, req, nil)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid capability executor")
	})

	t.Run("OK can subscribe without setting ExecutionHelper", func(t *testing.T) {
		mc := defaultNoDAGModCfg(t)
		m, err := NewModule(t.Context(), mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		triggers, err := getTriggersSpec(t, m, []byte(""))
		require.NoError(t, err)
		require.Equal(t, len(triggers.Subscriptions), 3)
	})
}

func Test_NoDAG_LoggingWithLimits(t *testing.T) {
	t.Parallel()
	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()
	mockExecutionHelper.EXPECT().GetDONTime().RunAndReturn(func() (time.Time, error) {
		return time.Now(), nil
	}).Maybe()

	logs := []string{}
	mockExecutionHelper.EXPECT().EmitUserLog(mock.Anything).RunAndReturn(func(s string) error {
		logs = append(logs, s)
		return nil
	})

	trigger := &basictrigger.Outputs{CoolOutput: anyTestTriggerValue}
	executeRequest := triggerExecuteRequest(t, 0, trigger)
	cfg := &ModuleConfig{
		Logger:              logger.Test(t),
		IsUncompressed:      true,
		MaxLogLenBytes:      20,
		MaxLogCountDONMode:  3,
		MaxLogCountNodeMode: 3,
	}

	binary := createTestBinary(loggingLimitsBinaryCmd, loggingLimitsBinaryLocation, true, t)

	m, err := NewModule(t.Context(), cfg, binary)
	require.NoError(t, err)

	_, err = m.Execute(t.Context(), executeRequest, mockExecutionHelper)
	require.NoError(t, err)

	// allowed 3 logs max, one of which got rejected because it was too long
	// so expect 2 logs to be emitted
	require.Equal(t, 2, len(logs))
	require.Equal(t, "short log 1", logs[0])
	require.Equal(t, "short log 3", logs[1])
}

func defaultNoDAGModCfg(t testing.TB) *ModuleConfig {
	return &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
	}
}

func getTriggersSpec(t *testing.T, m ModuleV2, config []byte) (*sdk.TriggerSubscriptionRequest, error) {
	helper := NewMockExecutionHelper(t)
	helper.EXPECT().GetWorkflowExecutionID().Return("Id")
	helper.EXPECT().GetNodeTime().Return(time.Now()).Maybe()
	execResult, err := m.Execute(t.Context(), &sdk.ExecuteRequest{
		Config:  config,
		Request: &sdk.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}},
	}, helper)

	if err != nil {
		return nil, err
	}

	switch r := execResult.Result.(type) {
	case *sdk.ExecutionResult_TriggerSubscriptions:
		return r.TriggerSubscriptions, nil
	case *sdk.ExecutionResult_Error:
		return nil, errors.New(r.Error)
	default:
		return nil, errors.New("unexpected response from WASM binary: got nil spec response")
	}
}

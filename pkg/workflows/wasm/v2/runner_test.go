package wasm

import (
	"encoding/base64"
	"testing"

	basictriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger/basic_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/testhelpers"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

const anyExecutionId = "execId"

var anyConfig = []byte("config")
var anyMaxResponseSize = uint64(2048)

var triggerId = (&basictriggermock.BasicCapability{}).ID()

var subscribeRequest = &pb.ExecuteRequest{
	Id:              anyExecutionId,
	Config:          anyConfig,
	MaxResponseSize: anyMaxResponseSize,
	Request:         &pb.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}},
}

var anyExecuteRequest = &pb.ExecuteRequest{
	Id:              anyExecutionId,
	Config:          anyConfig,
	MaxResponseSize: anyMaxResponseSize,
	Request: &pb.ExecuteRequest_Trigger{
		Trigger: &sdkpb.Trigger{
			Id:      triggerId,
			Payload: mustAny(testhelpers.TestWorkflowTrigger()),
		},
	},
}

func TestRunner_Config(t *testing.T) {
	dr := getTestDonRunner(t, anyExecuteRequest)
	assert.Equal(t, string(anyConfig), string(dr.Config()))
	dr = getTestDonRunner(t, subscribeRequest)
	assert.Equal(t, string(anyConfig), string(dr.Config()))
	nr := getTestNodeRunner(t, anyExecuteRequest)
	assert.Equal(t, string(anyConfig), string(nr.Config()))
	nr = getTestNodeRunner(t, subscribeRequest)
	assert.Equal(t, string(anyConfig), string(nr.Config()))
}

func TestRunner_LogWriter(t *testing.T) {
	dr := getTestDonRunner(t, anyExecuteRequest)
	assert.IsType(t, &writer{}, dr.LogWriter())
	dr = getTestDonRunner(t, subscribeRequest)
	assert.IsType(t, &writer{}, dr.LogWriter())
	nr := getTestNodeRunner(t, anyExecuteRequest)
	assert.IsType(t, &writer{}, nr.LogWriter())
	nr = getTestNodeRunner(t, subscribeRequest)
	assert.IsType(t, &writer{}, nr.LogWriter())
}

func TestRunner_Run(t *testing.T) {
	t.Run("runner gathers subscriptions", func(t *testing.T) {
		dr := getTestDonRunner(t, subscribeRequest)
		dr.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
			Handlers: []sdk.Handler[sdk.DonRuntime]{
				sdk.NewDonHandler(
					basictrigger.Basic{}.Trigger(testhelpers.TestWorkflowTriggerConfig()),
					func(_ sdk.DonRuntime, _ *basictrigger.Outputs) (int, error) {
						require.Fail(t, "Must not be called during registration to tiggers")
						return 0, nil
					}),
			},
		})

		response := &pb.ExecutionResult{}
		require.NoError(t, proto.Unmarshal(sentResponse, response))

		actual := &pb.ExecutionResult{}
		require.NoError(t, proto.Unmarshal(sentResponse, actual))
		assert.Equal(t, anyExecutionId, actual.Id)

		switch result := actual.Result.(type) {
		case *pb.ExecutionResult_TriggerSubscriptions:
			subscriptions := result.TriggerSubscriptions.Subscriptions
			require.Len(t, subscriptions, 1)
			subscription := subscriptions[0]
			assert.Equal(t, anyExecutionId, subscription.ExecId)
			assert.Equal(t, triggerId, subscription.Id)
			assert.Equal(t, "Trigger", subscription.Method)
			payload := &basictrigger.Config{}
			require.NoError(t, subscription.Payload.UnmarshalTo(payload))
			assert.True(t, proto.Equal(testhelpers.TestWorkflowTriggerConfig(), payload))
		default:
			assert.Fail(t, "unexpected result type", result)
		}
	})

	t.Run("makes callback with correct runner", func(t *testing.T) {
		testhelpers.SetupExpectedCalls(t)
		dr := getTestDonRunner(t, anyExecuteRequest)
		testhelpers.RunTestWorkflow(dr)

		actual := &pb.ExecutionResult{}
		require.NoError(t, proto.Unmarshal(sentResponse, actual))
		assert.Equal(t, anyExecutionId, actual.Id)

		switch result := actual.Result.(type) {
		case *pb.ExecutionResult_Value:
			v, err := values.FromProto(result.Value)
			require.NoError(t, err)
			returnedValue, err := v.Unwrap()
			require.NoError(t, err)
			assert.Equal(t, testhelpers.TestWorkflowExpectedResult(), returnedValue)
		default:
			assert.Fail(t, "unexpected result type", result)
		}
	})
}

func getTestDonRunner(tb testing.TB, request *pb.ExecuteRequest) sdk.DonRunner {
	initTestRunner(tb, request)
	return newDonRunner()
}

func getTestNodeRunner(tb testing.TB, request *pb.ExecuteRequest) sdk.NodeRunner {
	initTestRunner(tb, request)
	return newNodeRunner()
}

func initTestRunner(tb testing.TB, request *pb.ExecuteRequest) {
	initRunnerAndRuntimeForTest(tb, anyExecutionId)
	serialzied, err := proto.Marshal(request)
	require.NoError(tb, err)
	encoded := base64.StdEncoding.EncodeToString(serialzied)
	args = []string{"wasm", encoded}
}

func mustAny(msg proto.Message) *anypb.Any {
	a, err := anypb.New(msg)
	if err != nil {
		panic(err)
	}
	return a
}

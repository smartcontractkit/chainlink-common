package host

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/iancoleman/strcase"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/actionandtrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// See the README.md in standard_tests for more information.

var anyTestConfig = []byte("config")
var anyTestTriggerValue = "test value"

var testPath string

func init() {
	flag.StringVar(&testPath, "path", "./standard_tests", "Path to the standard tests")
}

func TestStandardConfig(t *testing.T) {
	t.Parallel()
	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	wrappedConfig := runWithBasicTrigger(t, mockExecutionHelper)
	actualConfig := wrappedConfig.GetValue().GetBytesValue()
	require.ElementsMatch(t, anyTestConfig, actualConfig)
}

func TestStandardErrors(t *testing.T) {
	t.Parallel()
	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	errMsg := runWithBasicTrigger(t, mockExecutionHelper)
	assert.Contains(t, errMsg.GetError(), "workflow execution failure")
}

func TestStandardCapabilityCallsAreAsync(t *testing.T) {
	t.Parallel()
	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	m := makeTestModule(t)
	request := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})
	callsSeen := map[bool]bool{}
	mt := sync.Mutex{}
	mt.Lock()
	mockExecutionHelper.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, request *pb.CapabilityRequest) (*pb.CapabilityResponse, error) {
		assert.Equal(t, "basic-test-action@1.0.0", request.Id)
		assert.Equal(t, "PerformAction", request.Method)
		input := &basicaction.Inputs{}
		assert.NoError(t, request.Payload.UnmarshalTo(input))
		assert.False(t, callsSeen[input.InputThing])
		callsSeen[input.InputThing] = true
		payload, err := anypb.New(&basicaction.Outputs{AdaptedThing: fmt.Sprintf("%t", input.InputThing)})
		require.NoError(t, err)

		// Don't return until the second call has been executed
		defer func() {
			if !input.InputThing {
				mt.Lock()
			}
			defer mt.Unlock()
		}()
		return &pb.CapabilityResponse{
			Response: &pb.CapabilityResponse_Payload{Payload: payload},
		}, nil
	})
	result, err := m.Execute(t.Context(), request, mockExecutionHelper)
	require.NoError(t, err)

	assert.Equal(t, "truefalse", result.GetValue().GetStringValue())
}

func TestStandardModeSwitch(t *testing.T) {
	t.Parallel()
	t.Run("successful mode switch", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
		mockExecutionHelper.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *pb.CapabilityRequest) (*pb.CapabilityResponse, error) {
			if request.Id == "basic-test-action@1.0.0" {
				input := &basicaction.Inputs{}
				assert.NoError(t, request.Payload.UnmarshalTo(input))
				assert.True(t, input.InputThing)
				payload, err := anypb.New(&basicaction.Outputs{AdaptedThing: fmt.Sprintf("test")})
				require.NoError(t, err)
				return &pb.CapabilityResponse{
					Response: &pb.CapabilityResponse_Payload{Payload: payload},
				}, nil
			}
			return setupNodeCallAndConsensusCall(t, 555)(ctx, request)
		})

		m := makeTestModule(t)
		request := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})
		result, err := m.Execute(t.Context(), request, mockExecutionHelper)
		require.NoError(t, err)
		require.Equal(t, "test555", result.GetValue().GetStringValue())
	})

	t.Run("node runtime in don mode", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
		m := makeTestModule(t)
		request := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})
		result, err := m.Execute(t.Context(), request, mockExecutionHelper)
		require.NoError(t, err)
		require.Contains(t, result.GetError(), "cannot use NodeRuntime outside RunInNodeMode")
	})

	t.Run("don runtime in node mode", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
		m := makeTestModule(t)
		request := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})
		result, err := m.Execute(t.Context(), request, mockExecutionHelper)
		require.NoError(t, err)
		require.Contains(t, result.GetError(), "cannot use Runtime inside RunInNodeMode")
	})
}

func TestStandardLogging(t *testing.T) {
	t.Parallel()
	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	mockExecutionHelper.EXPECT().EmitUserLog(mock.Anything).RunAndReturn(func(s string) error {
		assert.True(t, strings.Contains(s, "log from wasm!"))
		return nil
	}).Once()

	runWithBasicTrigger(t, mockExecutionHelper)
}

func TestStandardMultipleTriggers(t *testing.T) {
	t.Parallel()
	m := makeTestModule(t)
	t.Run("test registration", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")

		subscribe := &pb.ExecuteRequest{Request: &pb.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}}}
		actual, err := m.Execute(t.Context(), subscribe, mockExecutionHelper)
		require.NoError(t, err)

		payload0, err := anypb.New(&basictrigger.Config{
			Name:   "first-trigger",
			Number: 100,
		})
		require.NoError(t, err)

		payload1, err := anypb.New(&actionandtrigger.Config{
			Name:   "second-trigger",
			Number: 150,
		})
		require.NoError(t, err)

		payload2, err := anypb.New(&basictrigger.Config{
			Name:   "third-trigger",
			Number: 200,
		})
		require.NoError(t, err)

		expected := &pb.TriggerSubscriptionRequest{
			Subscriptions: []*pb.TriggerSubscription{
				{
					Id:      "basic-trigger@1.0.0",
					Payload: payload0,
					Method:  "Trigger",
				},
				{
					Id:      "basic-test-action-trigger@1.0.0",
					Payload: payload1,
					Method:  "Trigger",
				},
				{
					Id:      "basic-trigger@1.0.0",
					Payload: payload2,
					Method:  "Trigger",
				},
			},
		}

		assertProto(t, expected, actual.GetTriggerSubscriptions())
	})

	t.Run("first callback", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")

		request := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})
		result, err := m.Execute(t.Context(), request, mockExecutionHelper)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("called 0 with %v", anyTestTriggerValue), result.GetValue().GetStringValue())
	})

	t.Run("same trigger as first one but different registration", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")

		request := triggerExecuteRequest(t, 2, &basictrigger.Outputs{CoolOutput: "different"})
		result, err := m.Execute(t.Context(), request, mockExecutionHelper)
		require.NoError(t, err)

		require.Equal(t, "called 2 with different", result.GetValue().GetStringValue())
	})

	t.Run("different capability callback", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")

		request := triggerExecuteRequest(t, 1, &actionandtrigger.TriggerEvent{CoolOutput: "different"})
		result, err := m.Execute(t.Context(), request, mockExecutionHelper)
		require.NoError(t, err)

		require.Equal(t, "called 1 with different", result.GetValue().GetStringValue())
	})
}

func TestStandardRandom(t *testing.T) {
	t.Parallel()
	m := makeTestModule(t)

	// Test binary executes node mode code conditionally based on the value >= 100
	anyId := "Id"
	gte100Exec := NewMockExecutionHelper(t)
	gte100Exec.EXPECT().GetWorkflowExecutionID().Return(anyId)

	// RunAndReturn
	gte100Exec.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(setupNodeCallAndConsensusCall(t, 150))

	m.Start()
	defer m.Close()

	trigger := &basictrigger.Outputs{CoolOutput: "trigger1"}
	triggerPayload, err := anypb.New(trigger)
	require.NoError(t, err)
	anyRequest := &pb.ExecuteRequest{
		Request: &pb.ExecuteRequest_Trigger{
			Trigger: &pb.Trigger{
				Id:      uint64(0),
				Payload: triggerPayload,
			},
		},
	}
	execution1Result, err := m.Execute(t.Context(), anyRequest, gte100Exec)
	require.NoError(t, err)
	wrappedValue1, err := values.FromProto(execution1Result.GetValue())
	require.NoError(t, err)
	value1, err := wrappedValue1.Unwrap()
	require.NoError(t, err)

	t.Run("Same execution id gives the same randoms even if random is called in node mode", func(t *testing.T) {
		lt100Exec := NewMockExecutionHelper(t)
		lt100Exec.EXPECT().GetWorkflowExecutionID().Return(anyId)

		lt100Exec.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(setupNodeCallAndConsensusCall(t, 99))
		lt100Exec.EXPECT().EmitUserLog(mock.Anything).RunAndReturn(func(s string) error {
			_, err = strconv.ParseUint(s, 10, 64)
			require.NoError(t, err)
			return nil
		}).Once()

		execution2Result, err := m.Execute(t.Context(), anyRequest, lt100Exec)
		require.NoError(t, err)
		wrappedValue2, err := values.FromProto(execution2Result.GetValue())
		require.NoError(t, err)
		value2, err := wrappedValue2.Unwrap()
		require.NoError(t, err)
		require.Equal(t, value1, value2, "Expected the same random number to be generated for the same trigger")
	})

	t.Run("Different execution id give different randoms", func(t *testing.T) {
		require.NoError(t, err)

		gte100Exec2 := NewMockExecutionHelper(t)
		gte100Exec2.EXPECT().GetWorkflowExecutionID().Return("differentId")

		gte100Exec2.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(setupNodeCallAndConsensusCall(t, 120))

		executionResult2, err := m.Execute(t.Context(), anyRequest, gte100Exec2)
		require.NoError(t, err)
		wrappedValue2, err := values.FromProto(executionResult2.GetValue())
		require.NoError(t, err)
		value2, err := wrappedValue2.Unwrap()
		require.NoError(t, err)
		require.NotEqual(t, value1, value2, "Expected different random numbers for different triggers")
	})
}

func TestStandardSecrets(t *testing.T) {
	t.Parallel()

	m := makeTestModule(t)

	t.Run("returns the secret value", func(t *testing.T) {
		result := runSecretTest(t, m, &pb.SecretResponse{
			Response: &pb.SecretResponse_Secret{
				Secret: &pb.Secret{
					Value: "Bar",
				},
			},
		})
		require.Equal(t, "Bar", result.GetValue().GetStringValue())
	})

	t.Run("returns an error if the secret doesn't exist", func(t *testing.T) {
		resp := runSecretTest(t, m, &pb.SecretResponse{
			Response: &pb.SecretResponse_Error{
				Error: "could not find secret",
			},
		})
		assert.ErrorContains(t, errors.New(resp.GetError()), "could not find secret")
	})
}

func triggerExecuteRequest(t *testing.T, id uint64, trigger proto.Message) *pb.ExecuteRequest {
	wrappedTrigger, err := anypb.New(trigger)
	require.NoError(t, err)
	return &pb.ExecuteRequest{
		Config: anyTestConfig,
		Request: &pb.ExecuteRequest_Trigger{
			Trigger: &pb.Trigger{Id: id, Payload: wrappedTrigger},
		},
		MaxResponseSize: uint64(defaultMaxResponseSizeBytes),
	}
}

func runWithBasicTrigger(t *testing.T, executor ExecutionHelper) *pb.ExecutionResult {
	trigger := &basictrigger.Outputs{CoolOutput: anyTestTriggerValue}
	executeRequest := triggerExecuteRequest(t, 0, trigger)
	m := makeTestModule(t)
	response, err := m.Execute(t.Context(), executeRequest, executor)
	require.NoError(t, err)
	return response
}

// makeTestModule compiles the test module from the Makefile in the testPath directory
// The test to compile and run is determined by the test name.
// To re-use a binary, an outer test can create the module and use t.Run to run subtests using that module.
// When subtests have their own binaries, those binaries are expected to be nested in a subfolder.
func makeTestModule(t *testing.T) *module {
	testName := strcase.ToSnake(t.Name()[len("TestStandard"):])
	return makeTestModuleByName(t, testName)
}

func makeTestModuleByName(t *testing.T, testName string) *module {
	wasmName := path.Join(testName, "test.wasm")
	cmd := exec.Command("make", wasmName) // #nosec
	absPath, err := filepath.Abs(testPath)
	require.NoError(t, err, "Failed to get absolute path for test directory")
	cmd.Dir = absPath

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	binary, err := os.ReadFile(filepath.Join(absPath, wasmName))
	require.NoError(t, err)

	cfg := defaultNoDAGModCfg(t)
	mod, err := NewModule(cfg, binary)
	require.NoError(t, err)
	return mod
}

func defaultNoDAGModCfg(t testing.TB) *ModuleConfig {
	return &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
	}
}

func setupNodeCallAndConsensusCall(t *testing.T, output int32) func(_ context.Context, request *pb.CapabilityRequest) (*pb.CapabilityResponse, error) {
	return func(_ context.Context, request *pb.CapabilityRequest) (*pb.CapabilityResponse, error) {
		nodeResponse := &nodeaction.NodeOutputs{OutputThing: output}
		var err error
		var payload *anypb.Any
		switch request.Id {
		case "basic-test-node-action@1.0.0":
			input := &nodeaction.NodeInputs{}
			assert.NoError(t, request.Payload.UnmarshalTo(input))
			assert.True(t, input.InputThing)
			payload, err = anypb.New(nodeResponse)
			if err != nil {
				require.Fail(t, err.Error())
			}
		case "consensus@1.0.0":
			input := &pb.SimpleConsensusInputs{}
			require.NoError(t, request.Payload.UnmarshalTo(input))

			expectedObservation := wrapValue(t, nodeResponse)
			expectedInput := &pb.SimpleConsensusInputs{
				Observation: &pb.SimpleConsensusInputs_Value{Value: expectedObservation},
				Descriptors: &pb.ConsensusDescriptor{
					Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
						FieldsMap: &pb.FieldsMap{
							Fields: map[string]*pb.ConsensusDescriptor{
								"OutputThing": {
									Descriptor_: &pb.ConsensusDescriptor_Aggregation{
										Aggregation: pb.AggregationType_AGGREGATION_TYPE_IDENTICAL,
									},
								},
							},
						},
					},
				},
				Default: wrapValue(t, &nodeaction.NodeOutputs{OutputThing: 123}),
			}
			assertProto(t, expectedInput, input)
			payload, err = anypb.New(expectedObservation)
			require.NoError(t, err)
		default:
			err = fmt.Errorf("unexpected capability: %s", request.Id)
			assert.Fail(t, err.Error())
			return nil, err
		}

		return &pb.CapabilityResponse{
			Response: &pb.CapabilityResponse_Payload{
				Payload: payload,
			},
		}, nil
	}
}

func wrapValue(t *testing.T, nodeResponse *nodeaction.NodeOutputs) *valuespb.Value {
	valueInput, err := values.Wrap(nodeResponse)
	require.NoError(t, err)
	return values.Proto(valueInput)
}

func assertProto[T proto.Message](t *testing.T, expected, actual T) {
	t.Helper()
	diff := cmp.Diff(expected, actual, protocmp.Transform())

	var sb strings.Builder
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			sb.WriteString(line + "\n")
		}
	}
	assert.Empty(t, sb.String())
}

func runSecretTest(t *testing.T, m *module, secretResponse *pb.SecretResponse) *pb.ExecutionResult {
	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("Id")

	mockExecutionHelper.EXPECT().GetSecrets(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, request *pb.GetSecretsRequest) ([]*pb.SecretResponse, error) {
			assert.Len(t, request.Requests, 1)
			assert.Equal(t, "Foo", request.Requests[0].Id)
			return []*pb.SecretResponse{secretResponse}, nil
		}).
		Once()

	trigger := &basictrigger.Outputs{CoolOutput: anyTestTriggerValue}
	executeRequest := triggerExecuteRequest(t, 0, trigger)
	response, err := m.Execute(t.Context(), executeRequest, mockExecutionHelper)
	require.NoError(t, err)
	return response
}

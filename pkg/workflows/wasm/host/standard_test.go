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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/actionandtrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
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
	// Some languages call time during initiation of the executable before the main is called.
	// This would be in unknown mode, which would call Node mode by default.
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()
	wrappedConfig := runWithBasicTrigger(t, mockExecutionHelper)
	wrappedValue := wrappedConfig.GetValue()
	require.NotNil(t, wrappedValue, "Expected a value in the response")
	actualConfig := wrappedConfig.GetValue().GetBytesValue()
	require.ElementsMatch(t, anyTestConfig, actualConfig)
}

func TestStandardErrors(t *testing.T) {
	t.Parallel()
	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()
	errMsg := runWithBasicTrigger(t, mockExecutionHelper)
	assert.Contains(t, errMsg.GetError(), "workflow execution failure")
}

func TestStandardCapabilityCallsAreAsync(t *testing.T) {
	t.Parallel()
	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()
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

	result := executeWithResult[string](t, m, request, mockExecutionHelper)

	assert.Equal(t, "truefalse", result)
}

func TestStandardModeSwitch(t *testing.T) {
	t.Parallel()
	t.Run("successful mode switch", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
		// Node calls may occur on initialization depending on the language.
		var donCall1 bool
		var nodeCall bool
		var donCall2 bool
		mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			if donCall1 {
				nodeCall = true
			}
			return time.Now()
		})
		// We want to make sure time.Sleep() is called at least twice in DON mode and once in node Mode
		mockExecutionHelper.EXPECT().GetDONTime(mock.Anything).RunAndReturn(func(ctx context.Context) (time.Time, error) {
			donCall1 = true
			if nodeCall {
				donCall2 = true
			}
			return time.Now(), nil
		})
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
		result := executeWithResult[string](t, m, request, mockExecutionHelper)
		require.Equal(t, "test556", result)
		require.True(t, donCall1)
		require.True(t, nodeCall)
		require.True(t, donCall2)
	})

	t.Run("node runtime in don mode", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
		mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			return time.Now()
		}).Maybe()
		mockExecutionHelper.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, request *pb.CapabilityRequest) (*pb.CapabilityResponse, error) {
			response := values.NewString("hi")
			mapProto := &valuespb.Map{
				Fields: map[string]*valuespb.Value{
					rawsdk.ConsensusResponseMapKeyMetadata: {Value: &valuespb.Value_StringValue{StringValue: "test_metadata"}},
					rawsdk.ConsensusResponseMapKeyPayload:  values.Proto(response),
				},
			}
			payload, err := anypb.New(mapProto)
			require.NoError(t, err)
			return &pb.CapabilityResponse{
				Response: &pb.CapabilityResponse_Payload{
					Payload: payload,
				},
			}, nil
		}).Once()
		m := makeTestModule(t)
		request := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})
		errStr := executeWithError(t, m, request, mockExecutionHelper)
		require.Contains(t, errStr, "cannot use NodeRuntime outside RunInNodeMode")
	})

	t.Run("don runtime in node mode", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
		mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			return time.Now()
		}).Maybe()
		mockExecutionHelper.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, request *pb.CapabilityRequest) (*pb.CapabilityResponse, error) {
			assert.Equal(t, "consensus@1.0.0-alpha", request.Id)
			input := &pb.SimpleConsensusInputs{}
			require.NoError(t, request.Payload.UnmarshalTo(input))

			var errMsg string
			switch msg := input.Observation.(type) {
			case *pb.SimpleConsensusInputs_Error:
				errMsg = msg.Error
			default:
				require.Fail(t, "observation must be an error")
			}
			return &pb.CapabilityResponse{
				Response: &pb.CapabilityResponse_Error{Error: errMsg},
			}, nil
		}).Once()
		m := makeTestModule(t)
		request := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})

		errStr := executeWithError(t, m, request, mockExecutionHelper)

		require.Contains(t, errStr, "cannot use Runtime inside RunInNodeMode")
	})
}

func TestStandardLogging(t *testing.T) {
	t.Parallel()
	mockExecutionHelper := NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()
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
		mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			return time.Now()
		}).Maybe()

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
					Id:      "basic-test-trigger@1.0.0",
					Payload: payload0,
					Method:  "Trigger",
				},
				{
					Id:      "basic-test-action-trigger@1.0.0",
					Payload: payload1,
					Method:  "Trigger",
				},
				{
					Id:      "basic-test-trigger@1.0.0",
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
		mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			return time.Now()
		}).Maybe()

		request := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})

		result := executeWithResult[string](t, m, request, mockExecutionHelper)

		require.Equal(t, fmt.Sprintf("called 0 with %v", anyTestTriggerValue), result)
	})

	t.Run("same trigger as first one but different registration", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
		mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			return time.Now()
		}).Maybe()

		request := triggerExecuteRequest(t, 2, &basictrigger.Outputs{CoolOutput: "different"})
		result := executeWithResult[string](t, m, request, mockExecutionHelper)

		require.Equal(t, "called 2 with different", result)
	})

	t.Run("different capability callback", func(t *testing.T) {
		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
		mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			return time.Now()
		}).Maybe()

		request := triggerExecuteRequest(t, 1, &actionandtrigger.TriggerEvent{CoolOutput: "different"})
		result := executeWithResult[string](t, m, request, mockExecutionHelper)

		require.Equal(t, "called 1 with different", result)
	})
}

func TestStandardRandom(t *testing.T) {
	t.Parallel()
	m := makeTestModule(t)

	// Test binary executes node mode code conditionally based on the value >= 100
	anyId := "Id"
	gte100Exec := NewMockExecutionHelper(t)
	gte100Exec.EXPECT().GetWorkflowExecutionID().Return(anyId)
	gte100Exec.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()

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

	// any since uint64 can be int64 or *big.Int
	value1 := executeWithResult[any](t, m, anyRequest, gte100Exec)

	t.Run("Same execution id gives the same randoms even if random is called in node mode", func(t *testing.T) {
		lt100Exec := NewMockExecutionHelper(t)
		lt100Exec.EXPECT().GetWorkflowExecutionID().Return(anyId)
		lt100Exec.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			return time.Now()
		}).Maybe()

		lt100Exec.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(setupNodeCallAndConsensusCall(t, 99))
		lt100Exec.EXPECT().EmitUserLog(mock.Anything).RunAndReturn(func(s string) error {
			parts := strings.Split(s, "***")
			_, err = strconv.ParseUint(parts[1], 10, 64)
			require.NoError(t, err)
			return nil
		}).Once()

		value2 := executeWithResult[any](t, m, anyRequest, lt100Exec)
		require.Equal(t, value1, value2, "Expected the same random number to be generated for the same trigger")
	})

	t.Run("Different execution id give different randoms", func(t *testing.T) {
		require.NoError(t, err)

		gte100Exec2 := NewMockExecutionHelper(t)
		gte100Exec2.EXPECT().GetWorkflowExecutionID().Return("differentId")
		gte100Exec2.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
			return time.Now()
		}).Maybe()

		gte100Exec2.EXPECT().CallCapability(mock.Anything, mock.Anything).RunAndReturn(setupNodeCallAndConsensusCall(t, 120))

		value2 := executeWithResult[any](t, m, anyRequest, gte100Exec2)

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
				Error: &pb.SecretError{
					Error: "could not find secret",
				},
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
	return makeTestModuleByName(t, testName, nil)
}

func makeTestModuleWithCfg(t *testing.T, cfg *ModuleConfig) *module {
	testName := strcase.ToSnake(t.Name()[len("TestStandard"):])
	return makeTestModuleByName(t, testName, cfg)
}

func makeTestModuleByName(t *testing.T, testName string, cfg *ModuleConfig) *module {
	wasmName := path.Join(testName, "test.wasm")
	cmd := exec.Command("make", wasmName) // #nosec
	absPath, err := filepath.Abs(testPath)
	require.NoError(t, err, "Failed to get absolute path for test directory")
	cmd.Dir = absPath

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	binary, err := os.ReadFile(filepath.Join(absPath, wasmName))
	require.NoError(t, err)

	if cfg == nil {
		cfg = defaultNoDAGModCfg(t)
	}
	mod, err := NewModule(cfg, binary)
	require.NoError(t, err)
	return mod
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
		case "consensus@1.0.0-alpha":
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
										Aggregation: pb.AggregationType_AGGREGATION_TYPE_MEDIAN,
									},
								},
							},
						},
					},
				},
				Default: wrapValue(t, &nodeaction.NodeOutputs{OutputThing: 123}),
			}
			assertProto(t, expectedInput, input)
			cResponse := &nodeaction.NodeOutputs{OutputThing: output + 1}
			response := wrapValue(t, cResponse)
			mapProto := &valuespb.Map{
				Fields: map[string]*valuespb.Value{
					rawsdk.ConsensusResponseMapKeyMetadata: {Value: &valuespb.Value_StringValue{StringValue: "test_metadata"}},
					rawsdk.ConsensusResponseMapKeyPayload:  response,
				},
			}
			payload, err = anypb.New(mapProto)
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
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()

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

func executeWithResult[T any](t *testing.T, m *module, req *pb.ExecuteRequest, executor ExecutionHelper) T {
	res, err := m.Execute(t.Context(), req, executor)
	require.NoError(t, err)
	var result T
	switch v := res.Result.(type) {
	case *pb.ExecutionResult_Value:
		wrappedValue, err := values.FromProto(v.Value)
		require.NoError(t, err)
		require.NoError(t, wrappedValue.UnwrapTo(&result))
	case *pb.ExecutionResult_Error:
		require.Failf(t, "unexpected error in result", "error: %s", v.Error)
	default:
		require.Failf(t, "unexpected result type", "result: %v", res)
	}

	return result
}

func executeWithError(t *testing.T, m *module, req *pb.ExecuteRequest, executor ExecutionHelper) string {
	res, err := m.Execute(t.Context(), req, executor)
	require.NoError(t, err)
	switch e := res.Result.(type) {
	case *pb.ExecutionResult_Error:
		return e.Error
	default:
		require.Failf(t, "unexpected result type", "%T", e)
		return ""
	}
}

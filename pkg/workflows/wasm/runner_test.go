package wasm

import (
	"encoding/base64"
	"encoding/binary"
	"math/big"
	"testing"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictarget"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func Test_Runner_Config_InvalidRequest(t *testing.T) {
	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	runner := &Runner{
		sendResponse: responseFn,
		args:         []string{"wasm", "bla"},
	}
	c := runner.Config()
	assert.Nil(t, c)
	assert.Equal(t, unknownID, gotResponse.Id)
	assert.Contains(t, gotResponse.ErrMsg, "could not decode request")
}

func Test_Runner_Config_InvalidRequest_NotEnoughArgs(t *testing.T) {
	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	runner := &Runner{
		sendResponse: responseFn,
		args:         []string{"wasm"},
	}
	c := runner.Config()
	assert.Nil(t, c)
	assert.Equal(t, unknownID, gotResponse.Id)
	assert.Contains(t, gotResponse.ErrMsg, "request must contain a payload")
}

func marshalRequest(req *wasmpb.Request) (string, error) {
	rpb, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(rpb), nil
}

func Test_Runner_Config(t *testing.T) {
	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	cfg := []byte(`{"hello": "world"}`)
	request := &wasmpb.Request{
		Id:     uuid.New().String(),
		Config: cfg,
	}
	str, err := marshalRequest(request)
	require.NoError(t, err)
	runner := &Runner{
		sendResponse: responseFn,
		args:         []string{"wasm", str},
	}
	c := runner.Config()
	assert.Equal(t, cfg, c)
	assert.Nil(t, gotResponse)
}

type ValidStruct struct {
	SomeInt    int
	SomeString string
	SomeTime   time.Time
}

type PrivateFieldStruct struct {
	SomeInt         int
	SomeString      string
	somePrivateTime time.Time
}

func TestRunner_Run_ExecuteCompute(t *testing.T) {
	now := time.Now().UTC()

	cases := []struct {
		name           string
		expectedOutput any
		compute        func(*sdk.WorkflowSpecFactory, basictrigger.TriggerOutputsCap)
		errorString    string
	}{
		// Success cases
		{
			name:           "valid compute func - bigint",
			expectedOutput: big.NewInt(1),
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (*big.Int, error) {
						return big.NewInt(1), nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - bool",
			expectedOutput: true,
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
						return true, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - bytes",
			expectedOutput: []byte{3},
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) ([]byte, error) {
						return []byte{3}, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - decimal",
			expectedOutput: decimal.NewFromInt(2),
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (decimal.Decimal, error) {
						return decimal.NewFromInt(2), nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - float64",
			expectedOutput: float64(1.1),
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (float64, error) {
						return 1.1, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - int",
			expectedOutput: int64(1),
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (int, error) {
						return 1, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - list",
			expectedOutput: []interface{}([]interface{}{int64(1), int64(2), int64(3), int64(4)}),
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) ([]int, error) {
						return []int{1, 2, 3, 4}, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - map",
			expectedOutput: map[string]interface{}(map[string]interface{}{"test": int64(1)}),
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (map[string]int, error) {
						out := map[string]int{"test": 1}
						return out, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - deep map",
			expectedOutput: map[string]interface{}(map[string]interface{}{"test1": map[string]interface{}{"test2": int64(1)}}),
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (map[string]map[string]int, error) {
						out := map[string]map[string]int{"test1": {"test2": 1}}
						return out, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - string",
			expectedOutput: "hiya",
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (string, error) {
						return "hiya", nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - struct",
			expectedOutput: map[string]interface{}(map[string]interface{}{"SomeInt": int64(3), "SomeString": "hiya", "SomeTime": now}),
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (ValidStruct, error) {
						return ValidStruct{SomeString: "hiya", SomeTime: now, SomeInt: 3}, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - empty interface",
			expectedOutput: nil,
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (interface{}, error) {
						var empty interface{}
						return empty, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - time",
			expectedOutput: now,
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (time.Time, error) {
						return now, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - any",
			expectedOutput: now,
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (any, error) {
						return now, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - nil",
			expectedOutput: nil,
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (any, error) {
						return nil, nil
					},
				)
			},
			errorString: "",
		},
		{
			name:           "valid compute func - private struct",
			expectedOutput: map[string]interface{}(map[string]interface{}{"SomeInt": int64(3), "SomeString": "hiya"}),
			compute: func(workflow *sdk.WorkflowSpecFactory, trigger basictrigger.TriggerOutputsCap) {
				sdk.Compute1(
					workflow,
					"compute",
					sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
					func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (PrivateFieldStruct, error) {
						return PrivateFieldStruct{SomeString: "hiya", somePrivateTime: now, SomeInt: 3}, nil
					},
				)
			},
			errorString: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			workflow := sdk.NewWorkflowSpecFactory()

			trigger := basictrigger.TriggerConfig{Name: "trigger", Number: 100}.New(workflow)

			tt.compute(workflow, trigger)

			var gotResponse *wasmpb.Response
			responseFn := func(resp *wasmpb.Response) {
				gotResponse = resp
			}

			m, err := values.NewMap(map[string]any{
				"cool_output": "a trigger event",
			})
			require.NoError(t, err)

			req := capabilities.CapabilityRequest{
				Config: values.EmptyMap(),
				Inputs: m,
				Metadata: capabilities.RequestMetadata{
					ReferenceID: "compute",
				},
			}
			reqpb := capabilitiespb.CapabilityRequestToProto(req)
			request := &wasmpb.Request{
				Id: uuid.New().String(),
				Message: &wasmpb.Request_ComputeRequest{
					ComputeRequest: &wasmpb.ComputeRequest{
						Request: reqpb,
					},
				},
			}
			str, err := marshalRequest(request)
			require.NoError(t, err)
			runner := &Runner{
				args:         []string{"wasm", str},
				sendResponse: responseFn,
				sdkFactory: func(cfg *RuntimeConfig, _ ...func(*RuntimeConfig)) *Runtime {
					return nil
				},
			}
			runner.Run(workflow)

			if tt.errorString == "" {
				assert.NotNil(t, gotResponse.GetComputeResponse())
				resp := gotResponse.GetComputeResponse().GetResponse()
				assert.Equal(t, resp.Error, "")

				m, err = values.FromMapValueProto(resp.Value)
				require.NoError(t, err)

				unw, err := values.Unwrap(m)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedOutput, unw.(map[string]any)["Value"])
			} else {
				assert.Equal(t, tt.errorString, gotResponse.ErrMsg)
				assert.Nil(t, gotResponse.GetComputeResponse())
			}
		})
	}
}

func TestRunner_Run_GetWorkflowSpec(t *testing.T) {
	workflow := sdk.NewWorkflowSpecFactory()

	trigger := basictrigger.TriggerConfig{Name: "trigger", Number: 100}.New(workflow)
	// Define and add a target to the workflow
	targetInput := basictarget.TargetInput{CoolInput: trigger.CoolOutput()}
	targetConfig := basictarget.TargetConfig{Name: "basictarget", Number: 150}
	targetConfig.New(workflow, targetInput)
	computeFn := func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
		return true, nil
	}
	sdk.Compute1(
		workflow,
		"compute",
		sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
		computeFn,
	)

	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	request := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_SpecRequest{
			SpecRequest: &emptypb.Empty{},
		},
	}
	str, err := marshalRequest(request)
	require.NoError(t, err)
	runner := &Runner{
		args:         []string{"wasm", str},
		sendResponse: responseFn,
	}
	runner.Run(workflow)

	resp := gotResponse.GetSpecResponse()
	assert.NotNil(t, resp)

	spc, err := wasmpb.ProtoToWorkflowSpec(resp)
	require.NoError(t, err)

	gotSpec, err := workflow.Spec()
	require.NoError(t, err)

	// Do some massaging due to protos lossy conversion of types
	gotSpec.Triggers[0].Inputs.Mapping = map[string]any{}
	gotSpec.Triggers[0].Config["number"] = int64(gotSpec.Triggers[0].Config["number"].(uint64))
	gotSpec.Targets[0].Config["number"] = int64(gotSpec.Targets[0].Config["number"].(uint64))
	assert.Equal(t, &gotSpec, spc)

	// Verify the target is included in the workflow spec
	assert.Equal(t, targetConfig.Number, uint64(gotSpec.Targets[0].Config["number"].(int64)))
}

// Test_createEmitFn validates the runtime's emit function implementation.  Uses mocks of the
// imported wasip1 emit function.
func Test_createEmitFn(t *testing.T) {
	var (
		l         = logger.Test(t)
		reqId     = "random-id"
		sdkConfig = &RuntimeConfig{
			MaxFetchResponseSizeBytes: 1_000,
			Metadata: &capabilities.RequestMetadata{
				WorkflowID:          "workflow_id",
				WorkflowExecutionID: "workflow_execution_id",
				WorkflowName:        "workflow_name",
				WorkflowOwner:       "workflow_owner_address",
			},
			RequestID: &reqId,
		}
		giveMsg    = "testing guest"
		giveLabels = map[string]string{
			"some-key": "some-value",
		}
	)

	t.Run("success", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 0
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.NoError(t, err)
	})

	t.Run("successfully read error message when emit fails", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			// marshall the protobufs
			b, err := proto.Marshal(&wasmpb.EmitMessageResponse{
				Error: &wasmpb.Error{
					Message: assert.AnError.Error(),
				},
			})
			assert.NoError(t, err)

			// write the marshalled response message to memory
			resp := unsafe.Slice((*byte)(respptr), len(b))
			copy(resp, b)

			// write the length of the response to memory in little endian
			respLen := unsafe.Slice((*byte)(resplenptr), uint32Size)
			binary.LittleEndian.PutUint32(respLen, uint32(len(b)))

			return 0
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.Error(t, err)
		assert.ErrorContains(t, err, assert.AnError.Error())
	})

	t.Run("fail to deserialize response from memory", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			// b is a non-protobuf byte slice
			b := []byte(assert.AnError.Error())

			// write the marshalled response message to memory
			resp := unsafe.Slice((*byte)(respptr), len(b))
			copy(resp, b)

			// write the length of the response to memory in little endian
			respLen := unsafe.Slice((*byte)(resplenptr), uint32Size)
			binary.LittleEndian.PutUint32(respLen, uint32(len(b)))

			return 0
		}

		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "invalid wire-format data")
	})

	t.Run("fail with nonzero code from emit", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 42
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "emit failed with errno 42")
	})
}

func Test_createFetchFn(t *testing.T) {
	var (
		l         = logger.Test(t)
		requestID = uuid.New().String()
		sdkConfig = &RuntimeConfig{
			RequestID:                 &requestID,
			MaxFetchResponseSizeBytes: 1_000,
			Metadata: &capabilities.RequestMetadata{
				WorkflowID:          "workflow_id",
				WorkflowExecutionID: "workflow_execution_id",
				WorkflowName:        "workflow_name",
				WorkflowOwner:       "workflow_owner_address",
			},
		}
	)

	t.Run("OK-success", func(t *testing.T) {
		hostFetch := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 0
		}
		runtimeFetch := createFetchFn(sdkConfig, l, hostFetch)
		response, err := runtimeFetch(sdk.FetchRequest{})
		assert.NoError(t, err)
		assert.Equal(t, sdk.FetchResponse{
			Headers: map[string]any{},
		}, response)
	})

	t.Run("NOK-config_missing_request_id", func(t *testing.T) {
		invalidConfig := &RuntimeConfig{
			RequestID:                 nil,
			MaxFetchResponseSizeBytes: 1_000,
			Metadata: &capabilities.RequestMetadata{
				WorkflowID:          "workflow_id",
				WorkflowExecutionID: "workflow_execution_id",
				WorkflowName:        "workflow_name",
				WorkflowOwner:       "workflow_owner_address",
			},
		}
		hostFetch := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 0
		}
		runtimeFetch := createFetchFn(invalidConfig, l, hostFetch)
		_, err := runtimeFetch(sdk.FetchRequest{})
		assert.ErrorContains(t, err, "request ID is required to fetch")
	})

	t.Run("NOK-fetch_returns_handled_error", func(t *testing.T) {
		hostFetch := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			fetchResponse := &wasmpb.FetchResponse{
				ExecutionError: true,
				ErrorMessage:   assert.AnError.Error(),
			}
			respBytes, perr := proto.Marshal(fetchResponse)
			if perr != nil {
				return 0
			}

			// write the marshalled response message to memory
			resp := unsafe.Slice((*byte)(respptr), len(respBytes))
			copy(resp, respBytes)

			// write the length of the response to memory in little endian
			respLen := unsafe.Slice((*byte)(resplenptr), uint32Size)
			binary.LittleEndian.PutUint32(respLen, uint32(len(respBytes)))

			return 0
		}
		runtimeFetch := createFetchFn(sdkConfig, l, hostFetch)
		_, err := runtimeFetch(sdk.FetchRequest{})
		assert.ErrorContains(t, err, assert.AnError.Error())
	})
}

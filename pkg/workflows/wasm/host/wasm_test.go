package host

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

const (
	successBinaryLocation    = "test/success/cmd/testmodule.wasm"
	successBinaryCmd         = "test/success/cmd"
	failureBinaryLocation    = "test/fail/cmd/testmodule.wasm"
	failureBinaryCmd         = "test/fail/cmd"
	oomBinaryLocation        = "test/oom/cmd/testmodule.wasm"
	oomBinaryCmd             = "test/oom/cmd"
	sleepBinaryLocation      = "test/sleep/cmd/testmodule.wasm"
	sleepBinaryCmd           = "test/sleep/cmd"
	filesBinaryLocation      = "test/files/cmd/testmodule.wasm"
	filesBinaryCmd           = "test/files/cmd"
	dirsBinaryLocation       = "test/dirs/cmd/testmodule.wasm"
	dirsBinaryCmd            = "test/dirs/cmd"
	httpBinaryLocation       = "test/http/cmd/testmodule.wasm"
	httpBinaryCmd            = "test/http/cmd"
	envBinaryLocation        = "test/env/cmd/testmodule.wasm"
	envBinaryCmd             = "test/env/cmd"
	logBinaryLocation        = "test/log/cmd/testmodule.wasm"
	logBinaryCmd             = "test/log/cmd"
	fetchBinaryLocation      = "test/fetch/cmd/testmodule.wasm"
	fetchBinaryCmd           = "test/fetch/cmd"
	fetchlimitBinaryLocation = "test/fetchlimit/cmd/testmodule.wasm"
	fetchlimitBinaryCmd      = "test/fetchlimit/cmd"
	randBinaryLocation       = "test/rand/cmd/testmodule.wasm"
	randBinaryCmd            = "test/rand/cmd"
	emitBinaryLocation       = "test/emit/cmd/testmodule.wasm"
	emitBinaryCmd            = "test/emit/cmd"
)

func createTestBinary(outputPath, path string, uncompressed bool, t *testing.T) []byte {
	cmd := exec.Command("go", "build", "-o", path, fmt.Sprintf("github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/%s", outputPath)) // #nosec
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	binary, err := os.ReadFile(path)
	require.NoError(t, err)

	if uncompressed {
		return binary
	}

	var b bytes.Buffer
	bwr := brotli.NewWriter(&b)
	_, err = bwr.Write(binary)
	require.NoError(t, err)
	require.NoError(t, bwr.Close())

	cb, err := io.ReadAll(&b)
	require.NoError(t, err)
	return cb
}

func Test_GetWorkflowSpec(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(successBinaryCmd, successBinaryLocation, true, t)

	spec, err := GetWorkflowSpec(
		ctx,
		&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		},
		binary,
		[]byte(""),
	)
	require.NoError(t, err)

	assert.Equal(t, spec.Name, "tester")
	assert.Equal(t, spec.Owner, "ryan")
}

func Test_GetWorkflowSpec_UncompressedBinary(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(successBinaryCmd, successBinaryLocation, false, t)

	spec, err := GetWorkflowSpec(
		ctx,
		&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: false,
		},
		binary,
		[]byte(""),
	)
	require.NoError(t, err)

	assert.Equal(t, spec.Name, "tester")
	assert.Equal(t, spec.Owner, "ryan")
}

func Test_GetWorkflowSpec_BinaryErrors(t *testing.T) {
	ctx := tests.Context(t)
	failBinary := createTestBinary(failureBinaryCmd, failureBinaryLocation, true, t)

	_, err := GetWorkflowSpec(
		ctx,
		&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		},
		failBinary,
		[]byte(""),
	)
	// panic
	assert.ErrorContains(t, err, "status 2")
}

func Test_GetWorkflowSpec_Timeout(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(successBinaryCmd, successBinaryLocation, true, t)

	d := time.Duration(0)
	_, err := GetWorkflowSpec(
		ctx,
		&ModuleConfig{
			Timeout:        &d,
			Logger:         logger.Test(t),
			IsUncompressed: true,
		},
		binary, // use the success binary with a zero timeout
		[]byte(""),
	)
	// panic
	assert.ErrorContains(t, err, "wasm trap: interrupt")
}

func Test_Compute_Logs(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(logBinaryCmd, logBinaryLocation, true, t)

	logger, logs := logger.TestObserved(t, zapcore.InfoLevel)
	m, err := NewModule(&ModuleConfig{
		Logger:         logger,
		IsUncompressed: true,
		Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
			return nil, nil
		},
	}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{
			ComputeRequest: &wasmpb.ComputeRequest{
				Request: &capabilitiespb.CapabilityRequest{
					Inputs: &valuespb.Map{},
					Config: &valuespb.Map{},
					Metadata: &capabilitiespb.RequestMetadata{
						ReferenceId: "transform",
					},
				},
			},
		},
	}
	_, err = m.Run(ctx, req)
	assert.Nil(t, err)

	require.Len(t, logs.AllUntimed(), 1)
	expectedEntries := []Entry{
		{
			Log: zapcore.Entry{Level: zapcore.InfoLevel, Message: "building workflow..."},
			Fields: []zapcore.Field{
				zap.String("test-string-field-key", "this is a test field content"),
				zap.Float64("test-numeric-field-key", 6400000),
			},
		},
	}
	for i := range expectedEntries {
		assert.Equal(t, expectedEntries[i].Log.Level, logs.AllUntimed()[i].Entry.Level)
		assert.Equal(t, expectedEntries[i].Log.Message, logs.AllUntimed()[i].Entry.Message)
		assert.ElementsMatch(t, expectedEntries[i].Fields, logs.AllUntimed()[i].Context)
	}
}

func Test_Compute_Emit(t *testing.T) {
	t.Parallel()
	binary := createTestBinary(emitBinaryCmd, emitBinaryLocation, true, t)

	lggr := logger.Test(t)

	req := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{
			ComputeRequest: &wasmpb.ComputeRequest{
				Request: &capabilitiespb.CapabilityRequest{
					Inputs: &valuespb.Map{},
					Config: &valuespb.Map{},
					Metadata: &capabilitiespb.RequestMetadata{
						ReferenceId:         "transform",
						WorkflowId:          "workflow-id",
						WorkflowName:        "workflow-name",
						WorkflowOwner:       "workflow-owner",
						WorkflowExecutionId: "workflow-execution-id",
					},
				},
			},
		},
	}

	fetchFunc := func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
		return nil, nil
	}

	t.Run("successfully call emit with metadata in labels", func(t *testing.T) {
		ctxKey := "key"
		ctx := tests.Context(t)
		ctxValue := "test-value"
		ctx = context.WithValue(ctx, ctxKey, ctxValue)
		m, err := NewModule(&ModuleConfig{
			Logger:         lggr,
			Fetch:          fetchFunc,
			IsUncompressed: true,
			Labeler: newMockMessageEmitter(func(gotCtx context.Context, msg string, kvs map[string]string) error {
				t.Helper()

				v := ctx.Value(ctxKey)
				assert.Equal(t, ctxValue, v)

				assert.Equal(t, "testing emit", msg)
				assert.Equal(t, "this is a test field content", kvs["test-string-field-key"])
				assert.Equal(t, "workflow-id", kvs["workflow_id"])
				assert.Equal(t, "workflow-name", kvs["workflow_name"])
				assert.Equal(t, "workflow-owner", kvs["workflow_owner_address"])
				assert.Equal(t, "workflow-execution-id", kvs["workflow_execution_id"])
				return nil
			}),
		}, binary)
		require.NoError(t, err)

		m.Start()

		_, err = m.Run(ctx, req)
		assert.Nil(t, err)
	})

	t.Run("failure on emit writes to error chain and logs", func(t *testing.T) {
		lggr, logs := logger.TestObserved(t, zapcore.InfoLevel)

		m, err := NewModule(&ModuleConfig{
			Logger:         lggr,
			Fetch:          fetchFunc,
			IsUncompressed: true,
			Labeler: newMockMessageEmitter(func(_ context.Context, msg string, kvs map[string]string) error {
				t.Helper()

				assert.Equal(t, "testing emit", msg)
				assert.Equal(t, "this is a test field content", kvs["test-string-field-key"])
				assert.Equal(t, "workflow-id", kvs["workflow_id"])
				assert.Equal(t, "workflow-name", kvs["workflow_name"])
				assert.Equal(t, "workflow-owner", kvs["workflow_owner_address"])
				assert.Equal(t, "workflow-execution-id", kvs["workflow_execution_id"])

				return assert.AnError
			}),
		}, binary)
		require.NoError(t, err)

		m.Start()

		ctx := tests.Context(t)
		_, err = m.Run(ctx, req)
		assert.Error(t, err)
		assert.ErrorContains(t, err, assert.AnError.Error())

		require.Len(t, logs.AllUntimed(), 1)

		expectedEntries := []Entry{
			{
				Log: zapcore.Entry{Level: zapcore.ErrorLevel, Message: fmt.Sprintf("error emitting message: %s", assert.AnError)},
			},
		}
		for i := range expectedEntries {
			assert.Equal(t, expectedEntries[i].Log.Level, logs.AllUntimed()[i].Entry.Level)
			assert.Equal(t, expectedEntries[i].Log.Message, logs.AllUntimed()[i].Entry.Message)
		}
	})

	t.Run("failure on emit due to missing workflow identifying metadata", func(t *testing.T) {
		lggr := logger.Test(t)

		m, err := NewModule(&ModuleConfig{
			Logger:         lggr,
			Fetch:          fetchFunc,
			IsUncompressed: true,
			Labeler: newMockMessageEmitter(func(_ context.Context, msg string, labels map[string]string) error {
				return nil
			}), // never called
		}, binary)
		require.NoError(t, err)

		m.Start()

		req = &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
				},
			},
		}

		ctx := tests.Context(t)
		_, err = m.Run(ctx, req)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to create emission")
	})
}

func Test_Compute_Fetch(t *testing.T) {
	t.Parallel()
	binary := createTestBinary(fetchBinaryCmd, fetchBinaryLocation, true, t)

	t.Run("OK_default_runtime_cfg", func(t *testing.T) {
		t.Parallel()
		ctx := tests.Context(t)
		expected := sdk.FetchResponse{
			ExecutionError: false,
			Body:           []byte("valid-response"),
			StatusCode:     http.StatusOK,
			Headers:        map[string]any{},
		}

		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
				return &wasmpb.FetchResponse{
					ExecutionError: expected.ExecutionError,
					Body:           expected.Body,
					StatusCode:     uint32(expected.StatusCode),
				}, nil
			},
		}, binary)
		require.NoError(t, err)

		m.Start()

		req := &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
				},
			},
		}
		response, err := m.Run(ctx, req)
		assert.Nil(t, err)

		actual := sdk.FetchResponse{}
		r, err := pb.CapabilityResponseFromProto(response.GetComputeResponse().GetResponse())
		require.NoError(t, err)
		err = r.Value.Underlying["Value"].UnwrapTo(&actual)
		require.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("OK_custom_runtime_cfg", func(t *testing.T) {
		t.Parallel()
		ctx := tests.Context(t)
		expected := sdk.FetchResponse{
			ExecutionError: false,
			Body:           []byte("valid-response"),
			StatusCode:     http.StatusOK,
			Headers:        map[string]any{},
		}

		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
				return &wasmpb.FetchResponse{
					ExecutionError: expected.ExecutionError,
					Body:           expected.Body,
					StatusCode:     uint32(expected.StatusCode),
				}, nil
			},
		}, binary)
		require.NoError(t, err)

		m.Start()

		req := &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
					RuntimeConfig: &wasmpb.RuntimeConfig{
						MaxFetchResponseSizeBytes: 2 * 1024,
					},
				},
			},
		}
		response, err := m.Run(ctx, req)
		assert.Nil(t, err)

		actual := sdk.FetchResponse{}
		r, err := pb.CapabilityResponseFromProto(response.GetComputeResponse().GetResponse())
		require.NoError(t, err)
		err = r.Value.Underlying["Value"].UnwrapTo(&actual)
		require.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("NOK_fetch_error_returned", func(t *testing.T) {
		t.Parallel()
		ctx := tests.Context(t)
		logger, logs := logger.TestObserved(t, zapcore.InfoLevel)

		m, err := NewModule(&ModuleConfig{
			Logger:         logger,
			IsUncompressed: true,
			Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
				return nil, assert.AnError
			},
		}, binary)
		require.NoError(t, err)

		m.Start()

		req := &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
				},
			},
		}
		_, err = m.Run(ctx, req)
		assert.NotNil(t, err)
		assert.ErrorContains(t, err, assert.AnError.Error())
		require.Len(t, logs.AllUntimed(), 1)

		expectedEntries := []Entry{
			{
				Log: zapcore.Entry{Level: zapcore.ErrorLevel, Message: fmt.Sprintf("error calling fetch: %s", assert.AnError)},
			},
		}
		for i := range expectedEntries {
			assert.Equal(t, expectedEntries[i].Log.Level, logs.AllUntimed()[i].Entry.Level)
			assert.Equal(t, expectedEntries[i].Log.Message, logs.AllUntimed()[i].Entry.Message)
		}
	})

	t.Run("OK_context_propagation", func(t *testing.T) {
		t.Parallel()
		type testkey string
		var key testkey = "test-key"
		var expectedValue string = "test-value"

		expected := sdk.FetchResponse{
			ExecutionError: false,
			Body:           []byte(expectedValue),
			StatusCode:     http.StatusOK,
			Headers:        map[string]any{},
		}

		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
				return &wasmpb.FetchResponse{
					ExecutionError: expected.ExecutionError,
					Body:           []byte(ctx.Value(key).(string)),
					StatusCode:     uint32(expected.StatusCode),
				}, nil
			},
		}, binary)
		require.NoError(t, err)

		m.Start()

		req := &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
					RuntimeConfig: &wasmpb.RuntimeConfig{
						MaxFetchResponseSizeBytes: 2 * 1024,
					},
				},
			},
		}

		ctx := context.WithValue(tests.Context(t), key, expectedValue)
		response, err := m.Run(ctx, req)
		assert.Nil(t, err)

		actual := sdk.FetchResponse{}
		r, err := pb.CapabilityResponseFromProto(response.GetComputeResponse().GetResponse())
		require.NoError(t, err)
		err = r.Value.Underlying["Value"].UnwrapTo(&actual)
		require.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("OK_context_cancelation", func(t *testing.T) {
		t.Parallel()
		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
				select {
				case <-ctx.Done():
					return nil, assert.AnError
				default:
					return &wasmpb.FetchResponse{}, nil
				}
			},
		}, binary)
		require.NoError(t, err)

		m.Start()

		req := &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
					RuntimeConfig: &wasmpb.RuntimeConfig{
						MaxFetchResponseSizeBytes: 2 * 1024,
					},
				},
			},
		}

		ctx, cancel := context.WithCancel(tests.Context(t))
		cancel()
		_, err = m.Run(ctx, req)
		require.NotNil(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("error executing runner: error executing custom compute: %s", assert.AnError))
	})

	t.Run("NOK_exceed_amout_of_defined_max_fetch_calls", func(t *testing.T) {
		t.Parallel()
		binary := createTestBinary(fetchlimitBinaryCmd, fetchlimitBinaryLocation, true, t)
		ctx := tests.Context(t)
		expected := sdk.FetchResponse{
			ExecutionError: false,
			Body:           []byte("valid-response"),
			StatusCode:     http.StatusOK,
			Headers:        map[string]any{},
		}

		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
				return &wasmpb.FetchResponse{
					ExecutionError: expected.ExecutionError,
					Body:           expected.Body,
					StatusCode:     uint32(expected.StatusCode),
				}, nil
			},
			MaxFetchRequests: 1,
		}, binary)
		require.NoError(t, err)

		m.Start()

		req := &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
				},
			},
		}
		_, err = m.Run(ctx, req)
		require.NotNil(t, err)
	})

	t.Run("NOK_exceed_amout_of_default_max_fetch_calls", func(t *testing.T) {
		t.Parallel()
		binary := createTestBinary(fetchlimitBinaryCmd, fetchlimitBinaryLocation, true, t)
		ctx := tests.Context(t)
		expected := sdk.FetchResponse{
			ExecutionError: false,
			Body:           []byte("valid-response"),
			StatusCode:     http.StatusOK,
			Headers:        map[string]any{},
		}

		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
				return &wasmpb.FetchResponse{
					ExecutionError: expected.ExecutionError,
					Body:           expected.Body,
					StatusCode:     uint32(expected.StatusCode),
				}, nil
			},
		}, binary)
		require.NoError(t, err)

		m.Start()

		req := &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
				},
			},
		}
		_, err = m.Run(ctx, req)
		require.NotNil(t, err)
	})

	t.Run("OK_making_up_to_max_fetch_calls", func(t *testing.T) {
		t.Parallel()
		binary := createTestBinary(fetchlimitBinaryCmd, fetchlimitBinaryLocation, true, t)
		ctx := tests.Context(t)
		expected := sdk.FetchResponse{
			ExecutionError: false,
			Body:           []byte("valid-response"),
			StatusCode:     http.StatusOK,
			Headers:        map[string]any{},
		}

		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
				return &wasmpb.FetchResponse{
					ExecutionError: expected.ExecutionError,
					Body:           expected.Body,
					StatusCode:     uint32(expected.StatusCode),
				}, nil
			},
			MaxFetchRequests: 6,
		}, binary)
		require.NoError(t, err)

		m.Start()

		req := &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
				},
			},
		}
		_, err = m.Run(ctx, req)
		require.Nil(t, err)
	})

	t.Run("OK_multiple_request_reusing_module", func(t *testing.T) {
		t.Parallel()
		binary := createTestBinary(fetchlimitBinaryCmd, fetchlimitBinaryLocation, true, t)
		ctx := tests.Context(t)
		expected := sdk.FetchResponse{
			ExecutionError: false,
			Body:           []byte("valid-response"),
			StatusCode:     http.StatusOK,
			Headers:        map[string]any{},
		}

		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Fetch: func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
				return &wasmpb.FetchResponse{
					ExecutionError: expected.ExecutionError,
					Body:           expected.Body,
					StatusCode:     uint32(expected.StatusCode),
				}, nil
			},
			MaxFetchRequests: 6,
		}, binary)
		require.NoError(t, err)

		m.Start()

		req := &wasmpb.Request{
			Id: uuid.New().String(),
			Message: &wasmpb.Request_ComputeRequest{
				ComputeRequest: &wasmpb.ComputeRequest{
					Request: &capabilitiespb.CapabilityRequest{
						Inputs: &valuespb.Map{},
						Config: &valuespb.Map{},
						Metadata: &capabilitiespb.RequestMetadata{
							ReferenceId: "transform",
						},
					},
				},
			},
		}
		_, err = m.Run(ctx, req)
		require.Nil(t, err)

		// we can reuse the request because after completion it gets deleted from the store
		_, err = m.Run(ctx, req)
		require.Nil(t, err)
	})

}

func TestModule_Errors(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(successBinaryCmd, successBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{IsUncompressed: true, Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	_, err = m.Run(ctx, nil)
	assert.ErrorContains(t, err, "invalid request: can't be nil")

	req := &wasmpb.Request{
		Id: "",
	}
	_, err = m.Run(ctx, req)
	assert.ErrorContains(t, err, "invalid request: can't be empty")

	req = &wasmpb.Request{
		Id: uuid.New().String(),
	}
	_, err = m.Run(ctx, req)
	assert.ErrorContains(t, err, "invalid request: message must be SpecRequest or ComputeRequest")

	req = &wasmpb.Request{
		Id:      uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{},
	}
	_, err = m.Run(ctx, req)
	assert.ErrorContains(t, err, "invalid compute request: nil request")

	m.Start()

	req = &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{
			ComputeRequest: &wasmpb.ComputeRequest{
				Request: &capabilitiespb.CapabilityRequest{
					Metadata: &capabilitiespb.RequestMetadata{
						ReferenceId: "doesnt-exist",
					},
				},
			},
		},
	}
	_, err = m.Run(ctx, req)
	assert.ErrorContains(t, err, "invalid compute request: could not find compute function for id doesnt-exist")
}

func TestModule_Sandbox_Memory(t *testing.T) {
	ctx := tests.Context(t)
	binary := createTestBinary(oomBinaryCmd, oomBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{IsUncompressed: true, Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id:      uuid.New().String(),
		Message: &wasmpb.Request_SpecRequest{},
	}
	_, err = m.Run(ctx, req)
	assert.ErrorContains(t, err, "exit status 2")
}

func TestModule_Sandbox_SleepIsStubbedOut(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(sleepBinaryCmd, sleepBinaryLocation, true, t)

	d := 1 * time.Millisecond
	m, err := NewModule(&ModuleConfig{Timeout: &d, IsUncompressed: true, Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id:      uuid.New().String(),
		Message: &wasmpb.Request_SpecRequest{},
	}

	start := time.Now()
	_, err = m.Run(ctx, req)
	end := time.Now()

	// The binary sleeps for 1 hour,
	// but with our stubbed out functions,
	// it should execute and return almost immediately.
	assert.WithinDuration(t, start, end, 10*time.Second)
	assert.NotNil(t, err)
}

func TestModule_Sandbox_Timeout(t *testing.T) {
	ctx := tests.Context(t)
	binary := createTestBinary(sleepBinaryCmd, sleepBinaryLocation, true, t)

	tmt := 10 * time.Millisecond
	m, err := NewModule(&ModuleConfig{IsUncompressed: true, Logger: logger.Test(t), Timeout: &tmt}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id:      uuid.New().String(),
		Message: &wasmpb.Request_SpecRequest{},
	}

	_, err = m.Run(ctx, req)

	assert.ErrorContains(t, err, "interrupt")
}

func TestModule_Sandbox_CantReadFiles(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(filesBinaryCmd, filesBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{IsUncompressed: true, Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{
			ComputeRequest: &wasmpb.ComputeRequest{
				Request: &capabilitiespb.CapabilityRequest{
					Inputs: &valuespb.Map{},
					Config: &valuespb.Map{},
					Metadata: &capabilitiespb.RequestMetadata{
						ReferenceId: "transform",
					},
				},
			},
		},
	}
	_, err = m.Run(ctx, req)
	assert.ErrorContains(t, err, "open /tmp/file")
}

func TestModule_Sandbox_CantCreateDir(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(dirsBinaryCmd, dirsBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{IsUncompressed: true, Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{
			ComputeRequest: &wasmpb.ComputeRequest{
				Request: &capabilitiespb.CapabilityRequest{
					Inputs: &valuespb.Map{},
					Config: &valuespb.Map{},
					Metadata: &capabilitiespb.RequestMetadata{
						ReferenceId: "transform",
					},
				},
			},
		},
	}
	_, err = m.Run(ctx, req)
	assert.ErrorContains(t, err, "mkdir")
}

func TestModule_Sandbox_HTTPRequest(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(httpBinaryCmd, httpBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{IsUncompressed: true, Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{
			ComputeRequest: &wasmpb.ComputeRequest{
				Request: &capabilitiespb.CapabilityRequest{
					Inputs: &valuespb.Map{},
					Config: &valuespb.Map{},
					Metadata: &capabilitiespb.RequestMetadata{
						ReferenceId: "transform",
					},
				},
			},
		},
	}
	_, err = m.Run(ctx, req)
	assert.NotNil(t, err)
}

func TestModule_Sandbox_ReadEnv(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	binary := createTestBinary(envBinaryCmd, envBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{IsUncompressed: true, Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	m.Start()

	os.Setenv("FOO", "BAR")
	defer os.Unsetenv("FOO")

	req := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{
			ComputeRequest: &wasmpb.ComputeRequest{
				Request: &capabilitiespb.CapabilityRequest{
					Inputs: &valuespb.Map{},
					Config: &valuespb.Map{},
					Metadata: &capabilitiespb.RequestMetadata{
						ReferenceId: "transform",
					},
				},
			},
		},
	}
	// This will return an error if FOO == BAR in the WASM binary
	_, err = m.Run(ctx, req)
	assert.Nil(t, err)
}

func TestModule_Sandbox_RandomGet(t *testing.T) {
	t.Parallel()
	req := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{
			ComputeRequest: &wasmpb.ComputeRequest{
				Request: &capabilitiespb.CapabilityRequest{
					Inputs: &valuespb.Map{},
					Config: &valuespb.Map{},
					Metadata: &capabilitiespb.RequestMetadata{
						ReferenceId: "transform",
					},
				},
			},
		},
	}
	t.Run("success: deterministic override via module config", func(t *testing.T) {
		ctx := tests.Context(t)
		binary := createTestBinary(randBinaryCmd, randBinaryLocation, true, t)

		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Determinism: &DeterminismConfig{
				Seed: 42,
			},
		}, binary)
		require.NoError(t, err)

		m.Start()

		_, err = m.Run(ctx, req)
		assert.Nil(t, err)
	})

	t.Run("success: default module config is non deterministic", func(t *testing.T) {
		ctx := tests.Context(t)
		binary := createTestBinary(randBinaryCmd, randBinaryLocation, true, t)

		m, err := NewModule(&ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		}, binary)
		require.NoError(t, err)

		m.Start()

		_, err = m.Run(ctx, req)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "expected deterministic output")
	})
}

type Entry struct {
	Log    zapcore.Entry
	Fields []zapcore.Field
}

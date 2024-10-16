package host

import (
	"bytes"
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
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

const (
	successBinaryLocation = "test/success/cmd/testmodule.wasm"
	successBinaryCmd      = "test/success/cmd"
	failureBinaryLocation = "test/fail/cmd/testmodule.wasm"
	failureBinaryCmd      = "test/fail/cmd"
	oomBinaryLocation     = "test/oom/cmd/testmodule.wasm"
	oomBinaryCmd          = "test/oom/cmd"
	sleepBinaryLocation   = "test/sleep/cmd/testmodule.wasm"
	sleepBinaryCmd        = "test/sleep/cmd"
	filesBinaryLocation   = "test/files/cmd/testmodule.wasm"
	filesBinaryCmd        = "test/files/cmd"
	dirsBinaryLocation    = "test/dirs/cmd/testmodule.wasm"
	dirsBinaryCmd         = "test/dirs/cmd"
	httpBinaryLocation    = "test/http/cmd/testmodule.wasm"
	httpBinaryCmd         = "test/http/cmd"
	envBinaryLocation     = "test/env/cmd/testmodule.wasm"
	envBinaryCmd          = "test/env/cmd"
	logBinaryLocation     = "test/log/cmd/testmodule.wasm"
	logBinaryCmd          = "test/log/cmd"
	fetchBinaryLocation   = "test/fetch/cmd/testmodule.wasm"
	fetchBinaryCmd        = "test/fetch/cmd"
	randBinaryLocation    = "test/rand/cmd/testmodule.wasm"
	randBinaryCmd         = "test/rand/cmd"
)

func createTestBinary(outputPath, path string, compress bool, t *testing.T) []byte {
	cmd := exec.Command("go", "build", "-o", path, fmt.Sprintf("github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/%s", outputPath)) // #nosec
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	binary, err := os.ReadFile(path)
	require.NoError(t, err)

	if !compress {
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
	binary := createTestBinary(successBinaryCmd, successBinaryLocation, true, t)

	spec, err := GetWorkflowSpec(
		&ModuleConfig{
			Logger: logger.Test(t),
		},
		binary,
		[]byte(""),
	)
	require.NoError(t, err)

	assert.Equal(t, spec.Name, "tester")
	assert.Equal(t, spec.Owner, "ryan")
}

func Test_GetWorkflowSpec_UncompressedBinary(t *testing.T) {
	binary := createTestBinary(successBinaryCmd, successBinaryLocation, false, t)

	spec, err := GetWorkflowSpec(
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

func Test_GetWorkflowSpec_BinaryErrors(t *testing.T) {
	failBinary := createTestBinary(failureBinaryCmd, failureBinaryLocation, true, t)

	_, err := GetWorkflowSpec(
		&ModuleConfig{
			Logger: logger.Test(t),
		},
		failBinary,
		[]byte(""),
	)
	// panic
	assert.ErrorContains(t, err, "status 2")
}

func Test_GetWorkflowSpec_Timeout(t *testing.T) {
	binary := createTestBinary(successBinaryCmd, successBinaryLocation, true, t)

	d := time.Duration(0)
	_, err := GetWorkflowSpec(
		&ModuleConfig{
			Timeout: &d,
			Logger:  logger.Test(t),
		},
		binary, // use the success binary with a zero timeout
		[]byte(""),
	)
	// panic
	assert.ErrorContains(t, err, "wasm trap: interrupt")
}

func Test_Compute_Logs(t *testing.T) {
	binary := createTestBinary(logBinaryCmd, logBinaryLocation, true, t)

	logger, logs := logger.TestObserved(t, zapcore.InfoLevel)
	m, err := NewModule(&ModuleConfig{
		Logger: logger,
		Fetch: func(req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
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
	_, err = m.Run(req)
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

func Test_Compute_Fetch(t *testing.T) {
	binary := createTestBinary(fetchBinaryCmd, fetchBinaryLocation, true, t)

	t.Run("OK_default_runtime_cfg", func(t *testing.T) {
		expected := sdk.FetchResponse{
			ExecutionError: false,
			Body:           []byte("valid-response"),
			StatusCode:     http.StatusOK,
			Headers:        map[string]any{},
		}

		m, err := NewModule(&ModuleConfig{
			Logger: logger.Test(t),
			Fetch: func(req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
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
		response, err := m.Run(req)
		assert.Nil(t, err)

		actual := sdk.FetchResponse{}
		r, err := pb.CapabilityResponseFromProto(response.GetComputeResponse().GetResponse())
		require.NoError(t, err)
		err = r.Value.Underlying["Value"].UnwrapTo(&actual)
		require.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("OK_custom_runtime_cfg", func(t *testing.T) {
		expected := sdk.FetchResponse{
			ExecutionError: false,
			Body:           []byte("valid-response"),
			StatusCode:     http.StatusOK,
			Headers:        map[string]any{},
		}

		m, err := NewModule(&ModuleConfig{
			Logger: logger.Test(t),
			Fetch: func(req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
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
		response, err := m.Run(req)
		assert.Nil(t, err)

		actual := sdk.FetchResponse{}
		r, err := pb.CapabilityResponseFromProto(response.GetComputeResponse().GetResponse())
		require.NoError(t, err)
		err = r.Value.Underlying["Value"].UnwrapTo(&actual)
		require.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("NOK_fetch_error_returned", func(t *testing.T) {
		logger, logs := logger.TestObserved(t, zapcore.InfoLevel)

		m, err := NewModule(&ModuleConfig{
			Logger: logger,
			Fetch: func(req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
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
		_, err = m.Run(req)
		assert.NotNil(t, err)
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
}

func TestModule_Errors(t *testing.T) {
	binary := createTestBinary(successBinaryCmd, successBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	_, err = m.Run(nil)
	assert.ErrorContains(t, err, "invariant violation: invalid request to runner")

	req := &wasmpb.Request{
		Id: uuid.New().String(),
	}
	_, err = m.Run(req)
	assert.ErrorContains(t, err, "invalid request: message must be SpecRequest or ComputeRequest")

	req = &wasmpb.Request{
		Id:      uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{},
	}
	_, err = m.Run(req)
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
	_, err = m.Run(req)
	assert.ErrorContains(t, err, "invalid compute request: could not find compute function for id doesnt-exist")
}

func TestModule_Sandbox_Memory(t *testing.T) {
	binary := createTestBinary(oomBinaryCmd, oomBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id:      uuid.New().String(),
		Message: &wasmpb.Request_SpecRequest{},
	}
	_, err = m.Run(req)
	assert.ErrorContains(t, err, "exit status 2")
}

func TestModule_Sandbox_SleepIsStubbedOut(t *testing.T) {
	binary := createTestBinary(sleepBinaryCmd, sleepBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{Logger: logger.Test(t)}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id:      uuid.New().String(),
		Message: &wasmpb.Request_SpecRequest{},
	}

	start := time.Now()
	_, err = m.Run(req)
	end := time.Now()

	// The binary sleeps for 1 hour,
	// but with our stubbed out functions,
	// it should execute and return almost immediately.
	assert.WithinDuration(t, start, end, 10*time.Second)
	assert.NotNil(t, err)
}

func TestModule_Sandbox_Timeout(t *testing.T) {
	binary := createTestBinary(sleepBinaryCmd, sleepBinaryLocation, true, t)

	tmt := 10 * time.Millisecond
	m, err := NewModule(&ModuleConfig{Logger: logger.Test(t), Timeout: &tmt}, binary)
	require.NoError(t, err)

	m.Start()

	req := &wasmpb.Request{
		Id:      uuid.New().String(),
		Message: &wasmpb.Request_SpecRequest{},
	}

	_, err = m.Run(req)

	assert.ErrorContains(t, err, "interrupt")
}

func TestModule_Sandbox_CantReadFiles(t *testing.T) {
	binary := createTestBinary(filesBinaryCmd, filesBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{Logger: logger.Test(t)}, binary)
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
	_, err = m.Run(req)
	assert.ErrorContains(t, err, "open /tmp/file")
}

func TestModule_Sandbox_CantCreateDir(t *testing.T) {
	binary := createTestBinary(dirsBinaryCmd, dirsBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{Logger: logger.Test(t)}, binary)
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
	_, err = m.Run(req)
	assert.ErrorContains(t, err, "mkdir")
}

func TestModule_Sandbox_HTTPRequest(t *testing.T) {
	binary := createTestBinary(httpBinaryCmd, httpBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{Logger: logger.Test(t)}, binary)
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
	_, err = m.Run(req)
	assert.NotNil(t, err)
}

func TestModule_Sandbox_ReadEnv(t *testing.T) {
	binary := createTestBinary(envBinaryCmd, envBinaryLocation, true, t)

	m, err := NewModule(&ModuleConfig{Logger: logger.Test(t)}, binary)
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
	_, err = m.Run(req)
	assert.Nil(t, err)
}

func TestModule_Sandbox_RandomGet(t *testing.T) {
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
		binary := createTestBinary(randBinaryCmd, randBinaryLocation, true, t)

		m, err := NewModule(&ModuleConfig{
			Logger: logger.Test(t),
			Determinism: &DeterminismConfig{
				Seed: 42,
			},
		}, binary)
		require.NoError(t, err)

		m.Start()

		_, err = m.Run(req)
		assert.Nil(t, err)
	})

	t.Run("success: default module config is non deterministic", func(t *testing.T) {
		binary := createTestBinary(randBinaryCmd, randBinaryLocation, true, t)

		m, err := NewModule(&ModuleConfig{
			Logger: logger.Test(t),
		}, binary)
		require.NoError(t, err)

		m.Start()

		_, err = m.Run(req)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "expected deterministic output")
	})
}

type Entry struct {
	Log    zapcore.Entry
	Fields []zapcore.Field
}

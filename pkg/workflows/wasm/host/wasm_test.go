package host

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
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
)

func createTestBinary(outputPath, path string, t *testing.T) string {
	cmd := exec.Command("go", "build", "-o", path, fmt.Sprintf("github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/%s", outputPath)) // #nosec
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	return path
}

func Test_GetWorkflowSpec(t *testing.T) {
	binary, err := os.ReadFile(createTestBinary(successBinaryCmd, successBinaryLocation, t))
	require.NoError(t, err)

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

func Test_GetWorkflowSpec_BinaryErrors(t *testing.T) {
	failBinary, err := os.ReadFile(createTestBinary(failureBinaryCmd, failureBinaryLocation, t))
	require.NoError(t, err)

	_, err = GetWorkflowSpec(
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
	binary, err := os.ReadFile(createTestBinary(successBinaryCmd, successBinaryLocation, t))
	require.NoError(t, err)

	d := time.Duration(0)
	_, err = GetWorkflowSpec(
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

func TestModule_Errors(t *testing.T) {
	binary, err := os.ReadFile(createTestBinary(successBinaryCmd, successBinaryLocation, t))
	require.NoError(t, err)

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
	binary, err := os.ReadFile(createTestBinary(oomBinaryCmd, oomBinaryLocation, t))
	require.NoError(t, err)

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
	binary, err := os.ReadFile(createTestBinary(sleepBinaryCmd, sleepBinaryLocation, t))
	require.NoError(t, err)

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
	binary, err := os.ReadFile(createTestBinary(sleepBinaryCmd, sleepBinaryLocation, t))
	require.NoError(t, err)

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
	binary, err := os.ReadFile(createTestBinary(filesBinaryCmd, filesBinaryLocation, t))
	require.NoError(t, err)

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
	binary, err := os.ReadFile(createTestBinary(dirsBinaryCmd, dirsBinaryLocation, t))
	require.NoError(t, err)

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
	binary, err := os.ReadFile(createTestBinary(httpBinaryCmd, httpBinaryLocation, t))
	require.NoError(t, err)

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
	binary, err := os.ReadFile(createTestBinary(envBinaryCmd, envBinaryLocation, t))
	require.NoError(t, err)

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

package host

import (
	_ "embed"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

const (
	successBinaryLocation = "test/success/cmd/testmodule.wasm"
	failureBinaryLocation = "test/fail/cmd/testmodule.wasm"
)

func Test_GetWorkflowSpec(t *testing.T) {
	binary, err := os.ReadFile(createTestBinary(successBinaryLocation, t))
	require.NoError(t, err)

	spec, err := GetWorkflowSpec(
		tests.Context(t),
		ModuleConfig{},
		binary,
		[]byte(""),
	)
	require.NoError(t, err)

	assert.Equal(t, spec.Name, "tester")
	assert.Equal(t, spec.Owner, "ryan")
}

func createTestBinary(path string, t *testing.T) string {
	cmd := exec.Command("go", "build", "-o", path, "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/test/cmd")
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, output)

	return path
}

func Test_GetWorkflowSpec_BinaryErrors(t *testing.T) {
	failBinary, err := os.ReadFile(createTestBinary(failureBinaryLocation, t))
	require.NoError(t, err)

	_, err = GetWorkflowSpec(
		tests.Context(t),
		ModuleConfig{},
		failBinary,
		[]byte(""),
	)
	// panic
	assert.ErrorContains(t, err, "status 2")
}

func Test_GetWorkflowSpec_Timeout(t *testing.T) {
	binary, err := os.ReadFile(createTestBinary(successBinaryLocation, t))
	require.NoError(t, err)

	d := time.Duration(0)
	_, err = GetWorkflowSpec(
		tests.Context(t),
		ModuleConfig{
			Timeout: &d,
		},
		binary, // use the success binary with a zero timeout
		[]byte(""),
	)
	// panic
	assert.ErrorContains(t, err, "wasm trap: interrupt")
}

func TestModule_Errors(t *testing.T) {
	binary, err := os.ReadFile(createTestBinary(successBinaryLocation, t))
	require.NoError(t, err)

	m, err := NewModule(ModuleConfig{}, binary)
	require.NoError(t, err)

	_, err = m.Run(nil)
	assert.ErrorContains(t, err, "invalid request: request cannot be empty")

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

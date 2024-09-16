package host

import (
	_ "embed"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func Test_GetWorkflowSpec(t *testing.T) {
	binary, err := os.ReadFile(createTestBinary(t))
	require.NoError(t, err)

	spec, err := GetWorkflowSpec(
		tests.Context(t),
		binary,
		[]byte(""),
	)
	require.NoError(t, err)

	assert.Equal(t, spec.Name, "tester")
	assert.Equal(t, spec.Owner, "ryan")
}

func createTestBinary(t *testing.T) string {
	const testBinaryLocation = "test/cmd/testmodule.wasm"

	cmd := exec.Command("go", "build", "-o", testBinaryLocation, "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/test/cmd")
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, output)

	return testBinaryLocation
}

package host

import (
	_ "embed"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

//go:generate ./test/generate_wasm.sh

func Test_GetWorkflowSpec(t *testing.T) {
	binary, err := os.ReadFile("testmodule.wasm")
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

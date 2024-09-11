package wasm

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

var (
	//go:generate ./generate_wasm.sh
	//go:embed test/cmd/testmodule.wasm
	binary []byte
)

func Test_GetWorkflowSpec(t *testing.T) {
	spec, err := GetWorkflowSpec(
		tests.Context(t),
		binary,
		[]byte(""),
	)
	require.NoError(t, err)

	assert.Equal(t, spec.Name, "tester")
	assert.Equal(t, spec.Owner, "ryan")
}

package test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host"
)

//go:generate ./generate_wasm.sh

var (
	//go:embed cmd/testmodule.wasm
	binary []byte
)

func Test_GetWorkflowSpec(t *testing.T) {
	spec, err := host.GetWorkflowSpec(
		tests.Context(t),
		binary,
		[]byte(""),
	)
	require.NoError(t, err)

	assert.Equal(t, spec.Name, "tester")
	assert.Equal(t, spec.Owner, "ryan")
}

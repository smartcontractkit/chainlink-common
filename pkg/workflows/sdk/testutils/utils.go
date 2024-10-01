package testutils

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func AssertWorkflowSpec(t *testing.T, expectedSpec, testWorkflowSpec sdk.WorkflowSpec) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetIndent("", "  ")
	require.NoError(t, e.Encode(expectedSpec))
	expected := b.String()

	b.Reset()
	require.NoError(t, e.Encode(testWorkflowSpec))
	actual := b.String()

	assert.Equal(t, expected, actual)
}

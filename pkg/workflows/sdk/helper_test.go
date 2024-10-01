package sdk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func (w *WorkflowSpecFactory) MustSpec(t *testing.T) WorkflowSpec {
	t.Helper()
	s, err := w.Spec()
	require.NoError(t, err)
	return s
}

package prometheus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitize(t *testing.T) {
	defer func() { dropSanitizationGateEnabled = false }()

	require.Equal(t, "", NormalizeLabel(""), "")
	require.Equal(t, "key_test", NormalizeLabel("_test"))
	require.Equal(t, "key_0test", NormalizeLabel("0test"))
	require.Equal(t, "test", NormalizeLabel("test"))
	require.Equal(t, "test__", NormalizeLabel("test_/"))
	require.Equal(t, "__test", NormalizeLabel("__test"))
}

func TestSanitizeDropSanitization(t *testing.T) {
	defer func() { dropSanitizationGateEnabled = false }()

	require.Equal(t, "", NormalizeLabel(""))
	require.Equal(t, "_test", NormalizeLabel("_test"))
	require.Equal(t, "key_0test", NormalizeLabel("0test"))
	require.Equal(t, "test", NormalizeLabel("test"))
	require.Equal(t, "__test", NormalizeLabel("__test"))
}

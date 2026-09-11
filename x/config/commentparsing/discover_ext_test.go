// This file is an external test package so it can import the examples, which import the package
// under test. Discover needs real non-test source to parse, and the examples are exactly that.
package commentparsing_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested/upstream"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/simple"
)

const examplesPath = "github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples"

func TestDiscoverAcrossPackages(t *testing.T) {
	// Only the root is named. upstream is reached through Config's embedded and nested fields.
	docs, err := commentparsing.Discover(&nested.Config{})
	require.NoError(t, err)

	t.Run("Reaches a second package without being told about it", func(t *testing.T) {
		require.Equal(t, []string{examplesPath + "/nested", examplesPath + "/nested/upstream"}, docs.Packages())
	})

	t.Run("Documentation of an embedded type comes from its own package", func(t *testing.T) {
		require.Equal(t, "Region is the deployment region.",
			docs.Field(reflect.TypeFor[upstream.Settings](), "Region").Comment)
		require.Equal(t, "required,oneof=us eu ap",
			docs.Field(reflect.TypeFor[upstream.Settings](), "Region").Validate)
	})

	t.Run("Documentation of a nested section comes from its own package", func(t *testing.T) {
		require.Equal(t, "MaxAttempts is the total tries, including the first.",
			docs.Field(reflect.TypeFor[upstream.Retry](), "MaxAttempts").Comment)
	})

	t.Run("Pointer to a discovered type resolves to the same entry", func(t *testing.T) {
		fields, err := docs.Type(reflect.TypeFor[*upstream.Retry]())
		require.NoError(t, err)
		require.Contains(t, fields, "Backoff")
	})

	t.Run("A type outside the tree is reported", func(t *testing.T) {
		_, err := docs.Type(reflect.TypeFor[simple.Config]())
		require.ErrorContains(t, err, "not reachable from the root value")
	})
}

// A duration is decoded from a single string, so demanding fields of it would fail on a dependency
// that has every right to document none.
func TestDiscoverSkipsScalarStructs(t *testing.T) {
	docs, err := commentparsing.Discover(&simple.Config{})
	require.NoError(t, err)
	require.Equal(t, []string{examplesPath + "/simple"}, docs.Packages())
	require.Equal(t, "Retries is the attempt count after the first.",
		docs.Field(reflect.TypeFor[simple.Config](), "Retries").Comment)
}

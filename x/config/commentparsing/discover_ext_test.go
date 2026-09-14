// This file is an external test package so it can import the examples, which import the package
// under test. Discovery needs real non-test source to parse, and the examples are exactly that.
package commentparsing_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/simple"
)

const examplesPath = "github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples"

// fields finds a discovered type's documentation by package path and type name, the way a consumer
// keying on reflect.Type would.
func fields(t *testing.T, pkgs []commentparsing.Package, importPath, typeName string) map[string]commentparsing.FieldDoc {
	t.Helper()
	for _, pkg := range pkgs {
		if pkg.ImportPath != importPath {
			continue
		}
		typ, ok := pkg.Type(typeName)
		require.True(t, ok, "%s.%s", importPath, typeName)
		return typ.Fields
	}
	t.Fatalf("%s was not discovered", importPath)
	return nil
}

func TestDiscoverAcrossPackages(t *testing.T) {
	// Only the root is named. upstream is reached through Config's embedded and nested fields.
	dir := filepath.Join("examples", "nested")
	pkgs, err := commentparsing.Discover(dir, &nested.Config{})
	require.NoError(t, err)

	t.Run("Reaches a second package without being told about it", func(t *testing.T) {
		var paths []string
		for _, pkg := range pkgs {
			paths = append(paths, pkg.ImportPath)
		}
		require.Equal(t, []string{examplesPath + "/nested", examplesPath + "/nested/upstream"}, paths)
	})

	// A generator's file paths are resolved against the run directory, so a package's directory is
	// reported the same way.
	t.Run("Directories are relative to the run directory", func(t *testing.T) {
		for _, pkg := range pkgs {
			require.False(t, filepath.IsAbs(pkg.Dir), pkg.ImportPath)
		}
		require.Equal(t, ".", pkgs[0].Dir)
		require.Equal(t, "upstream", pkgs[1].Dir)
	})

	t.Run("Documentation of an embedded type comes from its own package", func(t *testing.T) {
		settings := fields(t, pkgs, examplesPath+"/nested/upstream", "Settings")
		require.Equal(t, "Region is the deployment region.", settings["Region"].Comment)
	})

	t.Run("Documentation of a nested section comes from its own package", func(t *testing.T) {
		retry := fields(t, pkgs, examplesPath+"/nested/upstream", "Retry")
		require.Equal(t, "MaxAttempts is the total tries, including the first.", retry["MaxAttempts"].Comment)
	})

	// The package declares internal types too, and carrying them would tell a generator nothing
	// about which of them a config file can name.
	t.Run("Only the types the walk reached are carried", func(t *testing.T) {
		upstream := fields(t, pkgs, examplesPath+"/nested", "Config")
		require.Contains(t, upstream, "Endpoint")

		for _, pkg := range pkgs {
			if pkg.ImportPath == examplesPath+"/nested" {
				require.Len(t, pkg.Types, 1)
			}
		}
	})
}

// A duration is decoded from a single string, so demanding fields of it would fail on a dependency
// that has every right to document none.
func TestDiscoverSkipsScalarStructs(t *testing.T) {
	pkgs, err := commentparsing.Discover(filepath.Join("examples", "simple"), &simple.Config{})
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Equal(t, "Retries is the attempt count after the first.",
		fields(t, pkgs, examplesPath+"/simple", "Config")["Retries"].Comment)
}

// An external test package, so it can import the examples, which import the package under test.
// Discovery needs real non-test source to parse, and the examples are exactly that.
package commentparsing_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/simple"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

const examplesPath = "github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples"

// fields finds a discovered type's documentation the way a consumer keying on reflect.Type would.
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

// discovered runs the walk and hands back what a generator would have been given. The walk is
// not exported: a caller reaches it by passing a [commentparsing.Generator] to Files or Run, so
// that is how it is tested too.
func discovered(t *testing.T, dir string, roots ...any) []commentparsing.Package {
	t.Helper()

	var pkgs []commentparsing.Package
	_, err := commentparsing.Files(
		commentparsing.RunArgs{Roots: roots, Dir: dir, Tool: "discover_ext_test", Markup: tomlmarkup.New()},
		func(p []commentparsing.Package) (map[string]string, error) {
			pkgs = p
			return nil, nil
		},
	)
	require.NoError(t, err)
	return pkgs
}

func TestDiscoverAcrossPackages(t *testing.T) {
	// Only the root is named. upstream is reached through Config's embedded and nested fields.
	pkgs := discovered(t, filepath.Join("examples", "nested"), &nested.Config{})

	t.Run("Reaches a second package without being told about it", func(t *testing.T) {
		paths := make([]string, 0, len(pkgs))
		for _, pkg := range pkgs {
			paths = append(paths, pkg.ImportPath)
		}
		require.Equal(t, []string{examplesPath + "/nested", examplesPath + "/nested/upstream"}, paths)
	})

	// A generator's file paths are resolved against the run directory, so a package's
	// directory is reported the same way.
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

	// The package declares internal types too, which no config file can name.
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

// A duration is decoded from a single string, so it is a leaf value rather than a config section.
func TestDiscoverSkipsScalarStructs(t *testing.T) {
	pkgs := discovered(t, filepath.Join("examples", "simple"), &simple.Config{})
	require.Len(t, pkgs, 1)
	require.Equal(t, "Retries is the attempt count after the first.",
		fields(t, pkgs, examplesPath+"/simple", "Config")["Retries"].Comment)
}

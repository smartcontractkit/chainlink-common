package settings

import (
	"embed"
	"flag"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
)

var update = flag.Bool("update", false, "update the golden files of this test")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

//go:embed testdata/toml
var tomlFiles embed.FS

//go:embed testdata/config.toml
var configTOML string

//go:generate go test -run TestCombineTOMLFiles -update
func TestCombineTOMLFiles(t *testing.T) {
	sub, err := fs.Sub(tomlFiles, "testdata/toml")
	require.NoError(t, err)
	b, err := CombineTOMLFiles(sub)
	require.NoError(t, err)
	if *update {
		require.NoError(t, os.WriteFile("testdata/config.toml", b, os.ModePerm))
		return
	}
	require.Equal(t, configTOML, string(b))
}

func Test_tomlSettings_GetScoped(t *testing.T) {
	s, err := newTOMLSettings([]byte(configTOML))
	require.NoError(t, err)
	r := tomlGetter{s}

	ctx := contexts.WithCRE(t.Context(), contexts.CRE{
		Org:      "123",
		Owner:    "0x8bd112d3f8f92e41c861939545ad387307af9703",
		Workflow: "15c631d295ef5e32deb99a10ee6804bc4af1385568f9b3363f6552ac6dbb2cef",
	})
	gotValue, err := r.GetScoped(ctx, ScopeGlobal, `Foo`)
	require.NoError(t, err)
	assert.Equal(t, "5", gotValue)

	gotValue, err = r.GetScoped(ctx, ScopeGlobal, "Bar.Baz")
	require.NoError(t, err)
	assert.Equal(t, "10", gotValue)

	gotValue, err = r.GetScoped(ctx, ScopeOrg, "Foo")
	require.NoError(t, err)
	assert.Equal(t, "42", gotValue)

	gotValue, err = r.GetScoped(ctx, ScopeOrg, "Bar.Baz")
	require.NoError(t, err)
	assert.Equal(t, "99", gotValue)

	gotValue, err = r.GetScoped(ctx, ScopeOwner, "Foo")
	require.NoError(t, err)
	assert.Equal(t, "13", gotValue)

	gotValue, err = r.GetScoped(ctx, ScopeOwner, "Bar.Baz")
	require.NoError(t, err)
	assert.Equal(t, "43", gotValue)

	gotValue, err = r.GetScoped(ctx, ScopeWorkflow, "Foo")
	require.NoError(t, err)
	assert.Equal(t, "20", gotValue)

	gotValue, err = r.GetScoped(ctx, ScopeWorkflow, "Bar.Baz")
	require.NoError(t, err)
	assert.Equal(t, "50", gotValue)
}

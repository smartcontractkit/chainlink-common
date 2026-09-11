package flags

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

// --- registration errors ---

func TestNilTargetIsRejected(t *testing.T) {
	require.ErrorContains(t, RegisterCommandFlags(newRoot(t), nil, DefaultTOMLOptions("TEST")), "target cannot be nil")
}

func TestNilPointerTargetIsRejected(t *testing.T) {
	type cfg struct {
		Host string `toml:"host"`
	}

	var c *cfg
	require.ErrorContains(t, RegisterCommandFlags(newRoot(t), c, DefaultTOMLOptions("TEST")), "cannot be nil")
}

func TestNonStructTargetIsRejected(t *testing.T) {
	host := "example.com"
	require.ErrorContains(t, RegisterCommandFlags(newRoot(t), &host, DefaultTOMLOptions("TEST")), "must be a struct")
}

// --- flag metadata ---

// The `usage` tag is the flag's help text, which is the only place a setting explains itself now
// that the flags carry no generated documentation.
func TestUsageTagBecomesTheFlagsHelpText(t *testing.T) {
	type cfg struct {
		Host string `toml:"host" usage:"the host to dial"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{Host: "example.com"}, DefaultTOMLOptions("TEST")))

	f := root.PersistentFlags().Lookup("host")
	require.NotNil(t, f)
	assert.Equal(t, "the host to dial", f.Usage)
	assert.Equal(t, "example.com", f.DefValue)
}

// A root registration binds persistent flags, so a subcommand can be given a root setting; a
// subcommand registration binds local ones, which stay off its siblings.
func TestRootBindsPersistentFlagsAndSubcommandBindsLocalOnes(t *testing.T) {
	type rootCfg struct {
		Host string `toml:"host"`
	}
	type subCfg struct {
		Retries int `toml:"retries"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &rootCfg{}, DefaultTOMLOptions("TEST")))

	sub := &cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }}
	require.NoError(t, RegisterSubcommandFlags(sub, "sub", &subCfg{}, DefaultTOMLOptions()))
	root.AddCommand(sub)

	assert.NotNil(t, root.PersistentFlags().Lookup("host"))
	assert.Nil(t, root.PersistentFlags().Lookup("retries"))
	assert.NotNil(t, sub.Flags().Lookup("retries"))
	assert.Nil(t, sub.LocalFlags().Lookup("host"))
}

// opts.Namespace on a subcommand prefixes its flags too, which is what lets two dependencies
// registered on one subcommand keep same-named settings apart.
func TestSubcommandNamespaceCanPrefixTheFlagsToo(t *testing.T) {
	type aCfg struct {
		URL string `toml:"url"`
	}
	type bCfg struct {
		URL string `toml:"url"`
	}

	root := newRoot(t)
	sub := &cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }}

	aOpts := DefaultTOMLOptions("TEST")
	aOpts.Namespace = "a"
	bOpts := DefaultTOMLOptions("TEST")
	bOpts.Namespace = "b"

	var a aCfg
	var b bCfg
	require.NoError(t, RegisterSubcommandFlags(sub, "a", &a, aOpts))
	require.NoError(t, RegisterSubcommandFlags(sub, "b", &b, bOpts))
	root.AddCommand(sub)

	root.SetArgs([]string{"sub", "--a.url", "https://a", "--b.url", "https://b"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "https://a", a.URL)
	assert.Equal(t, "https://b", b.URL)
}

// --- config file discovery ---

// With no --config, a config.toml in the working directory is read, so a binary can be run
// without naming its config every time.
func TestConfigFileIsDiscoveredInTheWorkingDirectory(t *testing.T) {
	type cfg struct {
		Host string `toml:"host"`
	}

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte("host = 'from-cwd'\n"), 0o600))
	t.Chdir(dir)

	var c cfg
	require.NoError(t, run(t, &c))
	assert.Equal(t, "from-cwd", c.Host)
}

// A missing config.toml is not an error: everything it could have supplied has a flag, an env var
// and a compiled-in default.
func TestMissingDiscoveredConfigFileIsNotAnError(t *testing.T) {
	type cfg struct {
		Host string `toml:"host"`
	}

	t.Chdir(t.TempDir())

	c := cfg{Host: "default"}
	require.NoError(t, run(t, &c))
	assert.Equal(t, "default", c.Host)
}

// A config file named explicitly is different: naming one that isn't there is a mistake worth
// reporting rather than silently running on defaults.
func TestExplicitConfigFileMustExist(t *testing.T) {
	type cfg struct {
		Host string `toml:"host"`
	}

	var c cfg
	require.ErrorContains(t, run(t, &c, "--config", filepath.Join(t.TempDir(), "nope.toml")), "failed to read specified config file")
}

func TestMalformedConfigFileIsReported(t *testing.T) {
	type cfg struct {
		Host string `toml:"host"`
	}

	var c cfg
	require.ErrorContains(t, run(t, &c, "--config", writeConfig(t, "host = \n")), "failed to read specified config file")
}

// Two commands can bind the same key without one's viper state clobbering the other's, which a
// single process-global viper.Viper could not guarantee.
func TestEachCommandKeepsItsOwnViperState(t *testing.T) {
	type subCfg struct {
		Value string `toml:"value"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &struct{}{}, DefaultTOMLOptions("TEST")))

	var foo, bar subCfg
	fooCmd := &cobra.Command{Use: "foo", RunE: func(*cobra.Command, []string) error { return nil }}
	barCmd := &cobra.Command{Use: "bar", RunE: func(*cobra.Command, []string) error { return nil }}
	require.NoError(t, RegisterSubcommandFlags(fooCmd, "foo", &foo, DefaultTOMLOptions()))
	require.NoError(t, RegisterSubcommandFlags(barCmd, "bar", &bar, DefaultTOMLOptions()))
	root.AddCommand(fooCmd, barCmd)

	root.SetArgs([]string{"foo", "--value", "mine"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "mine", foo.Value)
	assert.Empty(t, bar.Value)
}

// --- pkg/config types ---

// A type whose text form lives on its pointer receiver still gets a readable default: %v on a
// config.URL value is its struct fields, which nothing could read back.
func TestPkgConfigURLBindsThroughItsTextForm(t *testing.T) {
	type cfg struct {
		Endpoint *config.URL `toml:"endpoint" usage:"where to dial"`
	}

	root := newRoot(t)
	c := cfg{Endpoint: config.MustParseURL("https://example.com/rpc")}
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))
	assert.Equal(t, "https://example.com/rpc", root.PersistentFlags().Lookup("endpoint").DefValue)

	var set cfg
	require.NoError(t, run(t, &set, "--endpoint", "https://other.example/rpc"))
	require.NotNil(t, set.Endpoint)
	assert.Equal(t, "https://other.example/rpc", set.Endpoint.String())
}

func TestPkgConfigSizeDecodesFromItsSuffixForm(t *testing.T) {
	type cfg struct {
		MaxBody config.Size `toml:"max-body"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--max-body", "10mb"))
	assert.Equal(t, config.Size(10*1000*1000), c.MaxBody)
}

// A secret that marshals itself redacted keeps that redaction in the flag's default, so --help
// can't leak a compiled-in value.
func TestPkgConfigSecretURLDefaultIsRedacted(t *testing.T) {
	type cfg struct {
		Endpoint *config.SecretURL `toml:"endpoint"`
	}

	root := newRoot(t)
	c := cfg{Endpoint: config.MustSecretURL("https://user:pw@example.com")}
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))
	assert.NotContains(t, root.PersistentFlags().Lookup("endpoint").DefValue, "example.com")

	var set cfg
	require.NoError(t, run(t, &set, "--endpoint", "https://other.example"))
	require.NotNil(t, set.Endpoint)
	assert.Equal(t, "https://other.example", (*config.URL)(set.Endpoint).String())
}

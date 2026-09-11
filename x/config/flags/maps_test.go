package flags

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

// --- maps ---

func TestStringMapFromFlag(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", "env=prod;region=us"))
	assert.Equal(t, map[string]string{"env": "prod", "region": "us"}, c.Labels)
}

func TestStringMapFromRepeatedFlag(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", "env=prod", "--labels", "region=us"))
	assert.Equal(t, map[string]string{"env": "prod", "region": "us"}, c.Labels)
}

func TestStringMapFromEnv(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	t.Setenv("TEST_LABELS", "env=prod;region=us")

	var c cfg
	require.NoError(t, run(t, &c))
	assert.Equal(t, map[string]string{"env": "prod", "region": "us"}, c.Labels)
}

func TestStringMapFromConfigFile(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	path := writeConfig(t, "[labels]\nenv = 'prod'\nregion = 'us'\n")

	var c cfg
	require.NoError(t, run(t, &c, "--config", path))
	assert.Equal(t, map[string]string{"env": "prod", "region": "us"}, c.Labels)
}

// A map's values are converted to the map's own value type, not left as the strings a flag or
// env var delivers them as.
func TestMapOfPrimitiveValues(t *testing.T) {
	type cfg struct {
		Counts  map[string]int     `toml:"counts"`
		Enabled map[string]bool    `toml:"enabled"`
		Weights map[string]float64 `toml:"weights"`
	}

	var c cfg
	require.NoError(t, run(t, &c,
		"--counts", "a=3;b=4",
		"--enabled", "a=true;b=false",
		"--weights", "a=1.5",
	))
	assert.Equal(t, map[string]int{"a": 3, "b": 4}, c.Counts)
	assert.Equal(t, map[string]bool{"a": true, "b": false}, c.Enabled)
	assert.Equal(t, map[string]float64{"a": 1.5}, c.Weights)
}

// Non-string keys are converted the same way the values are.
func TestMapWithPrimitiveKeys(t *testing.T) {
	type cfg struct {
		Chains map[uint32]string `toml:"chains"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--chains", "1=mainnet;11155111=sepolia"))
	assert.Equal(t, map[uint32]string{1: "mainnet", 11155111: "sepolia"}, c.Chains)
}

// A value that unmarshals itself from text goes through its own UnmarshalText, from every
// source - so config.Duration keeps parsing "45s" and keeps rejecting a negative duration.
func TestMapOfTextUnmarshalerValues(t *testing.T) {
	type cfg struct {
		Timeouts map[string]config.Duration `toml:"timeouts"`
	}

	t.Run("flag", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--timeouts", "read=45s;write=1m30s"))
		assert.Equal(t, 45*time.Second, c.Timeouts["read"].Duration())
		assert.Equal(t, 90*time.Second, c.Timeouts["write"].Duration())
	})

	t.Run("env", func(t *testing.T) {
		t.Setenv("TEST_TIMEOUTS", "read=45s")

		var c cfg
		require.NoError(t, run(t, &c))
		assert.Equal(t, 45*time.Second, c.Timeouts["read"].Duration())
	})

	t.Run("config file", func(t *testing.T) {
		path := writeConfig(t, "[timeouts]\nread = '45s'\n")

		var c cfg
		require.NoError(t, run(t, &c, "--config", path))
		assert.Equal(t, 45*time.Second, c.Timeouts["read"].Duration())
	})

	t.Run("invalid", func(t *testing.T) {
		var c cfg
		require.ErrorContains(t, run(t, &c, "--timeouts", "read=-45s"), "cannot make negative time duration")
	})
}

func TestMapOfDurationValues(t *testing.T) {
	type cfg struct {
		Timeouts map[string]time.Duration `toml:"timeouts"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--timeouts", "read=45s"))
	assert.Equal(t, 45*time.Second, c.Timeouts["read"])
}

// A map left alone keeps whatever the caller's struct was constructed with, and its compiled-in
// entries show up as the flag's default rather than as a Go-printed "map[...]".
func TestMapDefaultUntouched(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	c := cfg{Labels: map[string]string{"env": "dev"}}
	require.NoError(t, run(t, &c))
	assert.Equal(t, map[string]string{"env": "dev"}, c.Labels)

	root := newRoot(t)
	c2 := cfg{Labels: map[string]string{"env": "dev"}}
	require.NoError(t, RegisterCommandFlags(root, &c2, DefaultTOMLOptions("TEST")))
	f := root.PersistentFlags().Lookup("labels")
	require.NotNil(t, f)
	assert.Equal(t, "key=value;...", f.Value.Type())
	assert.Equal(t, "[env=dev]", f.DefValue)
}

// A flag overrides the config file per key, not per map: setting --labels replaces the table.
func TestMapFlagOverridesConfigFile(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	path := writeConfig(t, "[labels]\nenv = 'prod'\nregion = 'us'\n")

	var c cfg
	require.NoError(t, run(t, &c, "--config", path, "--labels", "env=staging"))
	assert.Equal(t, map[string]string{"env": "staging"}, c.Labels)
}

// A map whose values can't be written as a single string has no flag form, so no flag is
// registered for it - one bound anyway would default to the string "map[]" and fail to decode.
// The config file still fills it in.
func TestMapOfStructsIsConfigFileOnly(t *testing.T) {
	type chain struct {
		RPC     string `toml:"rpc"`
		Enabled bool   `toml:"enabled"`
	}
	type cfg struct {
		Chains map[string]chain `toml:"chains"`
	}

	root := newRoot(t)
	var c cfg
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))
	assert.Nil(t, root.PersistentFlags().Lookup("chains"))

	path := writeConfig(t, "[chains.eth]\nrpc = 'http://x'\nenabled = true\n")
	require.NoError(t, run(t, &c, "--config", path))
	assert.Equal(t, map[string]chain{"eth": {RPC: "http://x", Enabled: true}}, c.Chains)
}

// A list value's elements are separated by a comma, which is why entries are separated by a
// semicolon rather than by the comma pflag's own StringToString uses.
func TestMapOfLists(t *testing.T) {
	type cfg struct {
		URLs map[string][]string `toml:"urls"`
	}

	want := map[string][]string{"primary": {"a", "b"}, "backup": {"c"}}

	t.Run("flag", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--urls", "primary=a,b;backup=c"))
		assert.Equal(t, want, c.URLs)
	})

	t.Run("repeated flag", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--urls", "primary=a,b", "--urls", "backup=c"))
		assert.Equal(t, want, c.URLs)
	})

	// AWS CLI shorthand brackets its list values; accepted so a list reads as one.
	t.Run("bracketed", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--urls", "primary=[a,b];backup=[c]"))
		assert.Equal(t, want, c.URLs)
	})

	t.Run("env", func(t *testing.T) {
		t.Setenv("TEST_URLS", "primary=a,b;backup=c")

		var c cfg
		require.NoError(t, run(t, &c))
		assert.Equal(t, want, c.URLs)
	})

	t.Run("config file", func(t *testing.T) {
		path := writeConfig(t, "[urls]\nprimary = ['a', 'b']\nbackup = ['c']\n")

		var c cfg
		require.NoError(t, run(t, &c, "--config", path))
		assert.Equal(t, want, c.URLs)
	})
}

// A list of a non-string primitive converts element by element, the same as a scalar value does.
func TestMapOfPrimitiveLists(t *testing.T) {
	type cfg struct {
		Ports    map[string][]int             `toml:"ports"`
		Timeouts map[string][]config.Duration `toml:"timeouts"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--ports", "a=1,2;b=3", "--timeouts", "a=1s,2s"))
	assert.Equal(t, map[string][]int{"a": {1, 2}, "b": {3}}, c.Ports)
	require.Len(t, c.Timeouts["a"], 2)
	assert.Equal(t, time.Second, c.Timeouts["a"][0].Duration())
	assert.Equal(t, 2*time.Second, c.Timeouts["a"][1].Duration())
}

func TestMapOfListsDefault(t *testing.T) {
	type cfg struct {
		URLs map[string][]string `toml:"urls"`
	}

	root := newRoot(t)
	c := cfg{URLs: map[string][]string{"primary": {"a", "b"}}}
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))
	f := root.PersistentFlags().Lookup("urls")
	require.NotNil(t, f)
	assert.Equal(t, "key=v1,v2;...", f.Value.Type())
	assert.Equal(t, "[primary=a,b]", f.DefValue)
}

// A map of []byte is text, not a list of numbers, so it gets no flag rather than one that
// decodes its value into bytes one comma-separated element at a time.
func TestMapOfBytesIsConfigFileOnly(t *testing.T) {
	type cfg struct {
		Keys map[string][]byte `toml:"keys"`
	}

	root := newRoot(t)
	var c cfg
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))
	assert.Nil(t, root.PersistentFlags().Lookup("keys"))
}

// The comma-separated form pflag's StringToString and kubectl use is the likeliest mistake
// here, and it would otherwise be stored whole as one value.
func TestMapFlagRejectsCommaSeparatedPairs(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.ErrorContains(t, run(t, &c, "--labels", "env=prod,region=us"), `separate map entries with ";"`)
}

// Quoting is the way out for a scalar value that really does contain a comma.
func TestMapFlagQuotedValueKeepsComma(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", `env="a,b=c";region=us`))
	assert.Equal(t, map[string]string{"env": "a,b=c", "region": "us"}, c.Labels)
}

// A quoted value survives the round trip out through the flag's own text and back, which is
// the trip viper makes for every flag type it doesn't parse itself.
func TestMapFlagQuotedValueKeepsSeparator(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", `q="a;b";region=us`))
	assert.Equal(t, map[string]string{"q": "a;b", "region": "us"}, c.Labels)
}

// A comma in a scalar value is only rejected when it looks like another pair.
func TestMapFlagAllowsPlainCommaValue(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", "env=a,b"))
	assert.Equal(t, map[string]string{"env": "a,b"}, c.Labels)
}

func TestMapFlagRejectsMalformedPair(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.Error(t, run(t, &c, "--labels", "env"))
}

func TestMapEnvRejectsMalformedPair(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	t.Setenv("TEST_LABELS", "env")

	var c cfg
	require.ErrorContains(t, run(t, &c), "key=value")
}

func TestNamespacedMapFlag(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	root := newRoot(t)
	var c cfg
	opts := DefaultTOMLOptions("TEST")
	opts.Namespace = "db"
	require.NoError(t, RegisterCommandFlags(root, &c, opts))
	root.SetArgs([]string{"--db.labels", "env=prod"})
	require.NoError(t, root.Execute())
	assert.Equal(t, map[string]string{"env": "prod"}, c.Labels)
}

// --- config.Duration ---

func TestConfigDurationDecodes(t *testing.T) {
	type cfg struct {
		Timeout config.Duration `toml:"timeout"`
	}

	c := cfg{Timeout: *config.MustNewDuration(5 * time.Second)}
	require.NoError(t, run(t, &c, "--timeout", "45s"))
	assert.Equal(t, 45*time.Second, c.Timeout.Duration())
}

func TestConfigDurationRejectsNegative(t *testing.T) {
	type cfg struct {
		Timeout config.Duration `toml:"timeout"`
	}

	// pflag accepts -5s as a duration; config.Duration is what refuses it.
	t.Setenv("TEST_TIMEOUT", "-5s")

	var c cfg
	require.ErrorContains(t, run(t, &c), "negative")
}

// --- string coercion applies to every Options, not just DefaultTOMLOptions ---

// upperString is a minimal encoding.TextUnmarshaler, standing in for any caller-defined type
// with its own text decoding - distinct from config.Duration, which the rest of the suite
// already covers.
type upperString string

func (u *upperString) UnmarshalText(text []byte) error {
	*u = upperString(strings.ToUpper(string(text)))
	return nil
}

// coercionCfg exercises every field kind bindLeafFlag special-cases: plain string, an int and a
// uint kind, bool, time.Duration, config.Duration (a TextUnmarshaler struct), a string slice, and
// a caller-defined TextUnmarshaler.
type coercionCfg struct {
	Str     string
	Count   int
	Limit   uint
	Enabled bool
	Timeout time.Duration
	Retry   config.Duration
	Tags    []string
	Label   upperString
}

func assertCoercionCfg(t *testing.T, c *coercionCfg) {
	t.Helper()
	assert.Equal(t, "hello", c.Str)
	assert.Equal(t, 42, c.Count)
	assert.Equal(t, uint(7), c.Limit)
	assert.True(t, c.Enabled)
	assert.Equal(t, 45*time.Second, c.Timeout)
	assert.Equal(t, 5*time.Second, c.Retry.Duration())
	assert.Equal(t, []string{"a", "b"}, c.Tags)
	assert.Equal(t, upperString("SHOUT"), c.Label)
}

func TestCoercion_EveryFieldKindFromEnv_WithBareOptions(t *testing.T) {
	t.Setenv("TEST_STR", "hello")
	t.Setenv("TEST_COUNT", "42")
	t.Setenv("TEST_LIMIT", "7")
	t.Setenv("TEST_ENABLED", "true")
	t.Setenv("TEST_TIMEOUT", "45s")
	t.Setenv("TEST_RETRY", "5s")
	t.Setenv("TEST_TAGS", "a,b")
	t.Setenv("TEST_LABEL", "shout")

	var c coercionCfg
	require.NoError(t, runWithOptions(t, &c, Options{Prefixes: []string{"TEST"}}))
	assertCoercionCfg(t, &c)
}

func TestCoercion_EveryFieldKindFromFlag_WithBareOptions(t *testing.T) {
	var c coercionCfg
	require.NoError(t, runWithOptions(t, &c, Options{Prefixes: []string{"TEST"}},
		"--str", "hello",
		"--count", "42",
		"--limit", "7",
		"--enabled",
		"--timeout", "45s",
		"--retry", "5s",
		"--tags", "a",
		"--tags", "b",
		"--label", "shout",
	))
	assertCoercionCfg(t, &c)
}

func TestCoercion_EveryFieldKindFromConfigFile_WithBareOptions(t *testing.T) {
	path := writeConfig(t, `
str = "hello"
count = 42
limit = 7
enabled = true
timeout = "45s"
retry = "5s"
tags = ["a", "b"]
label = "shout"
`)

	var c coercionCfg
	require.NoError(t, runWithOptions(t, &c, Options{Prefixes: []string{"TEST"}}, "--config", path))
	assertCoercionCfg(t, &c)
}

func TestCoercion_InvalidValueStillErrors_WithBareOptions(t *testing.T) {
	t.Setenv("TEST_COUNT", "not-a-number")

	var c coercionCfg
	err := runWithOptions(t, &c, Options{Prefixes: []string{"TEST"}})
	require.ErrorContains(t, err, "Count")
}

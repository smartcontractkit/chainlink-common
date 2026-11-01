package cli

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chain struct {
	RPC    string
	WS     string
	Labels map[string]string
}

type hasChains struct {
	Chains map[string]chain
}

func TestMapEntriesAreSetLikeFields(t *testing.T) {
	file := "[Chains.mainnet]\nWS = 'file'\n"
	for _, tc := range everySource([]string{"--chains.mainnet.rpc", "set"}, "set", "") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasChains
			require.NoError(t, run(t, &c, testOptions, supplyConfig(t, "TEST_CHAINS_mainnet_RPC", tc.env, file, tc.args...)...))
			assert.Equal(t, map[string]chain{"mainnet": {RPC: "set", WS: "file"}}, c.Chains)
		})
	}

	t.Setenv("TEST_CHAINS_mainnet_RPC", "env")
	var c hasChains
	require.NoError(t, run(t, &c, testOptions, "--chains.mainnet.rpc", "flag"))
	assert.Equal(t, map[string]chain{"mainnet": {RPC: "flag"}}, c.Chains)

	var pointerMap struct{ Map *map[string]string }
	require.NoError(t, run(t, &pointerMap, testOptions, "--map.a", "1"))
	assert.Equal(t, map[string]string{"a": "1"}, *pointerMap.Map)

	shared := &chain{WS: "default"}
	pointers := struct{ Chains map[string]*chain }{map[string]*chain{"a": shared}}
	require.NoError(t, run(t, &pointers, testOptions, "--chains.a.rpc", "flag"))
	assert.Equal(t, &chain{RPC: "flag", WS: "default"}, pointers.Chains["a"])
	assert.Equal(t, &chain{WS: "default"}, shared)
}

func TestMapEntriesNest(t *testing.T) {
	var c hasChains
	require.NoError(t, run(t, &c, testOptions, "--chains.a.labels", "x=1", "--chains.a.labels.y", "2"))
	assert.Equal(t, map[string]chain{"a": {Labels: map[string]string{"x": "1", "y": "2"}}}, c.Chains)

	var ptrs struct{ Ptrs map[string]*int }
	require.NoError(t, run(t, &ptrs, testOptions, "--ptrs.a", "1"))
	assert.Equal(t, 1, *ptrs.Ptrs["a"])

	var maps struct{ Maps map[string]map[string]string }
	require.NoError(t, run(t, &maps, testOptions, "--maps.a", "x.y=1", "--maps.b.c", "2"))
	assert.Equal(t, map[string]map[string]string{"a": {"x.y": "1"}, "b": {"c": "2"}}, maps.Maps)
	require.ErrorContains(t, run(t, &maps, testOptions, "--maps.a.x.y", "v"),
		`maps.a.x.y could be maps.<key>.<key> with keys ["a" "x.y"] or maps.<key>.<key> with keys ["a.x" "y"]`)
}

func TestMapKeysAreAsWritten(t *testing.T) {
	t.Setenv("TEST_MAP_lower", "env")
	t.Setenv("TEST_MAP_MixedCase", "env")
	t.Setenv("TEST_MAP_node_id", "env")
	t.Setenv("TEST_MAP_service.name", "env")
	var c hasStringMap
	require.NoError(t, run(t, &c, testOptions, "--map.Flag_Case", "flag", "--map.service.name", "flag"))
	assert.Equal(t, map[string]string{"lower": "env", "MixedCase": "env", "node_id": "env", "service.name": "flag", "Flag_Case": "flag"},
		c.Map)

	t.Setenv("TEST_MAP_", "env")
	t.Setenv("TEST_CHAINS_a_b_RPC", "env")
	t.Setenv("TEST_CHAINS_a_LABELS_x_y", "env")
	var chains struct {
		Chains map[string]chain
		Map    map[string]string
	}
	require.NoError(t, run(t, &chains, testOptions, "--chains.c.d.labels.e.f", "flag"))
	assert.Equal(t, map[string]chain{
		"a_b": {RPC: "env"},
		"a":   {Labels: map[string]string{"x_y": "env"}},
		"c.d": {Labels: map[string]string{"e.f": "flag"}},
	}, chains.Chains)
	assert.Equal(t, map[string]string{"lower": "env", "MixedCase": "env", "node_id": "env", "service.name": "env"}, chains.Map)
}

func TestMapKeysThatReadTwoWaysAreRejected(t *testing.T) {
	t.Setenv("TEST_MAPS_a_b_c", "env")
	require.ErrorContains(t, run(t, &struct{ Maps map[string]map[string]string }{}, testOptions),
		`app: TEST_MAPS_a_b_c could be TEST_MAPS_<KEY>_<KEY> with keys ["a" "b_c"] or TEST_MAPS_<KEY>_<KEY> with keys ["a_b" "c"]`)
}

func TestMapEntryEnvVars(t *testing.T) {
	t.Setenv("TEST_MAP_a", "")
	c := hasStringMap{Map: map[string]string{"a": "default"}}
	require.NoError(t, run(t, &c, testOptions))
	assert.Equal(t, map[string]string{"a": "default"}, c.Map)

	t.Setenv("TEST_UINT_KEYS_x", "a")
	require.ErrorContains(t, run(t, &struct{ UintKeys map[uint32]string }{}, testOptions),
		`TEST_UINT_KEYS_x: key "x": strconv.ParseUint: parsing "x": invalid syntax`)

	t.Setenv("TEST_INTS_a", "x")
	require.ErrorContains(t, run(t, &struct{ Ints map[string]int }{}, testOptions),
		`invalid value "x" for TEST_INTS_a: strconv.ParseInt: parsing "x": invalid syntax`)
}

func TestMapEntryEnvPrefixesTriedInOrder(t *testing.T) {
	t.Setenv("SECOND_MAP", "a=second")
	t.Setenv("FIRST_MAP_a", "first")
	t.Setenv("SECOND_MAP_a", "second")

	var c hasStringMap
	require.NoError(t, run(t, &c, Options{Markup: testOptions.Markup, Prefixes: []string{"FIRST", "SECOND"}}))
	assert.Equal(t, map[string]string{"a": "first"}, c.Map)
}

func TestMapEntrySetTwiceInOneSource(t *testing.T) {
	require.ErrorContains(t, run(t, &hasStringMap{}, testOptions, "--map", "a=1", "--map.a", "2"),
		`--map (key "a") and --map.a set the same value`)
	require.ErrorContains(t, run(t, &struct{ UintKeys map[uint32]string }{}, testOptions, "--uint-keys.0x10", "a", "--uint-keys.16", "b"),
		`--uint-keys.0x10 and --uint-keys.16 set the same value`)

	t.Setenv("TEST_MAP_a", "env")
	var c hasStringMap
	require.NoError(t, run(t, &c, testOptions, "--map", "a=flag"))
	assert.Equal(t, map[string]string{"a": "flag"}, c.Map)

	t.Setenv("TEST_MAP", "a=1")
	require.ErrorContains(t, run(t, &hasStringMap{}, testOptions), `TEST_MAP (key "a") and TEST_MAP_a set the same value`)
}

func TestMapEntryFlagErrors(t *testing.T) {
	require.ErrorContains(t, run(t, &struct{ UintKeys map[uint32]string }{}, testOptions, "--uint-keys.x", "a"),
		`invalid argument "a" for "--uint-keys.x" flag: key "x": strconv.ParseUint: parsing "x": invalid syntax`)
	require.ErrorContains(t, run(t, &hasStringMap{}, testOptions, "--map.<key>", "a"),
		`invalid argument "a" for "--map.<key>" flag: replace <key> with a map key`)
	require.ErrorContains(t, run(t, &hasStringMap{}, testOptions, "--other.a", "a"), "unknown flag: --other.a")
}

func TestMapEntryHelp(t *testing.T) {
	flags := flagsOf(t, &hasChains{}, testOptions)
	assert.Equal(t, "[env TEST_CHAINS_<KEY>_RPC]", flags.Lookup("chains.<key>.rpc").Usage)
	assert.NotNil(t, flags.Lookup("chains.<key>.labels.<key>"))

	type nestsItself struct{ Children map[string]nestsItself }
	assert.NotNil(t, flagsOf(t, &struct{ Tree map[string]nestsItself }{}, testOptions))
}

func TestFieldEnvVarsWinOverMapEntries(t *testing.T) {
	type hasMapAndField struct {
		Map    map[string]string
		MapTTL string
	}

	t.Setenv("TEST_MAP_TTL", "field")
	t.Setenv("TEST_MAP_a", "entry")
	var c hasMapAndField
	require.NoError(t, run(t, &c, testOptions, "--config", writeConfig(t, "Map = {ttl = 'file'}\n")))
	assert.Equal(t, hasMapAndField{Map: map[string]string{"a": "entry", "ttl": "file"}, MapTTL: "field"}, c)
}

func TestMapEntryEnvVarsMustNotClash(t *testing.T) {
	type hasTwoMaps struct {
		Chains    map[string]chain
		ChainsRPC map[string]string
	}
	require.ErrorContains(t, run(t, &hasTwoMaps{}, testOptions), "chains-rpc.<key> and chains.<key>.rpc could both be TEST_CHAINS_RPC_RPC")
}

func TestOwnNormalizationFuncStillApplies(t *testing.T) {
	root := newRoot(t)
	root.SetGlobalNormalizationFunc(func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
	})

	var c hasChains
	bind(t, root, &c, testOptions)
	root.SetArgs([]string{"--chains.a_b.rpc", "x"})
	require.NoError(t, root.Execute())
	assert.Equal(t, map[string]chain{"a-b": {RPC: "x"}}, c.Chains)
}

func TestFlagsMustNotShadowMapEntries(t *testing.T) {
	root := newRoot(t)
	bind(t, root, &hasStringMap{}, testOptions)
	root.Flags().String("map.a", "", "")
	require.ErrorContains(t, root.Execute(), "app: flag --map.a would set --map.<key>; rename it")
}

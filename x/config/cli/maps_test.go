package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

type hasStringMap struct {
	Map map[string]string
}

func TestStringMapFromEverySource(t *testing.T) {
	cases := everySource([]string{"--map", "a=1,b=2"}, "a=1,b=2", "[Map]\na = '1'\nb = '2'")
	cases = append(cases, sourceCase{name: "repeated flag", args: []string{"--map", "a=1", "--map", "b=2"}})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c hasStringMap
			require.NoError(t, run(t, &c, testOptions, supply(t, "TEST_MAP", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, map[string]string{"a": "1", "b": "2"}, c.Map)
		})
	}
}

func TestMapsMergeAcrossSources(t *testing.T) {
	t.Setenv("TEST_MAP", "e=env,a=env")
	c := hasStringMap{Map: map[string]string{"d": "default", "a": "default"}}
	require.NoError(t, run(t, &c, testOptions,
		"--config", writeConfig(t, "[Map]\na = 'first'\nf = 'first'\n"),
		"--config", writeConfig(t, "[Map]\na = 'second'\ns = 'second'\n"),
		"--map", "a=flag"))
	assert.Equal(t, map[string]string{"d": "default", "f": "first", "s": "second", "e": "env", "a": "flag"}, c.Map)
}

func TestMapEntryText(t *testing.T) {
	var c hasStringMap
	require.NoError(t, run(t, &c, testOptions, "--map", `"a=1,b=2",c=3`))
	assert.Equal(t, map[string]string{"a": "1,b=2", "c": "3"}, c.Map)

	t.Setenv("TEST_MAP", "a")
	require.ErrorContains(t, run(t, &hasStringMap{}, testOptions), `invalid value "a" for TEST_MAP: "a" must be formatted as key=value`)
}

func TestMapKeysNeedNotBeStrings(t *testing.T) {
	type hasNonStringKeys struct {
		UintKeys map[uint32]string
		TextKeys map[config.Duration]string
	}
	want := hasNonStringKeys{
		UintKeys: map[uint32]string{1: "a", 42: "b"},
		TextKeys: map[config.Duration]string{*config.MustNewDuration(45 * time.Second): "c"},
	}

	flag := []string{"--uint-keys", "1=a,0x2a=b", "--text-keys", "45s=c"}
	for _, tc := range everySource(flag, "", "[UintKeys]\n1 = 'a'\n0x2a = 'b'\n[TextKeys]\n45s = 'c'\n") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasNonStringKeys
			require.NoError(t, run(t, &c, testOptions, supply(t, "", "", tc.file, tc.args...)...))
			assert.Equal(t, want, c)
		})
	}

	for args, want := range map[string]string{"16=a,0x10=b": `keys "0x10" and "16" are both 16`, "16=a --uint-keys 0x10=b": `keys "16" and "0x10" are both 16`} {
		require.ErrorContains(t, run(t, &hasNonStringKeys{}, testOptions, append([]string{"--uint-keys"}, strings.Fields(args)...)...), want, args)
	}
	require.ErrorContains(t, run(t, &hasNonStringKeys{}, testOptions, "--config", writeConfig(t, "[UintKeys]\n16 = 'a'\n0x10 = 'b'\n")),
		`keys "0x10" and "16" are both 16`)
}

func TestHelpDefaultsCanBePassedBackIn(t *testing.T) {
	type hasListMapAndText struct {
		List []string
		Map  map[string]string
		Text *config.URL
	}

	c := hasListMapAndText{List: []string{"a", "b,c"}, Map: map[string]string{"a": "1", "b": "2,3"}, Text: config.MustParseURL("https://x/rpc")}
	flags := flagsOf(t, &c, testOptions)
	assert.Equal(t, "value,...", flags.Lookup("list").Value.Type())
	assert.Equal(t, "key=value,...", flags.Lookup("map").Value.Type())

	var back hasListMapAndText
	args := []string{"--list", flags.Lookup("list").DefValue, "--map", flags.Lookup("map").DefValue,
		"--text", flags.Lookup("text").DefValue}
	require.NoError(t, run(t, &back, testOptions, args...))
	assert.Equal(t, c, back)
}

type intKindText time.Duration

func (d *intKindText) UnmarshalText(text []byte) error {
	parsed, err := time.ParseDuration(string(text))
	*d = intKindText(parsed)
	return err
}

func TestTextUnmarshalerLeafFromEverySource(t *testing.T) {
	type hasIntKindText struct {
		Value intKindText
	}

	for _, tc := range everySource([]string{"--value", "30m"}, "30m", "Value = '30m'") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasIntKindText
			require.NoError(t, run(t, &c, testOptions, supply(t, "TEST_VALUE", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, 30*time.Minute, time.Duration(c.Value))
		})
	}
	require.ErrorContains(t, run(t, &hasIntKindText{}, testOptions, "--value", "not-a-duration"), `time: invalid duration "not-a-duration"`)
}

func TestLeafNestedInASectionListAndMap(t *testing.T) {
	type hasDuration struct {
		Value config.Duration
	}
	type hasDurationEverywhere struct {
		Section hasDuration
		List    []config.Duration
		Map     map[string]config.Duration
	}
	want := hasDurationEverywhere{
		Section: hasDuration{*config.MustNewDuration(time.Minute)},
		List:    []config.Duration{*config.MustNewDuration(time.Second), *config.MustNewDuration(2 * time.Second)},
		Map:     map[string]config.Duration{"a": *config.MustNewDuration(3 * time.Second)},
	}

	flag := []string{"--section.value", "1m", "--list", "1s,2s", "--map", "a=3s"}
	for _, tc := range everySource(flag, "", "List = ['1s', '2s']\n[Section]\nValue = '1m'\n[Map]\na = '3s'") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasDurationEverywhere
			require.NoError(t, run(t, &c, testOptions, supply(t, "", "", tc.file, tc.args...)...))
			assert.Equal(t, want, c)
		})
	}
}

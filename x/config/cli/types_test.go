package cli

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hasEveryKind struct {
	Int          int
	Int8         int8
	Uint         uint
	Float32      float32
	Bool         bool
	String       string
	Duration     time.Duration
	IntList      []int
	IntMap       map[string]int
	UintKeyedMap map[uint32]string
}

// The markup decodes a config file, so its type rules hold: here, tomlmarkup's text is no number.
func TestConfigFileValuesAreNotConverted(t *testing.T) {
	require.ErrorContains(t, run(t, &hasEveryKind{}, testOptions, "--config", writeConfig(t, "Int = '8080'")),
		"line 1: cannot decode TOML string into int")
}

func TestMapKeysAreParsedAsTheKeyType(t *testing.T) {
	for _, tc := range everySource([]string{"--uint-keyed-map", "1=mainnet,0x2a=kovan"}, "1=mainnet,0x2a=kovan", "[UintKeyedMap]\n1 = 'mainnet'\n0x2a = 'kovan'") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasEveryKind
			require.NoError(t, run(t, &c, testOptions, supply(t, "TEST_UINT_KEYED_MAP", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, map[uint32]string{1: "mainnet", 42: "kovan"}, c.UintKeyedMap)
		})
	}
	require.ErrorContains(t, run(t, &hasEveryKind{}, testOptions, "--config", writeConfig(t, "[UintKeyedMap]\nx = 'mainnet'")),
		`key "x": strconv.ParseUint: parsing "x": invalid syntax`)
}

func TestParserErrorsArePropagated(t *testing.T) {
	err := run(t, &hasEveryKind{}, testOptions, "--int=abc")
	require.ErrorContains(t, err, `invalid argument "abc" for "--int" flag: strconv.ParseInt: parsing "abc": invalid syntax`)

	require.ErrorContains(t, run(t, &hasEveryKind{}, testOptions, "--int-list=1,x"), `element "x": strconv.ParseInt: parsing "x"`)
	require.ErrorContains(t, run(t, &hasEveryKind{}, testOptions, "--int-map=a=x"), `value of "a": strconv.ParseInt: parsing "x"`)

	t.Setenv("TEST_INT", "abc")
	require.ErrorContains(t, run(t, &hasEveryKind{}, testOptions), `invalid value "abc" for TEST_INT: strconv.ParseInt: parsing "abc": invalid syntax`)
}

// A field with no text form has no flag, so it has no env var either.
func TestEnvIsIgnoredForAConfigFileOnlyField(t *testing.T) {
	t.Setenv("TEST_LIST", "a")

	var c struct{ List []BasicConfig }
	require.NoError(t, run(t, &c, testOptions))
	assert.Nil(t, c.List)
}

func TestParseTextRejectsATypeWithNoTextForm(t *testing.T) {
	_, err := parseText(reflect.TypeFor[struct{}](), "x")
	require.ErrorContains(t, err, "has no text form")
}

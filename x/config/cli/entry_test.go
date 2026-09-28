package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

func TestPrecedence(t *testing.T) {
	type hasTwoStrings struct {
		Value     string
		Untouched string
	}

	for _, tc := range []struct {
		name, env, file string
		args            []string
		want            string
	}{
		{"nothing", "", "", nil, "default"},
		{"file", "", "Value = 'file'", nil, "file"},
		{"file and env", "env", "Value = 'file'", nil, "env"},
		{"file, env and flag", "env", "Value = 'file'", []string{"--value", "flag"}, "flag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TEST_VALUE", "")
			c := hasTwoStrings{Value: "default", Untouched: "untouched"}
			require.NoError(t, run(t, &c, testOptions, supply(t, "TEST_VALUE", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, hasTwoStrings{tc.want, "untouched"}, c)
		})
	}
}

func TestEnvPrefixesTriedInOrder(t *testing.T) {
	t.Setenv("FIRST_VALUE", "first")
	t.Setenv("SECOND_VALUE", "second")

	var c BasicConfig
	require.NoError(t, run(t, &c, Options{Markup: tomlmarkup.New(), Prefixes: []string{"FIRST", "SECOND"}}))
	assert.Equal(t, "first", c.Value)
}

func TestValidationErrorsUseConfigKeys(t *testing.T) {
	// Renamed keys show a config file key comes from the tag, not the Go name.
	type nestsRenamedRequired struct {
		GoName RequiredField `toml:"Renamed"`
	}
	type hasExclusivePointers struct {
		Excluded *RequiredField `validate:"excluded_with=Excluder"`
		Excluder *RequiredField `toml:"RenamedExcluder"`
	}

	for _, tc := range []struct {
		name, namespace string
		target          any
		args            []string
		want            string
	}{
		{"a leaf names how to set it", "app", &nestsRenamedRequired{}, nil, "invalid configuration: app.Renamed.Value failed on the 'required' tag; " +
			"set it with --app.renamed.value, TEST_APP_RENAMED_VALUE, app.Renamed.Value in a config file"},
		{"a section and its rule's sibling", "", &hasExclusivePointers{}, []string{"--excluded.value", "v", "--renamed-excluder.value", "v"},
			"invalid configuration: Excluded failed on the 'excluded_with=RenamedExcluder' tag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorContains(t, runIn(t, tc.namespace, tc.target, tc.args...), tc.want)
		})
	}
}

func TestValidationErrorsInsideAListOrMapUseConfigKeys(t *testing.T) {
	type hasTaggedRequired struct {
		GoName string `toml:"renamed" validate:"required"`
	}
	type hasSiblingRule struct {
		Rule    string `validate:"required_without=Sibling"`
		Sibling string `toml:"renamed"`
	}

	type embedsRequired struct {
		RequiredField
	}
	type hasValidatedCollections struct {
		Structs []hasTaggedRequired `validate:"dive"`
		Strings []string            `validate:"dive,required"`
		Map     map[string]string   `validate:"dive,required"`
		Rules   []hasSiblingRule    `validate:"dive"`
		Embeds  []embedsRequired    `validate:"dive"`
	}

	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"a list element's field, by its tag", []string{"--config", writeConfig(t, "[app]\n[[app.Structs]]\nrenamed = 'a'\n[[app.Structs]]\nrenamed = ''\n")},
			"app.Structs.1.renamed failed on the 'required' tag"},
		{"a list element", []string{"--app.strings", "a,"}, "app.Strings.1 failed on the 'required' tag"},
		{"a map entry", []string{"--app.map", "k="}, "app.Map.k failed on the 'required' tag"},
		{"a rule's sibling in the same element", []string{"--config", writeConfig(t, "[app]\n[[app.Rules]]\n")},
			"app.Rules.0.Rule failed on the 'required_without=app.Rules.0.renamed' tag"},
		{"a flattened embed adds no segment", []string{"--config", writeConfig(t, "[app]\n[[app.Embeds]]\n")},
			"app.Embeds.0.Value failed on the 'required' tag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := runIn(t, "app", &hasValidatedCollections{}, tc.args...)
			require.ErrorContains(t, err, tc.want)
			// Nothing falls back to the validator's own text, which names Go fields.
			assert.NotContains(t, err.Error(), "Key: ")
		})
	}
}

// `set` failing is also what shows the `validate` tags run at all.
func TestSetRequiresASource(t *testing.T) {
	type hasSetBool struct {
		Value bool `validate:"set"` //nolint:revive // set is the Binder's own rule
	}

	for _, tc := range everySource([]string{"--value=false"}, "false", "Value = false") {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, run(t, &hasSetBool{}, testOptions, supply(t, "TEST_VALUE", tc.env, tc.file, tc.args...)...))
		})
	}
	require.ErrorContains(t, run(t, &hasSetBool{}, testOptions), "Value failed on the 'set' tag")
}

func TestSetIgnoresADefault(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		type hasSetInt struct {
			Value int `validate:"set"` //nolint:revive // set is the Binder's own rule
		}

		require.ErrorContains(t, run(t, &hasSetInt{Value: 5}, testOptions), "Value failed on the 'set' tag")
		require.NoError(t, run(t, &hasSetInt{Value: 5}, testOptions, "--value", "5"))
	})

	t.Run("pointer", func(t *testing.T) {
		type hasSetIntPointer struct {
			Value *int `validate:"set"` //nolint:revive // set is the Binder's own rule
		}

		value := 5
		require.ErrorContains(t, run(t, &hasSetIntPointer{Value: &value}, testOptions), "Value failed on the 'set' tag")
		require.NoError(t, run(t, &hasSetIntPointer{Value: &value}, testOptions, "--value", "5"))
		require.ErrorContains(t, run(t, &hasSetIntPointer{}, testOptions), "Value failed on the 'set' tag")
		require.NoError(t, run(t, &hasSetIntPointer{}, testOptions, "--value", "5"))
	})
}

func TestSetInANestedSection(t *testing.T) {
	type nestsSetString struct {
		Section struct {
			Value string `validate:"set"` //nolint:revive // set is the Binder's own rule
		}
	}

	require.ErrorContains(t, run(t, &nestsSetString{}, testOptions), "Section.Value failed")
	require.NoError(t, run(t, &nestsSetString{}, testOptions, "--section.value", ""))
}

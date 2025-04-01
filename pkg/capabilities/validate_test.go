package capabilities_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type TestConfig struct {
	Foo  *values.List `json:"foo"`
	Bar  int          `json:"bar"`
	Bonk values.Map   `json:"bonk"`
}

type TestInputs struct {
	Baz string `json:"baz" jsonschema:"pattern=^world$"`
	Qux int    `json:"qux"`
}

type TestOutputs struct {
	Quux  string `json:"quux"`
	Corge int    `json:"corge" jsonschema:"minimum=1"`
}

func TestValidator_ConfigSchema(t *testing.T) {
	v := capabilities.NewValidator[TestConfig, TestInputs, TestOutputs](capabilities.ValidatorArgs{})
	schema, err := v.ConfigSchema()
	assert.NoError(t, err)
	assert.NotEmpty(t, schema)
}

func TestValidator_InputsSchema(t *testing.T) {
	v := capabilities.NewValidator[TestConfig, TestInputs, TestOutputs](capabilities.ValidatorArgs{})
	schema, err := v.InputsSchema()
	assert.NoError(t, err)
	assert.NotEmpty(t, schema)
}

func TestValidator_OutputsSchema(t *testing.T) {
	v := capabilities.NewValidator[TestConfig, TestInputs, TestOutputs](capabilities.ValidatorArgs{})
	schema, err := v.OutputsSchema()
	assert.NoError(t, err)
	assert.NotEmpty(t, schema)
}

func TestValidator_Schema(t *testing.T) {
	v := capabilities.NewValidator[TestConfig, TestInputs, TestOutputs](capabilities.ValidatorArgs{})
	schema, err := v.Schema()
	assert.NoError(t, err)
	assert.NotEmpty(t, schema)
}

func TestValidator_ValidateSchema(t *testing.T) {
	v, err := values.NewMap(
		map[string]interface{}{
			"feedIds": []string{"0x1111111111111111111100000000000000000000000000000000000000000000"},
		},
	)
	assert.NoError(t, err)
	schema := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/feed-ids",
  "$ref": "#/$defs/FeedIds",
  "$defs": {
    "FeedIds": {
      "properties": {
        "feedIds": {
          "items": {
            "type": "string",
            "pattern": "^0x[0-9a-f]{64}$"
          },
          "type": "array"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "feedIds"
      ]
    }
  }
}`

	result, err := capabilities.ValidateAgainstSchema(v, schema)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	v, err = values.NewMap(
		map[string]interface{}{
			"feedIds": []string{"0x111111111111111111110F000000000000000000000000000000000000000000"},
		},
	)
	assert.NoError(t, err)

	result, err = capabilities.ValidateAgainstSchema(v, schema)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestValidator_ValidateConfig(t *testing.T) {
	t.Parallel()
	m, err := values.NewMap(map[string]interface{}{
		"baz": "world",
	})
	assert.NoError(t, err)

	l, err := values.NewList([]interface{}{"hello", "world"})
	assert.NoError(t, err)

	v := capabilities.NewValidator[TestConfig, TestInputs, TestOutputs](capabilities.ValidatorArgs{})
	config, err := values.NewMap(
		map[string]interface{}{
			"foo":  l,
			"bar":  123,
			"bonk": m,
		},
	)
	assert.NoError(t, err)
	result, err := v.ValidateConfig(config)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 123, result.Bar)

	var str string
	err = result.Bonk.Underlying["baz"].UnwrapTo(&str)
	assert.NoError(t, err)
	assert.Equal(t, "world", str)

	var list []string
	err = result.Foo.UnwrapTo(&list)
	assert.NoError(t, err)
	assert.Equal(t, []string{"hello", "world"}, list)
}

func TestValidator_ValidateInputs(t *testing.T) {
	t.Parallel()
	v := capabilities.NewValidator[TestConfig, TestInputs, TestOutputs](capabilities.ValidatorArgs{})
	inputs, err := values.NewMap(
		map[string]interface{}{
			"baz": "world",
			"qux": 456,
		},
	)
	assert.NoError(t, err)
	result, err := v.ValidateInputs(inputs)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	inputs, err = values.NewMap(
		map[string]interface{}{
			"baz": "world",
			"qux": -1,
		},
	)
	assert.NoError(t, err)
	result, err = v.ValidateInputs(inputs)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	inputs, err = values.NewMap(
		map[string]interface{}{
			"baz": "worl",
			"qux": -1,
		},
	)
	assert.NoError(t, err)
	result, err = v.ValidateInputs(inputs)
	assert.ErrorContains(t, err, "does not match pattern '^world$'")
	assert.Nil(t, result)
}

func TestValidator_ValidateOutputs(t *testing.T) {
	v := capabilities.NewValidator[TestConfig, TestInputs, TestOutputs](capabilities.ValidatorArgs{})

	outputs, err := values.NewMap(
		map[string]interface{}{
			"quux":  "world",
			"corge": 456,
		},
	)
	assert.NoError(t, err)
	result, err := v.ValidateOutputs(outputs)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	outputs, err = values.NewMap(
		map[string]interface{}{
			"quux":  "world",
			"corge": 0,
		},
	)
	assert.NoError(t, err)
	result, err = v.ValidateOutputs(outputs)
	assert.ErrorContains(t, err, "must be >= 1")
	assert.Nil(t, result)
}

func TestValidator_GenerateSchema(t *testing.T) {
	t.Parallel()
	capInfo := capabilities.CapabilityInfo{
		ID:             "test@1.0.0",
		CapabilityType: capabilities.CapabilityTypeTrigger,
		Description:    "test description",
	}
	v := capabilities.NewValidator[TestConfig, TestInputs, TestOutputs](capabilities.ValidatorArgs{Info: capInfo})
	schema, err := v.Schema()
	assert.NoError(t, err)
	var shouldUpdate = false
	if shouldUpdate {
		err = os.WriteFile("./testdata/fixtures/validator/schema.json", []byte(schema), 0600)
		assert.NoError(t, err)
	}

	fixture, err := os.ReadFile("./testdata/fixtures/validator/schema.json")
	assert.NoError(t, err)

	if diff := cmp.Diff(fixture, []byte(schema), transformJSON); diff != "" {
		t.Errorf("TestValidator_GenerateSchema() mismatch (-want +got):\n%s", diff)
		t.FailNow()
	}

	// We allow inputs to be nil, since triggers do not have inputs
	triggerValidator := capabilities.NewValidator[TestConfig, any, TestOutputs](capabilities.ValidatorArgs{})
	schema, err = triggerValidator.Schema()
	assert.NoError(t, err)
	if shouldUpdate {
		err = os.WriteFile("./testdata/fixtures/validator/trigger_schema.json", []byte(schema), 0600)
		assert.NoError(t, err)
	}

	fixture, err = os.ReadFile("./testdata/fixtures/validator/trigger_schema.json")

	assert.NoError(t, err)

	if diff := cmp.Diff(fixture, []byte(schema), transformJSON); diff != "" {
		t.Errorf("TestValidator_GenerateSchema() mismatch (-want +got):\n%s", diff)
		t.FailNow()
	}

	// We dont allow other permutations of nil types
	// Since we don't have the need for them currently
	invalidValidator1 := capabilities.NewValidator[TestConfig, any, any](capabilities.ValidatorArgs{})
	_, err = invalidValidator1.Schema()
	assert.Error(t, err)

	invalidValidator2 := capabilities.NewValidator[any, TestInputs, any](capabilities.ValidatorArgs{})
	_, err = invalidValidator2.Schema()
	assert.Error(t, err)

	invalidValidator3 := capabilities.NewValidator[any, any, TestOutputs](capabilities.ValidatorArgs{})
	_, err = invalidValidator3.Schema()
	assert.Error(t, err)
}

var transformJSON = cmp.FilterValues(func(x, y []byte) bool {
	return json.Valid(x) && json.Valid(y)
}, cmp.Transformer("ParseJSON", func(in []byte) (out interface{}) {
	if err := json.Unmarshal(in, &out); err != nil {
		panic(err) // should never occur given previous filter to ensure valid JSON
	}
	return out
}))

func TestCreateCapabilityIDRegex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		capabilityName string
		version        *semver.Version
		testIDs        map[string]bool // ID to expected match result
	}{
		{
			name:           "streams-trigger version 1.0.0",
			capabilityName: "streams-trigger",
			version:        mustSemVer(t, "1"),
			testIDs: map[string]bool{
				"streams-trigger@1.0.0":                              true,  // No labels
				"streams-trigger@1.77.0":                             true,  // Minor version OK
				"streams-trigger@1.0.13":                             true,  // Patch version OK
				"streams-trigger:DON_trigger@1.0.0":                  true,  // One label
				"streams-trigger:DON_trigger:format_json@1.0.0":      true,  // Two labels
				"streams-trigger:DON-trigger:format_json@1.0.0":      true,  // Hyphen in label
				"streams-trigger:a_b:c_d:e_f@1.0.0":                  true,  // Three labels
				"streamstrigger@1.0.0":                               false, // Wrong name
				"streams-trigger@2.0.0":                              false, // Wrong version
				"streams-trigger:DON trigger@1.0.0":                  false, // Space in label
				"streams-trigger:DON_trigger:format_json@1.0.0.beta": false, // Extra version part
				"streams-trigger:@1.0.0":                             false, // Empty label
				"streams-trigger:DON_trigger:@1.0.0":                 false, // Empty second label
				"streams-trigger@1.notnum.0":                         false, // Bad minor
				"streams-trigger@1.0.badpatch":                       false, // Bad patch
				"streams-trigger@1.0":                                false, // missing patch
				"streams-trigger@1":                                  false, // missing minor, patch

			},
		},
		{
			name:           "streams-trigger version 2.0.0",
			capabilityName: "streams-trigger",
			version:        mustSemVer(t, "2"),
			testIDs: map[string]bool{
				"streams-trigger@2.0.0":                              true,  // No labels
				"streams-trigger:DON_trigger@2.0.0":                  true,  // One label
				"streams-trigger:DON_trigger:format_json@2.0.0":      true,  // Two labels
				"streams-trigger:DON-trigger:format_json@2.0.0":      true,  // Hyphen in label
				"streams-trigger:a_b:c_d:e_f@2.0.0":                  true,  // Three labels
				"streams-trigger@2.1.0":                              true,  // minor version OK
				"streams-trigger@2.0.1":                              true,  // patch version OK
				"streamstrigger@2.0.0":                               false, // Wrong name
				"streams-trigger@1.0.0":                              false, // Wrong version
				"streams-trigger:DON trigger@2.0.0":                  false, // Space in label
				"streams-trigger:DON_trigger:format_json@2.0.0.beta": false, // Extra version part
				"streams-trigger:@2.0.0":                             false, // Empty label
				"streams-trigger:DON_trigger:@2.0.0":                 false, // Empty second label
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := capabilities.CreateCapabilityIDRegex(tt.capabilityName, tt.version)
			require.NoError(t, err, "Failed to create regex")

			for id, shouldMatch := range tt.testIDs {
				if shouldMatch {
					assert.True(t, regex.MatchString(id), "Expected ID to match, but it didn't: %s", id)
				} else {
					assert.False(t, regex.MatchString(id), "Expected ID not to match, but it did: %s", id)
				}
			}
		})
	}
}

func mustSemVer(t *testing.T, version string) *semver.Version {
	v, err := semver.NewVersion(version)
	require.NoError(t, err)
	return v
}

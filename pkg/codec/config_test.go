package codec_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
)

func TestModifiersConfig(t *testing.T) {
	jsonConfig := `[
    {
        "type": "extract element",
        "extractions": {
            "A": "first element"
        }
    },
    {
        "type": "rename",
        "fields": {
            "A": "Z"
        }
    },
    {
        "type": "drop",
        "fields": [
            "C"
        ]
    },
    {
        "type": "hard code",
        "offChainValues": {
            "B": 2
        }
    }
]`

	lowerJSONConfig := `[
    {
        "type": "extract element",
        "extractions": {
            "a": "first element"
        }
    },
    {
        "type": "rename",
        "fields": {
            "a": "z"
        }
    },
    {
        "type": "drop",
        "fields": [
            "c"
        ]
    },
    {
        "type": "hard code",
        "offChainValues": {
            "b": 2
        }
    }
]`

	for _, test := range []struct{ name, json string }{
		{"exact", jsonConfig},
		// Used to allow config to match on-chain names/convention
		{"lowercase", lowerJSONConfig},
	} {
		t.Run(test.name, func(t *testing.T) {
			conf := &codec.ModifiersConfig{}
			err := conf.UnmarshalJSON([]byte(test.json))
			require.NoError(t, err)
			modifier, err := conf.ToModifier()
			require.NoError(t, err)

			_, err = modifier.RetypeForOffChain(reflect.TypeOf(ModifiersConfigOnChainTestStruct{}), "")
			require.NoError(t, err)

			onChain := ModifiersConfigOnChainTestStruct{
				A: 1,
				C: 100,
			}

			offChain, err := modifier.TransformForOffChain(onChain, "")
			require.NoError(t, err)

			b, err := json.Marshal(offChain)
			require.NoError(t, err)
			actualMap := map[string]any{}
			err = json.Unmarshal(b, &actualMap)
			require.NoError(t, err)

			// when decoding to actualMap, the types are lost
			// the tests for the actual modifiers verify the types are correct
			expectedMap := map[string]any{
				"Z": []any{float64(1)},
				"B": float64(2),
			}

			assert.Equal(t, expectedMap, actualMap)
		})
	}
}

type ModifiersConfigOnChainTestStruct struct {
	A int
	C int
}

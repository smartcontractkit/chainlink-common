package jsonserializable_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/jsonserializable"
)

type hexable struct {
	h string
}

func (h hexable) Hex() string {
	return h.h
}

func TestMarshalJSONSerializable_replaceBytesWithHex(t *testing.T) {
	t.Parallel()

	type jsm = map[string]any

	toJSONSerializable := func(val jsm) *jsonserializable.JSONSerializable {
		return &jsonserializable.JSONSerializable{
			Valid: true,
			Val:   val,
		}
	}

	var (
		testAddr1 = hexable{h: "0x2ab9a2dc53736B361b72d900CdF9F78F9406f111"}
		testAddr2 = hexable{h: "0x2ab9a2dc53736B361b72d900CdF9F78F9406f222"}
		testHash1 = hexable{h: "0x317cfd032b5d6657995f17fe768f7cc4ea0ada27ad421c4caa685a9071eaf111"}
		testHash2 = hexable{h: "0x317cfd032b5d6657995f17fe768f7cc4ea0ada27ad421c4caa685a9071eaf222"}
	)

	tests := []struct {
		name     string
		input    *jsonserializable.JSONSerializable
		expected string
		err      error
	}{
		{"invalid input", &jsonserializable.JSONSerializable{Valid: false}, "null", nil},
		{"empty object", toJSONSerializable(jsm{}), "{}", nil},
		{"byte slice", toJSONSerializable(jsm{"slice": []byte{0x10, 0x20, 0x30}}),
			`{"slice":"0x102030"}`, nil},
		{"address", toJSONSerializable(jsm{"addr": testAddr1}),
			`{"addr":"0x2ab9a2dc53736B361b72d900CdF9F78F9406f111"}`, nil},
		{"hash", toJSONSerializable(jsm{"hash": testHash1}),
			`{"hash":"0x317cfd032b5d6657995f17fe768f7cc4ea0ada27ad421c4caa685a9071eaf111"}`, nil},
		{"slice of byte slice", toJSONSerializable(jsm{"slices": [][]byte{{0x10, 0x11, 0x12}, {0x20, 0x21, 0x22}}}),
			`{"slices":["0x101112","0x202122"]}`, nil},
		{"slice of addresses", toJSONSerializable(jsm{"addresses": []hexable{testAddr1, testAddr2}}),
			`{"addresses":["0x2ab9a2dc53736B361b72d900CdF9F78F9406f111","0x2ab9a2dc53736B361b72d900CdF9F78F9406f222"]}`, nil},
		{"slice of hashes", toJSONSerializable(jsm{"hashes": []hexable{testHash1, testHash2}}),
			`{"hashes":["0x317cfd032b5d6657995f17fe768f7cc4ea0ada27ad421c4caa685a9071eaf111","0x317cfd032b5d6657995f17fe768f7cc4ea0ada27ad421c4caa685a9071eaf222"]}`, nil},
		{"slice of interfaces", toJSONSerializable(jsm{"ifaces": []any{[]byte{0x10, 0x11, 0x12}, []byte{0x20, 0x21, 0x22}}}),
			`{"ifaces":["0x101112","0x202122"]}`, nil},
		{"map", toJSONSerializable(jsm{"map": jsm{"slice": []byte{0x10, 0x11, 0x12}, "addr": testAddr1}}),
			`{"map":{"addr":"0x2ab9a2dc53736B361b72d900CdF9F78F9406f111","slice":"0x101112"}}`, nil},
		{"byte array 4", toJSONSerializable(jsm{"ba4": [4]byte{1, 2, 3, 4}}),
			`{"ba4":"0x01020304"}`, nil},
		{"byte array 8", toJSONSerializable(jsm{"ba8": [8]uint8{1, 2, 3, 4, 5, 6, 7, 8}}),
			`{"ba8":"0x0102030405060708"}`, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bytes, err := test.input.MarshalJSON()
			assert.Equal(t, test.expected, string(bytes))
			assert.ErrorIs(t, test.err, err)
		})
	}
}

func TestUnmarshalJSONSerializable(t *testing.T) {
	t.Parallel()

	big, ok := new(big.Int).SetString("18446744073709551616", 10)
	assert.True(t, ok)

	tests := []struct {
		name, input string
		expected    any
	}{
		{"null json", `null`, nil},
		{"bool", `true`, true},
		{"string", `"foo"`, "foo"},
		{"object with int", `{"foo": 42}`, map[string]any{"foo": int64(42)}},
		{"object with float", `{"foo": 3.14}`, map[string]any{"foo": float64(3.14)}},
		{"object with big int", `{"foo": 18446744073709551616}`, map[string]any{"foo": big}},
		{"slice", `[42, 3.14]`, []any{int64(42), float64(3.14)}},
		{"nested map", `{"m": {"foo": 42}}`, map[string]any{"m": map[string]any{"foo": int64(42)}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var i jsonserializable.JSONSerializable
			err := json.Unmarshal([]byte(test.input), &i)
			require.NoError(t, err)
			if test.expected != nil {
				assert.True(t, i.Valid)
				assert.Equal(t, test.expected, i.Val)
			}
		})
	}
}

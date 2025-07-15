package json

import (
	"bytes"
	"math"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoderPrimitives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		target   any
		expected any
	}{
		// Integer types from strings
		{"int from string", `"42"`, new(int), int(42)},
		{"int8 from string", `"-128"`, new(int8), int8(-128)},
		{"int16 from string", `"32767"`, new(int16), int16(32767)},
		{"int32 from string", `"-2147483648"`, new(int32), int32(-2147483648)},
		{"int64 from string", `"9223372036854775807"`, new(int64), int64(9223372036854775807)},

		// Unsigned integer types from strings
		{"uint from string", `"42"`, new(uint), uint(42)},
		{"uint8 from string", `"255"`, new(uint8), uint8(255)},
		{"uint16 from string", `"65535"`, new(uint16), uint16(65535)},
		{"uint32 from string", `"4294967295"`, new(uint32), uint32(4294967295)},
		{"uint64 from string", `"18446744073709551615"`, new(uint64), uint64(18446744073709551615)},

		// Float types from strings
		{"float32 from string", `"3.14"`, new(float32), float32(3.14)},
		{"float64 from string", `"2.718281828"`, new(float64), float64(2.718281828)},
		{"float64 scientific from string", `"1.23e-10"`, new(float64), float64(1.23e-10)},
		{"float64 large from string", `"1.23e+20"`, new(float64), float64(1.23e+20)},

		// Also test reading numbers directly from JSON
		{"int from number", `42`, new(int), int(42)},
		{"int64 from number", `9223372036854775807`, new(int64), int64(9223372036854775807)},
		{"uint64 from number", `18446744073709551615`, new(uint64), uint64(18446744073709551615)},
		{"float64 from number", `3.14159`, new(float64), float64(3.14159)},

		// Non-numeric types
		{"string", `"hello world"`, new(string), "hello world"},
		{"bool true", `true`, new(bool), true},
		{"bool false", `false`, new(bool), false},

		// String that looks like number but target is string
		{"numeric string to string", `"12345"`, new(string), "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := newDecoder(strings.NewReader(tt.input))
			err := decoder.Decode(tt.target)
			require.NoError(t, err)

			// Dereference the pointer to get the actual value
			actual := reflect.ValueOf(tt.target).Elem().Interface()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestDecoderBigInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *big.Int
	}{
		{"small positive from string", `"42"`, big.NewInt(42)},
		{"small negative from string", `"-42"`, big.NewInt(-42)},
		{"large positive from string", `"18446744073709551615"`, new(big.Int).SetUint64(18446744073709551615)},
		{"very large from string", `"1208925819614629174706175"`, func() *big.Int {
			n := new(big.Int)
			n.SetString("1208925819614629174706175", 10)
			return n
		}()},
		{"from json number", `123456789012345678901234567890`, func() *big.Int {
			n := new(big.Int)
			n.SetString("123456789012345678901234567890", 10)
			return n
		}()},
		{"null big.Int", `null`, (*big.Int)(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result *big.Int
			decoder := newDecoder(strings.NewReader(tt.input))
			err := decoder.Decode(&result)
			require.NoError(t, err)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, 0, tt.expected.Cmp(result))
			}
		})
	}
}

func TestDecoderStructs(t *testing.T) {
	type NestedStruct struct {
		Value   int64   `json:"value"`
		Decimal float64 `json:"decimal"`
	}

	type TestStruct struct {
		ID           uint64       `json:"id"`
		Name         string       `json:"name"`
		Amount       *big.Int     `json:"amount"`
		IsActive     bool         `json:"is_active"`
		Score        float64      `json:"score"`
		Nested       NestedStruct `json:"nested"`
		OptionalInt  *int         `json:"optional_int,omitempty"`
		IgnoredField string       `json:"-"`
	}

	input := `{
		"id": "18446744073709551615",
		"name": "test",
		"amount": "123456789012345678901234567890",
		"is_active": true,
		"score": "99.99",
		"nested": {
			"value": "9223372036854775807",
			"decimal": "3.14159"
		}
	}`

	var result TestStruct
	decoder := newDecoder(strings.NewReader(input))
	err := decoder.Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, uint64(18446744073709551615), result.ID)
	assert.Equal(t, "test", result.Name)

	expectedAmount := new(big.Int)
	expectedAmount.SetString("123456789012345678901234567890", 10)
	assert.Equal(t, 0, expectedAmount.Cmp(result.Amount))

	assert.Equal(t, true, result.IsActive)
	assert.Equal(t, 99.99, result.Score)
	assert.Equal(t, int64(9223372036854775807), result.Nested.Value)
	assert.Equal(t, 3.14159, result.Nested.Decimal)
	assert.Nil(t, result.OptionalInt)
	assert.Equal(t, "", result.IgnoredField)
}

func TestDecoderSlicesAndArrays(t *testing.T) {
	t.Run("slice of integers from strings", func(t *testing.T) {
		input := `["1","2","3","9223372036854775807"]`

		var result []int64
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		expected := []int64{1, 2, 3, 9223372036854775807}
		assert.Equal(t, expected, result)
	})

	t.Run("array of floats from strings", func(t *testing.T) {
		input := `["1.1","2.2","3.3"]`

		var result [3]float64
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		expected := [3]float64{1.1, 2.2, 3.3}
		assert.Equal(t, expected, result)
	})

	t.Run("slice of big.Int from strings", func(t *testing.T) {
		input := `["100","200","18446744073709551615"]`

		var result []*big.Int
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Len(t, result, 3)
		assert.Equal(t, 0, big.NewInt(100).Cmp(result[0]))
		assert.Equal(t, 0, big.NewInt(200).Cmp(result[1]))
		assert.Equal(t, 0, new(big.Int).SetUint64(18446744073709551615).Cmp(result[2]))
	})

	t.Run("nested slices", func(t *testing.T) {
		input := `[["1","2"],["3","4"]]`

		var result [][]int
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		expected := [][]int{{1, 2}, {3, 4}}
		assert.Equal(t, expected, result)
	})
}

func TestDecoderMaps(t *testing.T) {
	t.Run("map with numeric string values", func(t *testing.T) {
		input := `{
			"int": "42",
			"uint": "84",
			"float": "3.14",
			"bigint": "999",
			"string": "hello",
			"bool": true
		}`

		var result map[string]any
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		// When decoding to any, string values remain as strings
		assert.Equal(t, "42", result["int"])
		assert.Equal(t, "84", result["uint"])
		assert.Equal(t, "3.14", result["float"])
		assert.Equal(t, "999", result["bigint"])
		assert.Equal(t, "hello", result["string"])
		assert.Equal(t, true, result["bool"])
	})

	t.Run("typed map", func(t *testing.T) {
		input := `{
			"one": "1",
			"two": "2",
			"three": "3"
		}`

		var result map[string]int
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		expected := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("nested maps", func(t *testing.T) {
		input := `{"level1":{"level2":{"value":"18446744073709551615"}}}`

		type Inner struct {
			Value uint64 `json:"value"`
		}
		type Middle struct {
			Level2 Inner `json:"level2"`
		}
		type Outer struct {
			Level1 Middle `json:"level1"`
		}

		var result Outer
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, uint64(18446744073709551615), result.Level1.Level2.Value)
	})
}

func TestDecoderComplexStructures(t *testing.T) {
	type Transaction struct {
		ID     uint64   `json:"id"`
		Amount *big.Int `json:"amount"`
		Fee    int64    `json:"fee"`
		Data   []byte   `json:"data"`
	}

	type Block struct {
		Number       uint64        `json:"number"`
		Transactions []Transaction `json:"transactions"`
		GasUsed      *big.Int      `json:"gas_used"`
		Timestamp    int64         `json:"timestamp"`
	}

	input := `{
		"number": "15000000",
		"transactions": [
			{
				"id": "1",
				"amount": "1000000000000000000",
				"fee": "21000",
				"data": "AQID"
			},
			{
				"id": "2",
				"amount": "2500000000000000000",
				"fee": "42000",
				"data": "BAUG"
			}
		],
		"gas_used": "21000000000000",
		"timestamp": "1234567890"
	}`

	var result Block
	decoder := newDecoder(strings.NewReader(input))
	err := decoder.Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, uint64(15000000), result.Number)
	assert.Equal(t, int64(1234567890), result.Timestamp)

	expectedGas := new(big.Int)
	expectedGas.SetString("21000000000000", 10)
	assert.Equal(t, 0, expectedGas.Cmp(result.GasUsed))

	assert.Len(t, result.Transactions, 2)

	assert.Equal(t, uint64(1), result.Transactions[0].ID)
	expectedAmount1 := new(big.Int)
	expectedAmount1.SetString("1000000000000000000", 10)
	assert.Equal(t, 0, expectedAmount1.Cmp(result.Transactions[0].Amount))
	assert.Equal(t, int64(21000), result.Transactions[0].Fee)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, result.Transactions[0].Data)

	assert.Equal(t, uint64(2), result.Transactions[1].ID)
	expectedAmount2 := new(big.Int)
	expectedAmount2.SetString("2500000000000000000", 10)
	assert.Equal(t, 0, expectedAmount2.Cmp(result.Transactions[1].Amount))
	assert.Equal(t, int64(42000), result.Transactions[1].Fee)
	assert.Equal(t, []byte{0x04, 0x05, 0x06}, result.Transactions[1].Data)
}

func TestDecoderNullValues(t *testing.T) {
	type TestStruct struct {
		IntPtr    *int     `json:"int_ptr"`
		FloatPtr  *float64 `json:"float_ptr"`
		StringPtr *string  `json:"string_ptr"`
		BigIntPtr *big.Int `json:"bigint_ptr"`
	}

	input := `{
		"int_ptr": null,
		"float_ptr": null,
		"string_ptr": null,
		"bigint_ptr": null
	}`

	var result TestStruct
	decoder := newDecoder(strings.NewReader(input))
	err := decoder.Decode(&result)
	require.NoError(t, err)

	assert.Nil(t, result.IntPtr)
	assert.Nil(t, result.FloatPtr)
	assert.Nil(t, result.StringPtr)
	assert.Nil(t, result.BigIntPtr)
}

func TestDecoderNullMapAndSlice(t *testing.T) {
	t.Run("null map[string]any", func(t *testing.T) {
		var m map[string]any
		decoder := newDecoder(strings.NewReader("null"))
		err := decoder.Decode(&m)
		require.NoError(t, err)
		assert.Nil(t, m, "map[string]any should be nil after decoding null")
	})

	t.Run("null map[string]any", func(t *testing.T) {
		var m map[string]any
		decoder := newDecoder(strings.NewReader("null"))
		err := decoder.Decode(&m)
		require.NoError(t, err)
		assert.Nil(t, m, "map[string]any should be nil after decoding null")
	})

	t.Run("null map[string]string", func(t *testing.T) {
		var m map[string]string
		decoder := newDecoder(strings.NewReader("null"))
		err := decoder.Decode(&m)
		require.NoError(t, err)
		assert.Nil(t, m, "map[string]string should be nil after decoding null")
	})

	t.Run("null slice", func(t *testing.T) {
		var s []string
		decoder := newDecoder(strings.NewReader("null"))
		err := decoder.Decode(&s)
		require.NoError(t, err)
		assert.Nil(t, s, "slice should be nil after decoding null")
	})

	t.Run("null map in struct", func(t *testing.T) {
		type TestStruct struct {
			Map map[string]int `json:"map"`
		}
		var result TestStruct
		decoder := newDecoder(strings.NewReader(`{"map": null}`))
		err := decoder.Decode(&result)
		require.NoError(t, err)
		assert.Nil(t, result.Map, "map field should be nil")
	})
}

func TestDecoderEdgeCases(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		type Empty struct{}
		input := `{}`

		var result Empty
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, Empty{}, result)
	})

	t.Run("empty slice", func(t *testing.T) {
		input := `[]`

		var result []int
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Empty(t, result)
	})

	t.Run("empty map", func(t *testing.T) {
		input := `{}`

		var result map[string]int
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Empty(t, result)
	})

	t.Run("zero values from strings", func(t *testing.T) {
		input := `{
			"int": "0",
			"float": "0",
			"string": "",
			"bool": false
		}`

		type Result struct {
			Int    int     `json:"int"`
			Float  float64 `json:"float"`
			String string  `json:"string"`
			Bool   bool    `json:"bool"`
		}

		var result Result
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, 0, result.Int)
		assert.Equal(t, float64(0), result.Float)
		assert.Equal(t, "", result.String)
		assert.Equal(t, false, result.Bool)
	})

	t.Run("mixed string and numeric JSON", func(t *testing.T) {
		input := `{
			"string_num": "42",
			"regular_num": 84,
			"big_string": "18446744073709551615",
			"big_num": 18446744073709551615
		}`

		type Result struct {
			StringNum  int64  `json:"string_num"`
			RegularNum int64  `json:"regular_num"`
			BigString  uint64 `json:"big_string"`
			BigNum     uint64 `json:"big_num"`
		}

		var result Result
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, int64(42), result.StringNum)
		assert.Equal(t, int64(84), result.RegularNum)
		assert.Equal(t, uint64(18446744073709551615), result.BigString)
		assert.Equal(t, uint64(18446744073709551615), result.BigNum)
	})
}

func TestDecoderErrorCases(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		input := `{invalid json}`

		var result map[string]any
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		assert.Error(t, err)
	})

	t.Run("type mismatch - string to int", func(t *testing.T) {
		input := `{"value": "not a number"}`

		type Result struct {
			Value int `json:"value"`
		}

		var result Result
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse string")
	})

	t.Run("type mismatch - array to struct", func(t *testing.T) {
		input := `[1, 2, 3]`

		type Result struct {
			Value int `json:"value"`
		}

		var result Result
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected object for struct")
	})

	t.Run("nil target", func(t *testing.T) {
		input := `{"value": "42"}`

		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot decode into a non-pointer or nil value")
	})

	t.Run("invalid big.Int string", func(t *testing.T) {
		input := `"not a valid number"`

		var result *big.Int
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})
}

func TestRoundTrip(t *testing.T) {
	type TestData struct {
		SmallInt    int              `json:"small_int"`
		LargeInt    int64            `json:"large_int"`
		UnsignedMax uint64           `json:"unsigned_max"`
		Float       float64          `json:"float"`
		BigNumber   *big.Int         `json:"big_number"`
		String      string           `json:"string"`
		Bool        bool             `json:"bool"`
		IntSlice    []int            `json:"int_slice"`
		IntMap      map[string]int64 `json:"int_map"`
	}

	bigNum := new(big.Int)
	bigNum.SetString("99999999999999999999999999999999999999", 10)

	original := TestData{
		SmallInt:    42,
		LargeInt:    9223372036854775807,
		UnsignedMax: 18446744073709551615,
		Float:       math.Pi,
		BigNumber:   bigNum,
		String:      "test string",
		Bool:        true,
		IntSlice:    []int{1, 2, 3, 4, 5},
		IntMap: map[string]int64{
			"one": 1,
			"max": 9223372036854775807,
		},
	}

	var buf bytes.Buffer
	encoder := newEncoder(&buf)
	err := encoder.Encode(original)
	require.NoError(t, err)

	var decoded TestData
	decoder := newDecoder(&buf)
	err = decoder.Decode(&decoded)
	require.NoError(t, err)

	assert.Equal(t, original.SmallInt, decoded.SmallInt)
	assert.Equal(t, original.LargeInt, decoded.LargeInt)
	assert.Equal(t, original.UnsignedMax, decoded.UnsignedMax)
	assert.Equal(t, original.Float, decoded.Float)
	assert.Equal(t, 0, original.BigNumber.Cmp(decoded.BigNumber))
	assert.Equal(t, original.String, decoded.String)
	assert.Equal(t, original.Bool, decoded.Bool)
	assert.Equal(t, original.IntSlice, decoded.IntSlice)
	assert.Equal(t, original.IntMap, decoded.IntMap)
}

func TestDecoderPointerFields(t *testing.T) {
	type StructWithPointers struct {
		IntPtr     *int       `json:"int_ptr"`
		Int8Ptr    *int8      `json:"int8_ptr"`
		Int16Ptr   *int16     `json:"int16_ptr"`
		Int32Ptr   *int32     `json:"int32_ptr"`
		Int64Ptr   *int64     `json:"int64_ptr"`
		UintPtr    *uint      `json:"uint_ptr"`
		Uint8Ptr   *uint8     `json:"uint8_ptr"`
		Uint16Ptr  *uint16    `json:"uint16_ptr"`
		Uint32Ptr  *uint32    `json:"uint32_ptr"`
		Uint64Ptr  *uint64    `json:"uint64_ptr"`
		Float32Ptr *float32   `json:"float32_ptr"`
		Float64Ptr *float64   `json:"float64_ptr"`
		StringPtr  *string    `json:"string_ptr"`
		BoolPtr    *bool      `json:"bool_ptr"`
		BigIntPtr  *big.Int   `json:"bigint_ptr"`
		TimePtr    *time.Time `json:"time_ptr"`
	}

	t.Run("all pointer fields with values from strings", func(t *testing.T) {
		input := `{
			"int_ptr": "42",
			"int8_ptr": "-128",
			"int16_ptr": "32767",
			"int32_ptr": "-2147483648",
			"int64_ptr": "9223372036854775807",
			"uint_ptr": "42",
			"uint8_ptr": "255",
			"uint16_ptr": "65535",
			"uint32_ptr": "4294967295",
			"uint64_ptr": "18446744073709551615",
			"float32_ptr": "3.14",
			"float64_ptr": "2.718281828",
			"string_ptr": "hello",
			"bool_ptr": true,
			"bigint_ptr": "123456789012345678901234567890",
			"time_ptr": "2023-01-01T00:00:00Z"
		}`

		var result StructWithPointers
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		require.NotNil(t, result.IntPtr)
		assert.Equal(t, 42, *result.IntPtr)

		require.NotNil(t, result.Int8Ptr)
		assert.Equal(t, int8(-128), *result.Int8Ptr)

		require.NotNil(t, result.Int16Ptr)
		assert.Equal(t, int16(32767), *result.Int16Ptr)

		require.NotNil(t, result.Int32Ptr)
		assert.Equal(t, int32(-2147483648), *result.Int32Ptr)

		require.NotNil(t, result.Int64Ptr)
		assert.Equal(t, int64(9223372036854775807), *result.Int64Ptr)

		require.NotNil(t, result.UintPtr)
		assert.Equal(t, uint(42), *result.UintPtr)

		require.NotNil(t, result.Uint8Ptr)
		assert.Equal(t, uint8(255), *result.Uint8Ptr)

		require.NotNil(t, result.Uint16Ptr)
		assert.Equal(t, uint16(65535), *result.Uint16Ptr)

		require.NotNil(t, result.Uint32Ptr)
		assert.Equal(t, uint32(4294967295), *result.Uint32Ptr)

		require.NotNil(t, result.Uint64Ptr)
		assert.Equal(t, uint64(18446744073709551615), *result.Uint64Ptr)

		require.NotNil(t, result.Float32Ptr)
		assert.Equal(t, float32(3.14), *result.Float32Ptr)

		require.NotNil(t, result.Float64Ptr)
		assert.Equal(t, 2.718281828, *result.Float64Ptr)

		require.NotNil(t, result.StringPtr)
		assert.Equal(t, "hello", *result.StringPtr)

		require.NotNil(t, result.BoolPtr)
		assert.Equal(t, true, *result.BoolPtr)

		require.NotNil(t, result.BigIntPtr)
		expected := new(big.Int)
		expected.SetString("123456789012345678901234567890", 10)
		assert.Equal(t, 0, expected.Cmp(result.BigIntPtr))

		require.NotNil(t, result.TimePtr)
		expectedTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		assert.True(t, expectedTime.Equal(*result.TimePtr))
	})

	t.Run("all pointer fields null", func(t *testing.T) {
		input := `{
			"int_ptr": null,
			"int8_ptr": null,
			"int16_ptr": null,
			"int32_ptr": null,
			"int64_ptr": null,
			"uint_ptr": null,
			"uint8_ptr": null,
			"uint16_ptr": null,
			"uint32_ptr": null,
			"uint64_ptr": null,
			"float32_ptr": null,
			"float64_ptr": null,
			"string_ptr": null,
			"bool_ptr": null,
			"bigint_ptr": null,
			"time_ptr": null
		}`

		var result StructWithPointers
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Nil(t, result.IntPtr)
		assert.Nil(t, result.Int8Ptr)
		assert.Nil(t, result.Int16Ptr)
		assert.Nil(t, result.Int32Ptr)
		assert.Nil(t, result.Int64Ptr)
		assert.Nil(t, result.UintPtr)
		assert.Nil(t, result.Uint8Ptr)
		assert.Nil(t, result.Uint16Ptr)
		assert.Nil(t, result.Uint32Ptr)
		assert.Nil(t, result.Uint64Ptr)
		assert.Nil(t, result.Float32Ptr)
		assert.Nil(t, result.Float64Ptr)
		assert.Nil(t, result.StringPtr)
		assert.Nil(t, result.BoolPtr)
		assert.Nil(t, result.BigIntPtr)
		assert.Nil(t, result.TimePtr)
	})

	t.Run("mixed null and non-null pointer fields", func(t *testing.T) {
		input := `{
			"int_ptr": "42",
			"int32_ptr": null,
			"string_ptr": "test",
			"bool_ptr": false,
			"bigint_ptr": null
		}`

		var result StructWithPointers
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		require.NotNil(t, result.IntPtr)
		assert.Equal(t, 42, *result.IntPtr)

		assert.Nil(t, result.Int32Ptr)

		require.NotNil(t, result.StringPtr)
		assert.Equal(t, "test", *result.StringPtr)

		require.NotNil(t, result.BoolPtr)
		assert.Equal(t, false, *result.BoolPtr)

		assert.Nil(t, result.BigIntPtr)
	})
}

func TestDecoderNestedPointers(t *testing.T) {
	t.Run("pointer to pointer", func(t *testing.T) {
		type StructWithDoublePointer struct {
			Value **int `json:"value"`
		}

		input := `{"value": "42"}`
		var result StructWithDoublePointer
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		require.NotNil(t, result.Value)
		require.NotNil(t, *result.Value)
		assert.Equal(t, 42, **result.Value)

		input = `{"value": null}`
		result = StructWithDoublePointer{}
		decoder = newDecoder(strings.NewReader(input))
		err = decoder.Decode(&result)
		require.NoError(t, err)
		assert.Nil(t, result.Value)
	})

	t.Run("pointer to slice", func(t *testing.T) {
		type StructWithPointerToSlice struct {
			Values *[]int `json:"values"`
		}

		input := `{"values": ["1", "2", "3"]}`
		var result StructWithPointerToSlice
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		require.NotNil(t, result.Values)
		assert.Equal(t, []int{1, 2, 3}, *result.Values)
	})

	t.Run("slice of pointers", func(t *testing.T) {
		type StructWithSliceOfPointers struct {
			Values []*int `json:"values"`
		}

		input := `{"values": ["1", null, "3"]}`
		var result StructWithSliceOfPointers
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		require.Len(t, result.Values, 3)

		require.NotNil(t, result.Values[0])
		assert.Equal(t, 1, *result.Values[0])

		assert.Nil(t, result.Values[1])

		require.NotNil(t, result.Values[2])
		assert.Equal(t, 3, *result.Values[2])
	})
}

func TestDecoderByteSlices(t *testing.T) {
	t.Run("byte slice from base64", func(t *testing.T) {
		type StructWithBytes struct {
			Data []byte `json:"data"`
		}

		input := `{"data": "SGVsbG8gV29ybGQ="}` // "Hello World" in base64

		var result StructWithBytes
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, []byte("Hello World"), result.Data)
	})

	t.Run("null byte slice", func(t *testing.T) {
		type StructWithBytes struct {
			Data []byte `json:"data"`
		}

		input := `{"data": null}`

		var result StructWithBytes
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Nil(t, result.Data)
	})

	t.Run("pointer to byte slice", func(t *testing.T) {
		type StructWithBytesPtr struct {
			Data *[]byte `json:"data"`
		}

		input := `{"data": "AQIDBA=="}` // [1,2,3,4] in base64

		var result StructWithBytesPtr
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		require.NotNil(t, result.Data)
		assert.Equal(t, []byte{1, 2, 3, 4}, *result.Data)
	})
}

func TestDecoderTimeHandling(t *testing.T) {
	t.Run("time.Time field", func(t *testing.T) {
		type StructWithTime struct {
			CreatedAt time.Time `json:"created_at"`
		}

		input := `{"created_at": "2023-06-15T14:30:00Z"}`

		var result StructWithTime
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		expected, _ := time.Parse(time.RFC3339, "2023-06-15T14:30:00Z")
		assert.True(t, expected.Equal(result.CreatedAt))
	})

	t.Run("*time.Time field non-null", func(t *testing.T) {
		type StructWithTimePtr struct {
			UpdatedAt *time.Time `json:"updated_at"`
		}

		input := `{"updated_at": "2023-12-25T00:00:00Z"}`

		var result StructWithTimePtr
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		require.NotNil(t, result.UpdatedAt)
		expected, _ := time.Parse(time.RFC3339, "2023-12-25T00:00:00Z")
		assert.True(t, expected.Equal(*result.UpdatedAt))
	})

	t.Run("*time.Time field null", func(t *testing.T) {
		type StructWithTimePtr struct {
			UpdatedAt *time.Time `json:"updated_at"`
		}

		input := `{"updated_at": null}`

		var result StructWithTimePtr
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Nil(t, result.UpdatedAt)
	})

	t.Run("slice of time.Time", func(t *testing.T) {
		type StructWithTimeSlice struct {
			Timestamps []time.Time `json:"timestamps"`
		}

		input := `{"timestamps": ["2023-01-01T00:00:00Z", "2023-06-01T12:00:00Z", "2023-12-31T23:59:59Z"]}`

		var result StructWithTimeSlice
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		require.Len(t, result.Timestamps, 3)

		expected1, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		assert.True(t, expected1.Equal(result.Timestamps[0]))

		expected2, _ := time.Parse(time.RFC3339, "2023-06-01T12:00:00Z")
		assert.True(t, expected2.Equal(result.Timestamps[1]))

		expected3, _ := time.Parse(time.RFC3339, "2023-12-31T23:59:59Z")
		assert.True(t, expected3.Equal(result.Timestamps[2]))
	})
}

func TestDecoderInterfaceFields(t *testing.T) {
	type StructWithInterface struct {
		Value    any            `json:"value"`
		Values   []any          `json:"values"`
		Metadata map[string]any `json:"metadata"`
	}

	t.Run("interface with string number", func(t *testing.T) {
		input := `{
			"value": "42",
			"values": ["1", "2", "3"],
			"metadata": {
				"count": "100",
				"ratio": "3.14"
			}
		}`

		var result StructWithInterface
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		// String numbers remain as strings when decoding to any
		assert.Equal(t, "42", result.Value)
		assert.Equal(t, []any{"1", "2", "3"}, result.Values)
		assert.Equal(t, "100", result.Metadata["count"])
		assert.Equal(t, "3.14", result.Metadata["ratio"])
	})

	t.Run("interface with mixed types", func(t *testing.T) {
		input := `{
			"value": true,
			"values": ["string", true, null, "123"],
			"metadata": {
				"name": "test",
				"active": true,
				"count": "999"
			}
		}`

		var result StructWithInterface
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, true, result.Value)
		assert.Equal(t, []any{"string", true, nil, "123"}, result.Values)
		assert.Equal(t, "test", result.Metadata["name"])
		assert.Equal(t, true, result.Metadata["active"])
		assert.Equal(t, "999", result.Metadata["count"])
	})
}

func TestDecoderArrayFields(t *testing.T) {
	type StructWithArrays struct {
		SmallArray [3]int   `json:"small_array"`
		ByteArray  [32]byte `json:"byte_array"`
		BoolArray  [2]bool  `json:"bool_array"`
		MixedArray [4]any   `json:"mixed_array"`
	}

	t.Run("arrays from JSON", func(t *testing.T) {
		input := `{
			"small_array": ["1", "2", "3"],
			"byte_array": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			"bool_array": [true, false],
			"mixed_array": ["str", "123", true, null]
		}`

		var result StructWithArrays
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, [3]int{1, 2, 3}, result.SmallArray)
		assert.Equal(t, [32]byte{}, result.ByteArray) // All zeros
		assert.Equal(t, [2]bool{true, false}, result.BoolArray)
		assert.Equal(t, [4]any{"str", "123", true, nil}, result.MixedArray)
	})

	t.Run("array size mismatch error", func(t *testing.T) {
		input := `{
			"small_array": ["1", "2"]
		}`

		var result StructWithArrays
		decoder := newDecoder(strings.NewReader(input))
		err := decoder.Decode(&result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "array size mismatch")
	})
}

func TestDecoderComplexNestedStructs(t *testing.T) {
	type Inner struct {
		ID    *int64   `json:"id"`
		Value *big.Int `json:"value"`
		Tags  []string `json:"tags"`
	}

	type Middle struct {
		Name      string              `json:"name"`
		Inner     *Inner              `json:"inner"`
		Metadata  map[string]*float64 `json:"metadata"`
		Timestamp *time.Time          `json:"timestamp"`
	}

	type Outer struct {
		Version  int                       `json:"version"`
		Middles  []*Middle                 `json:"middles"`
		Settings map[string]map[string]any `json:"settings"`
	}

	input := `{
		"version": "1",
		"middles": [
			{
				"name": "first",
				"inner": {
					"id": "12345",
					"value": "999999999999999999999999999",
					"tags": ["tag1", "tag2"]
				},
				"metadata": {
					"score": "98.5",
					"ratio": "0.75"
				},
				"timestamp": "2023-01-01T00:00:00Z"
			},
			null,
			{
				"name": "third",
				"inner": null,
				"metadata": {
					"score": null,
					"ratio": "1.5"
				},
				"timestamp": null
			}
		],
		"settings": {
			"feature1": {
				"enabled": true,
				"value": "100"
			},
			"feature2": {
				"enabled": false,
				"value": "200"
			}
		}
	}`

	var result Outer
	decoder := newDecoder(strings.NewReader(input))
	err := decoder.Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Version)
	require.Len(t, result.Middles, 3)

	require.NotNil(t, result.Middles[0])
	assert.Equal(t, "first", result.Middles[0].Name)
	require.NotNil(t, result.Middles[0].Inner)
	require.NotNil(t, result.Middles[0].Inner.ID)
	assert.Equal(t, int64(12345), *result.Middles[0].Inner.ID)

	expectedValue := new(big.Int)
	expectedValue.SetString("999999999999999999999999999", 10)
	assert.Equal(t, 0, expectedValue.Cmp(result.Middles[0].Inner.Value))

	assert.Equal(t, []string{"tag1", "tag2"}, result.Middles[0].Inner.Tags)

	require.NotNil(t, result.Middles[0].Metadata["score"])
	assert.Equal(t, 98.5, *result.Middles[0].Metadata["score"])
	require.NotNil(t, result.Middles[0].Metadata["ratio"])
	assert.Equal(t, 0.75, *result.Middles[0].Metadata["ratio"])

	assert.Nil(t, result.Middles[1])

	require.NotNil(t, result.Middles[2])
	assert.Equal(t, "third", result.Middles[2].Name)
	assert.Nil(t, result.Middles[2].Inner)
	assert.Nil(t, result.Middles[2].Metadata["score"])
	require.NotNil(t, result.Middles[2].Metadata["ratio"])
	assert.Equal(t, 1.5, *result.Middles[2].Metadata["ratio"])
	assert.Nil(t, result.Middles[2].Timestamp)

	assert.Equal(t, true, result.Settings["feature1"]["enabled"])
	assert.Equal(t, "100", result.Settings["feature1"]["value"])
	assert.Equal(t, false, result.Settings["feature2"]["enabled"])
	assert.Equal(t, "200", result.Settings["feature2"]["value"])
}

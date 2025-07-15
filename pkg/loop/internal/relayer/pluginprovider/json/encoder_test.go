package json

import (
	"bytes"
	"encoding/json"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoderPrimitives(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		// Integer types
		{"int", int(42), `"42"`},
		{"int8", int8(-128), `"-128"`},
		{"int16", int16(32767), `"32767"`},
		{"int32", int32(-2147483648), `"-2147483648"`},
		{"int64", int64(9223372036854775807), `"9223372036854775807"`},

		// Unsigned integer types
		{"uint", uint(42), `"42"`},
		{"uint8", uint8(255), `"255"`},
		{"uint16", uint16(65535), `"65535"`},
		{"uint32", uint32(4294967295), `"4294967295"`},
		{"uint64", uint64(18446744073709551615), `"18446744073709551615"`},

		// Float types
		{"float32", float32(3.14), `"3.140000104904175"`}, // float32 has precision limitations
		{"float64", float64(2.718281828), `"2.718281828"`},
		{"float64 scientific", float64(1.23e-10), `"1.23e-10"`},
		{"float64 large", float64(1.23e+20), `"1.23e+20"`},

		// Special float values
		{"float64 max", math.MaxFloat64, `"1.7976931348623157e+308"`},
		{"float64 smallest", math.SmallestNonzeroFloat64, `"5e-324"`},

		// Non-numeric types (should not be converted)
		{"string", "hello world", `"hello world"`},
		{"bool true", true, `true`},
		{"bool false", false, `false`},
		{"nil", nil, `null`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := newEncoder(&buf)
			err := encoder.Encode(tt.input)
			require.NoError(t, err)

			// Trim newline added by encoder
			result := bytes.TrimSpace(buf.Bytes())
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestEncoderBigInt(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected string
	}{
		{"small positive", big.NewInt(42), `"42"`},
		{"small negative", big.NewInt(-42), `"-42"`},
		{"large positive", new(big.Int).SetUint64(18446744073709551615), `"18446744073709551615"`},
		{"very large", new(big.Int).SetBytes([]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255}), `"1208925819614629174706175"`},
		{"nil big.Int", (*big.Int)(nil), `null`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := newEncoder(&buf)
			err := encoder.Encode(tt.input)
			require.NoError(t, err)

			result := bytes.TrimSpace(buf.Bytes())
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestEncoderStructs(t *testing.T) {
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
		unexported   int
	}

	bigAmount := new(big.Int)
	bigAmount.SetString("123456789012345678901234567890", 10)

	input := TestStruct{
		ID:       18446744073709551615, // max uint64
		Name:     "test",
		Amount:   bigAmount,
		IsActive: true,
		Score:    99.99,
		Nested: NestedStruct{
			Value:   9223372036854775807, // max int64
			Decimal: 3.14159,
		},
		IgnoredField: "should not appear",
		unexported:   42,
	}

	var buf bytes.Buffer
	encoder := newEncoder(&buf)
	err := encoder.Encode(input)
	require.NoError(t, err)

	// Parse the result to verify structure
	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "18446744073709551615", result["id"])
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "123456789012345678901234567890", result["amount"])
	assert.Equal(t, true, result["is_active"])
	assert.Equal(t, "99.99", result["score"])

	nested := result["nested"].(map[string]any)
	assert.Equal(t, "9223372036854775807", nested["value"])
	assert.Equal(t, "3.14159", nested["decimal"])

	// Verify omitted and ignored fields
	_, hasOptional := result["optional_int"]
	assert.False(t, hasOptional, "omitempty field should not be present")

	_, hasIgnored := result["IgnoredField"]
	assert.False(t, hasIgnored, "ignored field should not be present")
}

func TestEncoderSlicesAndArrays(t *testing.T) {
	t.Run("slice of integers", func(t *testing.T) {
		input := []int64{1, 2, 3, 9223372036854775807}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `["1","2","3","9223372036854775807"]`
		result := bytes.TrimSpace(buf.Bytes())
		assert.Equal(t, expected, string(result))
	})

	t.Run("array of floats", func(t *testing.T) {
		input := [3]float64{1.1, 2.2, 3.3}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `["1.1","2.2","3.3"]`
		result := bytes.TrimSpace(buf.Bytes())
		assert.Equal(t, expected, string(result))
	})

	t.Run("slice of big.Int", func(t *testing.T) {
		input := []*big.Int{
			big.NewInt(100),
			big.NewInt(200),
			new(big.Int).SetUint64(18446744073709551615),
		}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `["100","200","18446744073709551615"]`
		result := bytes.TrimSpace(buf.Bytes())
		assert.Equal(t, expected, string(result))
	})

	t.Run("nested slices", func(t *testing.T) {
		input := [][]int{{1, 2}, {3, 4}}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `[["1","2"],["3","4"]]`
		result := bytes.TrimSpace(buf.Bytes())
		assert.Equal(t, expected, string(result))
	})
}

func TestEncoderMaps(t *testing.T) {
	t.Run("map with numeric values", func(t *testing.T) {
		input := map[string]any{
			"int":    int64(42),
			"uint":   uint64(84),
			"float":  3.14,
			"bigint": big.NewInt(999),
			"string": "hello",
			"bool":   true,
		}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "42", result["int"])
		assert.Equal(t, "84", result["uint"])
		assert.Equal(t, "3.14", result["float"])
		assert.Equal(t, "999", result["bigint"])
		assert.Equal(t, "hello", result["string"])
		assert.Equal(t, true, result["bool"])
	})

	t.Run("nested maps", func(t *testing.T) {
		input := map[string]any{
			"level1": map[string]any{
				"level2": map[string]any{
					"value": uint64(18446744073709551615),
				},
			},
		}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `{"level1":{"level2":{"value":"18446744073709551615"}}}`
		result := bytes.TrimSpace(buf.Bytes())
		assert.JSONEq(t, expected, string(result))
	})
}

func TestEncoderComplexStructures(t *testing.T) {
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

	// Create test data
	amount1 := new(big.Int)
	amount1.SetString("1000000000000000000", 10) // 1 ETH in wei

	amount2 := new(big.Int)
	amount2.SetString("2500000000000000000", 10) // 2.5 ETH in wei

	gasUsed := new(big.Int)
	gasUsed.SetString("21000000000000", 10)

	block := Block{
		Number: 15000000,
		Transactions: []Transaction{
			{
				ID:     1,
				Amount: amount1,
				Fee:    21000,
				Data:   []byte{0x01, 0x02, 0x03},
			},
			{
				ID:     2,
				Amount: amount2,
				Fee:    42000,
				Data:   []byte{0x04, 0x05, 0x06},
			},
		},
		GasUsed:   gasUsed,
		Timestamp: 1234567890,
	}

	var buf bytes.Buffer
	encoder := newEncoder(&buf)
	err := encoder.Encode(block)
	require.NoError(t, err)

	// Verify the encoded JSON
	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "15000000", result["number"])
	assert.Equal(t, "21000000000000", result["gas_used"])
	assert.Equal(t, "1234567890", result["timestamp"])

	txs := result["transactions"].([]any)
	assert.Len(t, txs, 2)

	tx1 := txs[0].(map[string]any)
	assert.Equal(t, "1", tx1["id"])
	assert.Equal(t, "1000000000000000000", tx1["amount"])
	assert.Equal(t, "21000", tx1["fee"])
}

func TestEncoderPointers(t *testing.T) {
	t.Run("pointer to int", func(t *testing.T) {
		value := int64(42)
		input := &value

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `"42"`
		result := bytes.TrimSpace(buf.Bytes())
		assert.Equal(t, expected, string(result))
	})

	t.Run("nil pointer", func(t *testing.T) {
		var input *int64

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `null`
		result := bytes.TrimSpace(buf.Bytes())
		assert.Equal(t, expected, string(result))
	})

	t.Run("pointer to struct", func(t *testing.T) {
		type Simple struct {
			Value int `json:"value"`
		}

		input := &Simple{Value: 100}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `{"value":"100"}`
		result := bytes.TrimSpace(buf.Bytes())
		assert.JSONEq(t, expected, string(result))
	})
}

func TestEncoderEdgeCases(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		type Empty struct{}
		input := Empty{}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `{}`
		result := bytes.TrimSpace(buf.Bytes())
		assert.Equal(t, expected, string(result))
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []int{}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `[]`
		result := bytes.TrimSpace(buf.Bytes())
		assert.Equal(t, expected, string(result))
	})

	t.Run("empty map", func(t *testing.T) {
		input := map[string]int{}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `{}`
		result := bytes.TrimSpace(buf.Bytes())
		assert.Equal(t, expected, string(result))
	})

	t.Run("zero values", func(t *testing.T) {
		input := struct {
			Int    int     `json:"int"`
			Float  float64 `json:"float"`
			String string  `json:"string"`
			Bool   bool    `json:"bool"`
		}{}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `{"int":"0","float":"0","string":"","bool":false}`
		result := bytes.TrimSpace(buf.Bytes())
		assert.JSONEq(t, expected, string(result))
	})
}

// TestEncoderPointerFields tests encoding structs with pointer fields
func TestEncoderPointerFields(t *testing.T) {
	type StructWithPointers struct {
		IntPtr     *int       `json:"int_ptr"`
		Int32Ptr   *int32     `json:"int32_ptr"`
		Int64Ptr   *int64     `json:"int64_ptr"`
		Uint64Ptr  *uint64    `json:"uint64_ptr"`
		Float64Ptr *float64   `json:"float64_ptr"`
		StringPtr  *string    `json:"string_ptr"`
		BoolPtr    *bool      `json:"bool_ptr"`
		BigIntPtr  *big.Int   `json:"bigint_ptr"`
		TimePtr    *time.Time `json:"time_ptr"`
	}

	t.Run("all pointer fields with values", func(t *testing.T) {
		intVal := 42
		int32Val := int32(-2147483648)
		int64Val := int64(9223372036854775807)
		uint64Val := uint64(18446744073709551615)
		float64Val := 3.14159
		stringVal := "hello"
		boolVal := true
		bigIntVal := new(big.Int)
		bigIntVal.SetString("123456789012345678901234567890", 10)
		timeVal := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		input := StructWithPointers{
			IntPtr:     &intVal,
			Int32Ptr:   &int32Val,
			Int64Ptr:   &int64Val,
			Uint64Ptr:  &uint64Val,
			Float64Ptr: &float64Val,
			StringPtr:  &stringVal,
			BoolPtr:    &boolVal,
			BigIntPtr:  bigIntVal,
			TimePtr:    &timeVal,
		}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "42", result["int_ptr"])
		assert.Equal(t, "-2147483648", result["int32_ptr"])
		assert.Equal(t, "9223372036854775807", result["int64_ptr"])
		assert.Equal(t, "18446744073709551615", result["uint64_ptr"])
		assert.Equal(t, "3.14159", result["float64_ptr"])
		assert.Equal(t, "hello", result["string_ptr"])
		assert.Equal(t, true, result["bool_ptr"])
		assert.Equal(t, "123456789012345678901234567890", result["bigint_ptr"])
		assert.Equal(t, "2023-01-01T00:00:00Z", result["time_ptr"])
	})

	t.Run("all pointer fields nil", func(t *testing.T) {
		input := StructWithPointers{}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Nil(t, result["int_ptr"])
		assert.Nil(t, result["int32_ptr"])
		assert.Nil(t, result["int64_ptr"])
		assert.Nil(t, result["uint64_ptr"])
		assert.Nil(t, result["float64_ptr"])
		assert.Nil(t, result["string_ptr"])
		assert.Nil(t, result["bool_ptr"])
		assert.Nil(t, result["bigint_ptr"])
		assert.Nil(t, result["time_ptr"])
	})

	t.Run("mixed nil and non-nil pointers", func(t *testing.T) {
		intVal := 100
		stringVal := "test"

		input := StructWithPointers{
			IntPtr:    &intVal,
			Int32Ptr:  nil,
			StringPtr: &stringVal,
			BoolPtr:   nil,
		}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "100", result["int_ptr"])
		assert.Nil(t, result["int32_ptr"])
		assert.Equal(t, "test", result["string_ptr"])
		assert.Nil(t, result["bool_ptr"])
	})
}

// TestEncoderNestedPointers tests encoding nested pointer types
func TestEncoderNestedPointers(t *testing.T) {
	t.Run("pointer to pointer", func(t *testing.T) {
		type StructWithDoublePointer struct {
			Value **int `json:"value"`
		}

		// Non-nil case
		val := 42
		ptr := &val
		input := StructWithDoublePointer{Value: &ptr}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `{"value":"42"}`
		assert.JSONEq(t, expected, buf.String())

		// Nil cases
		input = StructWithDoublePointer{Value: nil}
		buf.Reset()
		err = encoder.Encode(input)
		require.NoError(t, err)

		expected = `{"value":null}`
		assert.JSONEq(t, expected, buf.String())

		// Pointer to nil
		var nilPtr *int
		input = StructWithDoublePointer{Value: &nilPtr}
		buf.Reset()
		err = encoder.Encode(input)
		require.NoError(t, err)

		expected = `{"value":null}`
		assert.JSONEq(t, expected, buf.String())
	})

	t.Run("slice of pointers", func(t *testing.T) {
		val1, val3 := 1, 3
		input := []*int{&val1, nil, &val3}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `["1",null,"3"]`
		assert.JSONEq(t, expected, buf.String())
	})

	t.Run("pointer to slice", func(t *testing.T) {
		slice := []int{1, 2, 3}
		input := &slice

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `["1","2","3"]`
		assert.JSONEq(t, expected, buf.String())
	})

	t.Run("map of pointers", func(t *testing.T) {
		val1, val2 := int64(100), int64(200)
		input := map[string]*int64{
			"first":  &val1,
			"second": &val2,
			"third":  nil,
		}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "100", result["first"])
		assert.Equal(t, "200", result["second"])
		assert.Nil(t, result["third"])
	})
}

// TestEncoderByteSlices tests byte slice encoding
func TestEncoderByteSlices(t *testing.T) {
	t.Run("byte slice", func(t *testing.T) {
		input := []byte("Hello World")

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `"SGVsbG8gV29ybGQ="` // Base64 encoded
		assert.Equal(t, expected, string(bytes.TrimSpace(buf.Bytes())))
	})

	t.Run("nil byte slice", func(t *testing.T) {
		var input []byte

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `null`
		assert.Equal(t, expected, string(bytes.TrimSpace(buf.Bytes())))
	})

	t.Run("empty byte slice", func(t *testing.T) {
		input := []byte{}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `""` // Empty base64
		assert.Equal(t, expected, string(bytes.TrimSpace(buf.Bytes())))
	})

	t.Run("struct with byte slice", func(t *testing.T) {
		type DataStruct struct {
			ID   int    `json:"id"`
			Data []byte `json:"data"`
		}

		input := DataStruct{
			ID:   1,
			Data: []byte{1, 2, 3, 4},
		}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "1", result["id"])
		assert.Equal(t, "AQIDBA==", result["data"]) // Base64 for [1,2,3,4]
	})
}

// TestEncoderTimeHandling tests time.Time encoding
func TestEncoderTimeHandling(t *testing.T) {
	t.Run("time.Time value", func(t *testing.T) {
		input := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `"2023-06-15T14:30:00Z"`
		assert.Equal(t, expected, string(bytes.TrimSpace(buf.Bytes())))
	})

	t.Run("*time.Time pointer", func(t *testing.T) {
		timeVal := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
		input := &timeVal

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `"2023-12-25T00:00:00Z"`
		assert.Equal(t, expected, string(bytes.TrimSpace(buf.Bytes())))
	})

	t.Run("struct with time fields", func(t *testing.T) {
		type TimeStruct struct {
			Created time.Time  `json:"created"`
			Updated *time.Time `json:"updated"`
			Deleted *time.Time `json:"deleted"`
		}

		updated := time.Date(2023, 7, 1, 12, 0, 0, 0, time.UTC)
		input := TimeStruct{
			Created: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			Updated: &updated,
			Deleted: nil,
		}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "2023-01-01T00:00:00Z", result["created"])
		assert.Equal(t, "2023-07-01T12:00:00Z", result["updated"])
		assert.Nil(t, result["deleted"])
	})
}

// TestEncoderArrays tests fixed-size array encoding
func TestEncoderArrays(t *testing.T) {
	t.Run("int array", func(t *testing.T) {
		input := [3]int{1, 2, 3}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `["1","2","3"]`
		assert.JSONEq(t, expected, buf.String())
	})

	t.Run("byte array (not []byte)", func(t *testing.T) {
		input := [4]byte{0x01, 0x02, 0x03, 0x04}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		// Byte arrays are encoded as regular arrays, not base64
		expected := `["1","2","3","4"]`
		assert.JSONEq(t, expected, buf.String())
	})

	t.Run("struct with arrays", func(t *testing.T) {
		type ArrayStruct struct {
			Nums   [3]int64   `json:"nums"`
			Flags  [2]bool    `json:"flags"`
			Values [4]float64 `json:"values"`
		}

		input := ArrayStruct{
			Nums:   [3]int64{100, 200, 300},
			Flags:  [2]bool{true, false},
			Values: [4]float64{1.1, 2.2, 3.3, 4.4},
		}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		nums := result["nums"].([]any)
		assert.Equal(t, []any{"100", "200", "300"}, nums)

		flags := result["flags"].([]any)
		assert.Equal(t, []any{true, false}, flags)

		values := result["values"].([]any)
		assert.Equal(t, []any{"1.1", "2.2", "3.3", "4.4"}, values)
	})
}

// customTime is a type that implements TextMarshaler for testing
type customTime struct {
	T time.Time
}

func (ct customTime) MarshalText() ([]byte, error) {
	return []byte("custom-time-format"), nil
}

// TestEncoderTextMarshaler tests types implementing TextMarshaler
func TestEncoderTextMarshaler(t *testing.T) {
	t.Run("struct implementing TextMarshaler", func(t *testing.T) {
		input := customTime{T: time.Now()}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		// Should use the TextMarshaler implementation
		expected := `"custom-time-format"`
		assert.Equal(t, expected, string(bytes.TrimSpace(buf.Bytes())))
	})

	t.Run("pointer to TextMarshaler", func(t *testing.T) {
		input := &customTime{T: time.Now()}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		expected := `"custom-time-format"`
		assert.Equal(t, expected, string(bytes.TrimSpace(buf.Bytes())))
	})
}

// customData is a type that implements json.Marshaler for testing
type customData struct {
	value string
}

func (cd customData) MarshalJSON() ([]byte, error) {
	return []byte(`{"custom":true,"value":"` + cd.value + `"}`), nil
}

// TestEncoderJSONMarshaler tests types implementing json.Marshaler
func TestEncoderJSONMarshaler(t *testing.T) {
	t.Run("struct implementing json.Marshaler", func(t *testing.T) {
		input := customData{value: "test"}

		var buf bytes.Buffer
		encoder := newEncoder(&buf)
		err := encoder.Encode(input)
		require.NoError(t, err)

		// Should use the MarshalJSON implementation
		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, true, result["custom"])
		assert.Equal(t, "test", result["value"])
	})
}

// TestEncoderComplexNestedStructures tests deeply nested complex structures
func TestEncoderComplexNestedStructures(t *testing.T) {
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

	// Build complex structure
	id1 := int64(12345)
	value1 := new(big.Int)
	value1.SetString("999999999999999999999999999", 10)
	score1 := 98.5
	ratio1 := 0.75
	timestamp1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	ratio2 := 1.5

	outer := Outer{
		Version: 1,
		Middles: []*Middle{
			{
				Name: "first",
				Inner: &Inner{
					ID:    &id1,
					Value: value1,
					Tags:  []string{"tag1", "tag2"},
				},
				Metadata: map[string]*float64{
					"score": &score1,
					"ratio": &ratio1,
				},
				Timestamp: &timestamp1,
			},
			nil,
			{
				Name:  "third",
				Inner: nil,
				Metadata: map[string]*float64{
					"score": nil,
					"ratio": &ratio2,
				},
				Timestamp: nil,
			},
		},
		Settings: map[string]map[string]any{
			"feature1": {
				"enabled": true,
				"value":   int64(100),
			},
			"feature2": {
				"enabled": false,
				"value":   int64(200),
			},
		},
	}

	var buf bytes.Buffer
	encoder := newEncoder(&buf)
	err := encoder.Encode(outer)
	require.NoError(t, err)

	// Verify structure
	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "1", result["version"])

	middles := result["middles"].([]any)
	require.Len(t, middles, 3)

	// First middle
	middle1 := middles[0].(map[string]any)
	assert.Equal(t, "first", middle1["name"])

	inner1 := middle1["inner"].(map[string]any)
	assert.Equal(t, "12345", inner1["id"])
	assert.Equal(t, "999999999999999999999999999", inner1["value"])

	metadata1 := middle1["metadata"].(map[string]any)
	assert.Equal(t, "98.5", metadata1["score"])
	assert.Equal(t, "0.75", metadata1["ratio"])

	// Second middle is nil
	assert.Nil(t, middles[1])

	// Settings
	settings := result["settings"].(map[string]any)
	feature1 := settings["feature1"].(map[string]any)
	assert.Equal(t, true, feature1["enabled"])
	assert.Equal(t, "100", feature1["value"])
}

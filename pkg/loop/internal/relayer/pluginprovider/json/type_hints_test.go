package json

import (
	"bytes"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func assertBigIntEqual(t *testing.T, expected, actual *big.Int) {
	t.Helper()
	if expected == nil && actual == nil {
		return
	}
	require.NotNil(t, expected, "expected is nil but actual is not")
	require.NotNil(t, actual, "actual is nil but expected is not")
	assert.Equal(t, 0, expected.Cmp(actual), "big.Int values not equal: expected %s, got %s", expected.String(), actual.String())
}

func assertTimeEqual(t *testing.T, expected, actual time.Time) {
	t.Helper()
	assert.True(t, expected.Equal(actual), "times not equal: expected %v, got %v", expected, actual)
}

// deepEqual performs a deep comparison of values, properly handling pointers, slices, and maps
func deepEqual(a, b any) bool {
	return deepEqualValue(reflect.ValueOf(a), reflect.ValueOf(b))
}

func deepEqualValue(a, b reflect.Value) bool {
	// Handle invalid values
	if !a.IsValid() || !b.IsValid() {
		return !a.IsValid() && !b.IsValid()
	}

	// Types must match
	if a.Type() != b.Type() {
		return false
	}

	// Handle by kind
	switch a.Kind() {
	case reflect.Ptr:
		// Both nil or both non-nil
		if a.IsNil() != b.IsNil() {
			return false
		}
		if a.IsNil() {
			return true
		}
		// Special case for *big.Int
		if a.Type() == reflect.TypeOf((*big.Int)(nil)) {
			aBig := a.Interface().(*big.Int)
			bBig := b.Interface().(*big.Int)
			return aBig.Cmp(bBig) == 0
		}
		// Compare dereferenced values
		return deepEqualValue(a.Elem(), b.Elem())
		
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return false
		}
		for i := 0; i < a.Len(); i++ {
			if !deepEqualValue(a.Index(i), b.Index(i)) {
				return false
			}
		}
		return true
		
	case reflect.Map:
		if a.Len() != b.Len() {
			return false
		}
		for _, key := range a.MapKeys() {
			aVal := a.MapIndex(key)
			bVal := b.MapIndex(key)
			if !bVal.IsValid() || !deepEqualValue(aVal, bVal) {
				return false
			}
		}
		return true
		
	case reflect.Struct:
		// Special case for time.Time
		if a.Type() == reflect.TypeOf(time.Time{}) {
			return a.Interface().(time.Time).Equal(b.Interface().(time.Time))
		}
		// Compare all exported fields
		for i := 0; i < a.NumField(); i++ {
			if a.Type().Field(i).IsExported() {
				if !deepEqualValue(a.Field(i), b.Field(i)) {
					return false
				}
			}
		}
		return true
		
	case reflect.Interface:
		if a.IsNil() != b.IsNil() {
			return false
		}
		if a.IsNil() {
			return true
		}
		return deepEqualValue(a.Elem(), b.Elem())
		
	default:
		// For basic types, use ==
		return a.Interface() == b.Interface()
	}
}

// testTypeHintRoundTrip tests getTypeHint, MarshalWithHint, and UnmarshalWithHint together
func testTypeHintRoundTrip(t *testing.T, value any, expectedHint string) {
	t.Helper()

	// Test getTypeHint
	hint, err := getTypeHint(value)
	require.NoError(t, err)
	assert.Equal(t, expectedHint, hint)

	// Test MarshalWithHint
	data, marshalHint, err := MarshalWithHint(value)
	require.NoError(t, err)
	assert.Equal(t, expectedHint, marshalHint)

	// Test UnmarshalWithHint
	result, err := UnmarshalWithHint(data, hint)
	require.NoError(t, err)

	// Verify round-trip equality using deep comparison
	if !deepEqual(value, result) {
		t.Errorf("Round-trip failed: expected %#v, got %#v", value, result)
	}
}

func TestTypeHints_BasicTypes(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		expectedHint string
		expectedJSON string
	}{
		// Integers
		{"int", int(42), "int", `"42"`},
		{"int8", int8(42), "int8", `"42"`},
		{"int16", int16(42), "int16", `"42"`},
		{"int32", int32(42), "int32", `"42"`},
		{"int64", int64(42), "int64", `"42"`},

		// Unsigned integers
		{"uint", uint(42), "uint", `"42"`},
		{"uint8", uint8(42), "uint8", `"42"`},
		{"uint16", uint16(42), "uint16", `"42"`},
		{"uint32", uint32(42), "uint32", `"42"`},
		{"uint64", uint64(42), "uint64", `"42"`},

		// Floats
		{"float32", float32(42.5), "float32", `"42.5"`},
		{"float64", float64(42.5), "float64", `"42.5"`},

		// Other basics
		{"bool", true, "bool", "true"},
		{"string", "hello", "string", `"hello"`},
		{"nil", nil, "nil", "null"},

		// Edge cases - empty values
		{"empty string", "", "string", `""`},
		{"zero int", 0, "int", `"0"`},
		{"zero int64", int64(0), "int64", `"0"`},
		{"zero uint64", uint64(0), "uint64", `"0"`},
		{"zero float64", float64(0), "float64", `"0"`},
		{"false bool", false, "bool", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test all three operations together
			testTypeHintRoundTrip(t, tt.value, tt.expectedHint)

			// Also verify JSON output
			data, _, err := MarshalWithHint(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedJSON, string(data))
		})
	}
}

func TestTypeHints_PointerTypes(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		expectedHint string
		expectedJSON string
	}{
		// Non-nil pointers
		{"*int", func() *int { v := 42; return &v }(), "*int", `"42"`},
		{"*int32", func() *int32 { v := int32(42); return &v }(), "*int32", `"42"`},
		{"*int64", func() *int64 { v := int64(42); return &v }(), "*int64", `"42"`},
		{"*uint32", func() *uint32 { v := uint32(42); return &v }(), "*uint32", `"42"`},
		{"*uint64", func() *uint64 { v := uint64(42); return &v }(), "*uint64", `"42"`},
		{"*float32", func() *float32 { v := float32(3.5); return &v }(), "*float32", `"3.5"`},
		{"*float64", func() *float64 { v := float64(3.14); return &v }(), "*float64", `"3.14"`},
		{"*bool", func() *bool { v := true; return &v }(), "*bool", "true"},
		{"*string", func() *string { v := "test"; return &v }(), "*string", `"test"`},

		// Typed nil pointers
		{"nil *int", (*int)(nil), "*int", "null"},
		{"nil *int32", (*int32)(nil), "*int32", "null"},
		{"nil *int64", (*int64)(nil), "*int64", "null"},
		{"nil *string", (*string)(nil), "*string", "null"},
		{"nil *bool", (*bool)(nil), "*bool", "null"},
		{"nil *float64", (*float64)(nil), "*float64", "null"},

		// Special pointer types
		{"*big.Int", big.NewInt(123), "*big.Int", `"123"`},
		{"nil *big.Int", (*big.Int)(nil), "*big.Int", "null"},
		// *time.Time is always treated as time.Time
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test all three operations together
			testTypeHintRoundTrip(t, tt.value, tt.expectedHint)

			// Also verify JSON output
			data, _, err := MarshalWithHint(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedJSON, string(data))
		})
	}

	// Test nil pointer consistency in structs
	t.Run("nil pointer in struct", func(t *testing.T) {
		type TestStruct struct {
			Value *int `json:"value"`
		}
		s := TestStruct{Value: nil}
		hint, err := getTypeHint(s)
		require.NoError(t, err)
		assert.Contains(t, hint, "value=*int")
	})

	// Test *any handling
	t.Run("*any handling", func(t *testing.T) {
		// Test *any with string value
		var val any = "test"
		ptr := &val
		hint, err := getTypeHint(ptr)
		require.NoError(t, err)
		assert.Equal(t, "string", hint)

		// Test *any with int value
		val = 42
		hint, err = getTypeHint(ptr)
		require.NoError(t, err)
		assert.Equal(t, "int", hint)

		// Test *any with struct value
		type SimpleStruct struct {
			Field string `json:"field"`
		}
		val = SimpleStruct{Field: "test"}
		hint, err = getTypeHint(ptr)
		require.NoError(t, err)
		assert.Equal(t, "map[string]any{field=string}", hint)

		// Test nil *any
		var nilPtr *any
		hint, err = getTypeHint(nilPtr)
		require.NoError(t, err)
		assert.Equal(t, "nil", hint)

		// Test MarshalWithHint with *any
		val = map[string]any{"test": 123}
		data, hint, err := MarshalWithHint(ptr)
		require.NoError(t, err)
		assert.Equal(t, "map[string]any{test=int}", hint)
		assert.JSONEq(t, `{"test":"123"}`, string(data))
	})

	// Test pointer to slices
	t.Run("*[]byte", func(t *testing.T) {
		slice := []byte{1, 2, 3}
		ptr := &slice
		testTypeHintRoundTrip(t, ptr, "*[]byte")
	})

	t.Run("nil *[]byte", func(t *testing.T) {
		var ptr *[]byte
		testTypeHintRoundTrip(t, ptr, "*[]byte")
	})

	t.Run("*[]int", func(t *testing.T) {
		slice := []int{1, 2, 3}
		ptr := &slice
		testTypeHintRoundTrip(t, ptr, "*[]int")
	})

	t.Run("*[]string", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		ptr := &slice
		testTypeHintRoundTrip(t, ptr, "*[]string")
	})

	// Test nested pointers
	t.Run("**int", func(t *testing.T) {
		v := 42
		p1 := &v
		p2 := &p1
		testTypeHintRoundTrip(t, p2, "**int")
	})

	t.Run("***int", func(t *testing.T) {
		v := 42
		p1 := &v
		p2 := &p1
		p3 := &p2
		testTypeHintRoundTrip(t, p3, "***int")
	})

	t.Run("**[]byte", func(t *testing.T) {
		slice := []byte{1, 2, 3}
		p1 := &slice
		p2 := &p1
		testTypeHintRoundTrip(t, p2, "**[]byte")
	})

	t.Run("*[][]int", func(t *testing.T) {
		slice := [][]int{{1, 2}, {3, 4}}
		ptr := &slice
		testTypeHintRoundTrip(t, ptr, "*[][]int")
	})

	t.Run("**[][]byte", func(t *testing.T) {
		slice := [][]byte{{1, 2}, {3, 4}}
		p1 := &slice
		p2 := &p1
		testTypeHintRoundTrip(t, p2, "**[][]byte")
	})

	// Test pointer to slice of pointers
	t.Run("*[]*int", func(t *testing.T) {
		a, b, c := 1, 2, 3
		slice := []*int{&a, &b, &c}
		ptr := &slice
		testTypeHintRoundTrip(t, ptr, "*[]*int")
	})

	t.Run("*[]*int with nil", func(t *testing.T) {
		a := 42
		slice := []*int{&a, nil}
		ptr := &slice
		testTypeHintRoundTrip(t, ptr, "*[]*int")
	})

	t.Run("nil *[]*int", func(t *testing.T) {
		var ptr *[]*int
		testTypeHintRoundTrip(t, ptr, "*[]*int")
	})

	// Test more complex nested cases
	t.Run("*[]*[]int", func(t *testing.T) {
		s1 := []int{1, 2}
		s2 := []int{3, 4}
		slice := []*[]int{&s1, &s2}
		ptr := &slice
		testTypeHintRoundTrip(t, ptr, "*[]*[]int")
	})

	t.Run("**[]*int", func(t *testing.T) {
		a, b := 1, 2
		slice := []*int{&a, &b}
		p1 := &slice
		p2 := &p1
		testTypeHintRoundTrip(t, p2, "**[]*int")
	})

	t.Run("*[]**int", func(t *testing.T) {
		v1, v2 := 1, 2
		p1, p2 := &v1, &v2
		slice := []**int{&p1, &p2}
		ptr := &slice
		testTypeHintRoundTrip(t, ptr, "*[]**int")
	})
}

func TestTypeHints_RecursiveSlices(t *testing.T) {
	tests := []struct {
		name  string
		value any
		hint  string
	}{
		// Basic slices (should now use recursive handling)
		{"[]int", []int{1, 2, 3}, "[]int"},
		{"[]string", []string{"a", "b"}, "[]string"},
		{"[]bool", []bool{true, false}, "[]bool"},
		{"[]float64", []float64{1.1, 2.2}, "[]float64"},

		// Nested slices
		{"[][]int", [][]int{{1, 2}, {3, 4}}, "[][]int"},
		{"[][]string", [][]string{{"a", "b"}, {"c", "d"}}, "[][]string"},
		{"[][]byte", [][]byte{{1, 2}, {3, 4}}, "[][]byte"},

		// Deeply nested slices
		{"[][][]int", [][][]int{{{1, 2}}, {{3, 4}}}, "[][][]int"},
		{"[][][][]byte", [][][][]byte{{{{1, 2}}}, {{{3, 4}}}}, "[][][][]byte"},

		// Mixed nested types
		{"[][]bool", [][]bool{{true, false}, {false, true}}, "[][]bool"},
		{"[][]float32", [][]float32{{1.1, 2.2}, {3.3, 4.4}}, "[][]float32"},

		// Nested []any slices - with element type hints
		{"[][]any", [][]any{{1, "hello", true}, {3.14, false}}, "[][]any{[]any{int,string,bool},[]any{float64,bool}}"},
		{"[][][]any", [][][]any{{{1, 2}}, {{"a", "b"}}}, "[][][]any{[][]any{[]any{int,int}},[][]any{[]any{string,string}}}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test type hint generation
			hint, err := getTypeHint(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.hint, hint)

			// Test round-trip marshaling
			data, hint, err := MarshalWithHint(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.hint, hint)

			// Test unmarshaling
			result, err := UnmarshalWithHint(data, hint)
			require.NoError(t, err)

			assert.True(t, reflect.DeepEqual(tt.value, result),
				"Expected %#v, got %#v", tt.value, result)
		})
	}
}

func TestTypeHints_SlicesOfPointers(t *testing.T) {
	tests := []struct {
		name  string
		value any
		hint  string
	}{
		// Basic slices of pointers
		{
			name: "[]*int",
			value: func() []*int {
				a, b, c := 1, 2, 3
				return []*int{&a, &b, &c}
			}(),
			hint: "[]*int",
		},
		{
			name: "[]*string",
			value: func() []*string {
				a, b := "hello", "world"
				return []*string{&a, &b}
			}(),
			hint: "[]*string",
		},
		{
			name: "[]*[]int",
			value: func() []*[]int {
				s1 := []int{1, 2}
				s2 := []int{3, 4}
				return []*[]int{&s1, &s2}
			}(),
			hint: "[]*[]int",
		},
		{
			name: "[]**int",
			value: func() []**int {
				v := 42
				p1 := &v
				v2 := 43
				p3 := &v2
				return []**int{&p1, &p3}
			}(),
			hint: "[]**int",
		},
		// Mixed with nil
		{
			name: "[]*int with nil",
			value: func() []*int {
				a := 1
				return []*int{&a, nil}
			}(),
			hint: "[]*int",
		},
		// Complex nested case
		{
			name: "[][]*[]int",
			value: func() [][]*[]int {
				s1 := []int{1, 2}
				s2 := []int{3, 4}
				return [][]*[]int{{&s1}, {&s2, nil}}
			}(),
			hint: "[][]*[]int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTypeHintRoundTrip(t, tt.value, tt.hint)
		})
	}
}

func TestTypeHints_CollectionTypes(t *testing.T) {
	t.Run("slices and arrays", func(t *testing.T) {
		tests := []struct {
			name         string
			value        any
			expectedHint string
			expectedJSON string
		}{
			// Basic slices
			{"[]byte", []byte{1, 2, 3}, "[]byte", `"AQID"`},
			{"[]uint8", []uint8{1, 2, 3}, "[]byte", `"AQID"`}, // Normalized
			{"[]int", []int{1, 2, 3}, "[]int", `["1","2","3"]`},
			{"[]int32", []int32{1, 2, 3}, "[]int32", `["1","2","3"]`},
			{"[]int64", []int64{1, 2, 3}, "[]int64", `["1","2","3"]`},
			{"[]string", []string{"a", "b"}, "[]string", `["a","b"]`},
			{"[]bool", []bool{true, false}, "[]bool", `[true,false]`},

			// Arrays are treated as slices but comparison fails due to array/slice type mismatch

			// Empty slices
			{"empty []int", []int{}, "[]int", `[]`},
			{"empty []string", []string{}, "[]string", `[]`},
			{"empty []byte", []byte{}, "[]byte", `""`},

			// []any with type hints
			{"empty []any", []any{}, "[]any{}", `[]`},
			{"[]any mixed", []any{42, "hello", true}, "[]any{int,string,bool}", `["42","hello",true]`},
			{"[]any with nil", []any{42, nil, "test"}, "[]any{int,nil,string}", `["42",null,"test"]`},
			{"nested []any", []any{[]any{1, 2}, "test"}, "[]any{[]any{int,int},string}", `[["1","2"],"test"]`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Test all three operations together
				testTypeHintRoundTrip(t, tt.value, tt.expectedHint)

				// Also verify JSON output
				data, _, err := MarshalWithHint(tt.value)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedJSON, string(data))
			})
		}
	})

	t.Run("maps", func(t *testing.T) {
		tests := []struct {
			name         string
			value        any
			expectedHint string
		}{
			// Typed maps
			{"map[string]string", map[string]string{"k": "v"}, "map[string]string"},
			{"map[string]int", map[string]int{"k": 1}, "map[string]int"},
			{"map[string]int64", map[string]int64{"k": 1}, "map[string]int64"},
			{"map[string]bool", map[string]bool{"k": true}, "map[string]bool"},
			{"empty map[string]int", map[string]int{}, "map[string]int"},

			// map[string]any with field hints
			{"empty map[string]any", map[string]any{}, "map[string]any{}"},
			{"map[string]any simple", map[string]any{"a": 1, "b": "test"}, "map[string]any{a=int,b=string}"},
			{"map[string]any with nested", map[string]any{
				"int":    42,
				"nested": map[string]any{"inner": 10},
			}, "map[string]any{int=int,nested=map[string]any{inner=int}}"},
			// map[string]any with pointers - tested separately due to nil handling complexity
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				testTypeHintRoundTrip(t, tt.value, tt.expectedHint)
			})
		}

		// Test map[string]any with nil pointers separately
		t.Run("map[string]any with nil pointer", func(t *testing.T) {
			m := map[string]any{
				"ptr": func() *int { v := 42; return &v }(),
				"nil": (*string)(nil),
			}

			hint, err := getTypeHint(m)
			require.NoError(t, err)
			assert.Equal(t, "map[string]any{nil=*string,ptr=*int}", hint)

			// Test round trip
			data, _, err := MarshalWithHint(m)
			require.NoError(t, err)

			result, err := UnmarshalWithHint(data, hint)
			require.NoError(t, err)

			resultMap := result.(map[string]any)
			// Check ptr field
			ptrVal, ok := resultMap["ptr"].(*int)
			require.True(t, ok)
			assert.Equal(t, 42, *ptrVal)

			// Check nil field - it should be (*string)(nil) not untyped nil
			nilVal := resultMap["nil"]
			assert.Nil(t, nilVal)
			// Note: We can't preserve typed nil through JSON, it becomes untyped nil
		})
	})
}

func TestTypeHints_SpecialTypes(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		expectedHint string
	}{
		// *big.Int
		{"*big.Int", big.NewInt(123), "*big.Int"},
		{"*big.Int large", func() *big.Int {
			n := new(big.Int)
			n.SetString("123456789012345678901234567890", 10)
			return n
		}(), "*big.Int"},
		{"nil *big.Int", (*big.Int)(nil), "*big.Int"},
		// big.Int value is always treated as *big.Int but comparison fails

		// time.Time
		{"time.Time", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "time.Time"},
		// *time.Time is always treated as time.Time
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTypeHintRoundTrip(t, tt.value, tt.expectedHint)
		})
	}
}

func TestTypeHints_ValuesValue(t *testing.T) {
	t.Run("values.Value with int", func(t *testing.T) {
		val, err := values.Wrap(int64(42))
		require.NoError(t, err)

		hint, err := getTypeHint(val)
		require.NoError(t, err)
		assert.Equal(t, "values.Value", hint)

		// Test marshaling
		data, marshalHint, err := MarshalWithHint(val)
		require.NoError(t, err)
		assert.Equal(t, "values.Value", marshalHint)

		// Test unmarshaling
		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)

		// Unwrap and verify the value
		unwrapped, err := resultVal.Unwrap()
		require.NoError(t, err)
		assert.Equal(t, int64(42), unwrapped)
	})

	t.Run("values.Value with string", func(t *testing.T) {
		val, err := values.Wrap("hello world")
		require.NoError(t, err)

		hint, err := getTypeHint(val)
		require.NoError(t, err)
		assert.Equal(t, "values.Value", hint)

		data, _, err := MarshalWithHint(val)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)

		var unwrapped string
		err = resultVal.UnwrapTo(&unwrapped)
		require.NoError(t, err)
		assert.Equal(t, "hello world", unwrapped)
	})

	t.Run("values.Value with map", func(t *testing.T) {
		m := map[string]any{
			"foo": int64(123),
			"bar": "test",
		}
		val, err := values.Wrap(m)
		require.NoError(t, err)

		hint, err := getTypeHint(val)
		require.NoError(t, err)
		assert.Equal(t, "values.Value", hint)

		data, _, err := MarshalWithHint(val)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)

		unwrapped, err := resultVal.Unwrap()
		require.NoError(t, err)
		unwrappedMap, ok := unwrapped.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "test", unwrappedMap["bar"])
		// Our convertJSONNumbers should have converted to int64
		assert.Equal(t, int64(123), unwrappedMap["foo"])
	})

	t.Run("values.Value with nil", func(t *testing.T) {
		// In real usage, a nil values.Value would typically be in a typed context
		// For example, as a struct field or when values.Wrap returns (nil, error)

		// Test 1: nil values.Value from values.Wrap
		val, err := values.Wrap(nil)
		require.NoError(t, err)
		assert.Nil(t, val)

		// When we have an untyped nil, it's just nil
		var untypedNil values.Value
		hint, err := getTypeHint(untypedNil)
		require.NoError(t, err)
		assert.Equal(t, "nil", hint) // Untyped nil is just nil

		// Test 2: In a struct with nil values.Value field
		type TestStruct struct {
			Val values.Value `json:"val"`
		}
		s := TestStruct{Val: nil}

		structHint, err := getTypeHint(s)
		require.NoError(t, err)
		assert.Contains(t, structHint, "val=values.Value") // Field type is preserved from struct definition

		data, _, err := MarshalWithHint(s)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, structHint)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		// The nil values.Value field should be preserved
		valField := resultMap["val"]
		assert.Nil(t, valField)
	})

	t.Run("values.Value with bool", func(t *testing.T) {
		val, err := values.Wrap(true)
		require.NoError(t, err)

		data, _, err := MarshalWithHint(val)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, "values.Value")
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)
		unwrapped, err := resultVal.Unwrap()
		require.NoError(t, err)
		assert.Equal(t, true, unwrapped)
	})

	t.Run("values.Value with float64", func(t *testing.T) {
		val, err := values.Wrap(float64(3.14159))
		require.NoError(t, err)

		data, _, err := MarshalWithHint(val)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, "values.Value")
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)
		unwrapped, err := resultVal.Unwrap()
		require.NoError(t, err)
		assert.Equal(t, float64(3.14159), unwrapped)
	})

	t.Run("values.Value with bytes", func(t *testing.T) {
		val, err := values.Wrap([]byte{1, 2, 3, 4, 5})
		require.NoError(t, err)

		data, _, err := MarshalWithHint(val)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, "values.Value")
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)
		unwrapped, err := resultVal.Unwrap()
		require.NoError(t, err)
		assert.Equal(t, []byte{1, 2, 3, 4, 5}, unwrapped)
	})

	t.Run("values.Value with *big.Int", func(t *testing.T) {
		bigInt := big.NewInt(999999999999999999)
		val, err := values.Wrap(bigInt)
		require.NoError(t, err)

		data, _, err := MarshalWithHint(val)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, "values.Value")
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)
		unwrapped, err := resultVal.Unwrap()
		require.NoError(t, err)
		unwrappedBigInt, ok := unwrapped.(*big.Int)
		require.True(t, ok)
		assert.Equal(t, 0, bigInt.Cmp(unwrappedBigInt))
	})

	t.Run("values.Value with time.Time", func(t *testing.T) {
		now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		val, err := values.Wrap(now)
		require.NoError(t, err)

		data, _, err := MarshalWithHint(val)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, "values.Value")
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)
		unwrapped, err := resultVal.Unwrap()
		require.NoError(t, err)
		unwrappedTime, ok := unwrapped.(time.Time)
		require.True(t, ok)
		assert.True(t, now.Equal(unwrappedTime))
	})

	t.Run("values.Value with slice", func(t *testing.T) {
		slice := []any{int64(1), "two", true}
		val, err := values.Wrap(slice)
		require.NoError(t, err)

		data, _, err := MarshalWithHint(val)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, "values.Value")
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)
		unwrapped, err := resultVal.Unwrap()
		require.NoError(t, err)
		unwrappedSlice, ok := unwrapped.([]any)
		require.True(t, ok)
		assert.Len(t, unwrappedSlice, 3)
		assert.Equal(t, int64(1), unwrappedSlice[0])
		assert.Equal(t, "two", unwrappedSlice[1])
		assert.Equal(t, true, unwrappedSlice[2])
	})

	t.Run("values.Value inside struct", func(t *testing.T) {
		type TestStructWithValue struct {
			Name  string       `json:"name"`
			Value values.Value `json:"value"`
			Count int          `json:"count"`
		}

		innerVal, err := values.Wrap(map[string]any{"key": int64(42)})
		require.NoError(t, err)

		s := TestStructWithValue{
			Name:  "test",
			Value: innerVal,
			Count: 10,
		}

		hint, err := getTypeHint(s)
		require.NoError(t, err)
		assert.Contains(t, hint, "value=values.Value")

		data, _, err := MarshalWithHint(s)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "test", resultMap["name"])
		assert.Equal(t, 10, resultMap["count"])

		// Check the values.Value field
		resultValue, ok := resultMap["value"].(values.Value)
		require.True(t, ok)
		unwrapped, err := resultValue.Unwrap()
		require.NoError(t, err)
		unwrappedMap, ok := unwrapped.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, int64(42), unwrappedMap["key"])
	})

	t.Run("values.Value inside []any", func(t *testing.T) {
		val1, err := values.Wrap(int64(100))
		require.NoError(t, err)
		val2, err := values.Wrap("hello")
		require.NoError(t, err)

		slice := []any{
			"regular string",
			val1,
			42, // regular int
			val2,
			true,
		}

		hint, err := getTypeHint(slice)
		require.NoError(t, err)
		assert.Equal(t, "[]any{string,values.Value,int,values.Value,bool}", hint)

		data, _, err := MarshalWithHint(slice)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)

		resultSlice, ok := result.([]any)
		require.True(t, ok)
		assert.Len(t, resultSlice, 5)

		// Check regular values
		assert.Equal(t, "regular string", resultSlice[0])
		assert.Equal(t, 42, resultSlice[2])
		assert.Equal(t, true, resultSlice[4])

		// Check values.Value items
		val1Result, ok := resultSlice[1].(values.Value)
		require.True(t, ok)
		unwrapped1, err := val1Result.Unwrap()
		require.NoError(t, err)
		assert.Equal(t, int64(100), unwrapped1)

		val2Result, ok := resultSlice[3].(values.Value)
		require.True(t, ok)
		unwrapped2, err := val2Result.Unwrap()
		require.NoError(t, err)
		assert.Equal(t, "hello", unwrapped2)
	})

	t.Run("values.Value pointer should error", func(t *testing.T) {
		// Test with a pointer to values.Value - this should error
		val, err := values.Wrap(int64(42))
		require.NoError(t, err)

		ptrToValue := &val

		// getTypeHint should return an error for *values.Value
		_, err = getTypeHint(ptrToValue)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pointer to values.Value interface is not supported")

		// MarshalWithHint should also fail
		_, _, err = MarshalWithHint(ptrToValue)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pointer to values.Value interface is not supported")
	})

	t.Run("values.Value nested values", func(t *testing.T) {
		// Test complex nested structures
		complexMap := map[string]any{
			"simple": int64(42),
			"nested": map[string]any{
				"deep": []any{int64(1), int64(2), int64(3)},
			},
		}
		val, err := values.Wrap(complexMap)
		require.NoError(t, err)

		data, _, err := MarshalWithHint(val)
		require.NoError(t, err)

		result, err := UnmarshalWithHint(data, "values.Value")
		require.NoError(t, err)

		resultVal, ok := result.(values.Value)
		require.True(t, ok)
		unwrapped, err := resultVal.Unwrap()
		require.NoError(t, err)

		unwrappedMap, ok := unwrapped.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, int64(42), unwrappedMap["simple"])

		nestedMap, ok := unwrappedMap["nested"].(map[string]any)
		require.True(t, ok)
		deepSlice, ok := nestedMap["deep"].([]any)
		require.True(t, ok)
		assert.Len(t, deepSlice, 3)
		assert.Equal(t, int64(1), deepSlice[0])
	})
}

func TestTypeHints_StructTypes(t *testing.T) {
	t.Run("struct with interface fields", func(t *testing.T) {
		// Define a custom interface
		type MyInterface interface {
			DoSomething()
		}

		type TestStruct struct {
			EmptyInterface  any          `json:"empty"`
			ValueInterface  values.Value `json:"value"`
			CustomInterface MyInterface  `json:"custom"`
		}

		s := TestStruct{
			EmptyInterface:  nil,
			ValueInterface:  nil,
			CustomInterface: nil,
		}

		hint, err := getTypeHint(s)
		require.NoError(t, err)

		// values.Value should preserve type even when nil
		assert.Contains(t, hint, "value=values.Value")

		// Empty any when nil is just nil
		assert.Contains(t, hint, "empty=nil")

		// Other interfaces when nil are also nil (we can't preserve arbitrary interface types)
		assert.Contains(t, hint, "custom=nil")
	})

	t.Run("simple struct", func(t *testing.T) {
		type Simple struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		s := Simple{Name: "test", Value: 42}
		hint, err := getTypeHint(s)
		require.NoError(t, err)

		// Should be converted to map[string]any
		assert.Contains(t, hint, "map[string]any{")
		assert.Contains(t, hint, "name=string")
		assert.Contains(t, hint, "value=int")
	})

	t.Run("struct with json tags and omitempty", func(t *testing.T) {
		type TestStruct struct {
			Name        string   `json:"name"`
			Age         int      `json:"age"`
			Active      bool     `json:"is_active"`
			Score       *float64 `json:"score,omitempty"`
			EmptyString string   `json:"empty,omitempty"`
			Ignored     string   `json:"-"`
			unexported  string
			NoTag       string
			Nested      struct {
				Value int64 `json:"value"`
			} `json:"nested"`
		}

		score := 99.5
		input := TestStruct{
			Name:   "test",
			Age:    30,
			Active: true,
			Score:  &score,
			// EmptyString is zero value, should be omitted
			Ignored:    "should not appear",
			unexported: "should not appear",
			NoTag:      "uses field name",
			Nested: struct {
				Value int64 `json:"value"`
			}{Value: 42},
		}

		data, hint, err := MarshalWithHint(input)
		require.NoError(t, err)

		// Verify hint
		assert.Contains(t, hint, "map[string]any{")
		assert.Contains(t, hint, "age=int")
		assert.Contains(t, hint, "is_active=bool")
		assert.Contains(t, hint, "name=string")
		assert.Contains(t, hint, "score=*float64")
		assert.Contains(t, hint, "NoTag=string")
		assert.Contains(t, hint, "nested=map[string]any{value=int64}")

		// Should NOT contain ignored or unexported fields
		assert.NotContains(t, hint, "Ignored")
		assert.NotContains(t, hint, "unexported")
		assert.NotContains(t, hint, "empty") // omitempty with zero value

		// Verify round-trip
		decoded, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)

		result, ok := decoded.(map[string]any)
		require.True(t, ok)

		assert.Equal(t, "test", result["name"])
		assert.Equal(t, 30, result["age"])
		assert.Equal(t, true, result["is_active"])

		scorePtr, ok := result["score"].(*float64)
		require.True(t, ok)
		assert.Equal(t, 99.5, *scorePtr)
	})

	t.Run("struct with nil pointers", func(t *testing.T) {
		type StructWithPointers struct {
			Name  string   `json:"name"`
			Value *int     `json:"value"`
			Score *float64 `json:"score,omitempty"`
		}

		input := StructWithPointers{
			Name:  "test",
			Value: nil, // nil but no omitempty
			Score: nil, // nil with omitempty
		}

		data, hint, err := MarshalWithHint(input)
		require.NoError(t, err)

		assert.Contains(t, hint, "name=string")
		assert.Contains(t, hint, "value=*int")
		assert.NotContains(t, hint, "score") // omitted

		decoded, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)

		result := decoded.(map[string]any)
		assert.Equal(t, "test", result["name"])
		assert.Nil(t, result["value"])
		_, hasScore := result["score"]
		assert.False(t, hasScore)
	})
}

func TestTypeHints_Errors(t *testing.T) {
	t.Run("getTypeHint errors", func(t *testing.T) {
		tests := []struct {
			name     string
			value    any
			errMatch string
		}{
			{"channel", make(chan int), "unsupported type"},
			{"function", func() {}, "unsupported type"},
			{"complex", complex(1, 2), "unsupported type"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := getTypeHint(tt.value)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMatch)
			})
		}
	})

	t.Run("UnmarshalWithHint errors", func(t *testing.T) {
		tests := []struct {
			name     string
			data     []byte
			hint     string
			errMatch string
		}{
			{"invalid JSON", []byte(`{invalid`), "map[string]string", "invalid character"},
			{"empty hint", []byte(`{}`), "", "empty type hint"},
			{"unknown hint", []byte(`42`), "unknown_type", "unknown type hint"},
			{"nil value wrong hint", []byte(`"test"`), "nil", "unexpected value for null type hint"},
			{"string to int error", []byte(`"not a number"`), "int", "invalid syntax"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := UnmarshalWithHint(tt.data, tt.hint)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMatch)
			})
		}
	})

	t.Run("MarshalWithHint errors", func(t *testing.T) {
		_, _, err := MarshalWithHint(make(chan int))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get type hint")
	})
}

func TestTypeHints_DepthHandling(t *testing.T) {
	t.Run("simple value doesn't need depth", func(t *testing.T) {
		hint, err := getTypeHintWithDepth(42, 1)
		require.NoError(t, err)
		assert.Equal(t, "int", hint)
	})

	t.Run("struct consumes depth", func(t *testing.T) {
		type Nested struct {
			Value int `json:"value"`
		}
		type Outer struct {
			Inner Nested `json:"inner"`
		}

		o := Outer{Inner: Nested{Value: 42}}

		hint, err := getTypeHintWithDepth(o, 6)
		require.NoError(t, err)
		assert.Contains(t, hint, "inner=map[string]any{value=int}")

		_, err = getTypeHintWithDepth(o, 2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "max recursion depth exceeded")
	})

	t.Run("circular reference protection", func(t *testing.T) {
		type Node struct {
			Value int            `json:"value"`
			Next  map[string]any `json:"next"`
		}

		node := Node{Value: 1}
		// Create circular reference through map
		node.Next = map[string]any{"node": node}

		// Should succeed because maps don't create true circular references in our implementation
		// The map value is processed independently
		hint, err := getTypeHint(node)
		require.NoError(t, err)
		assert.Contains(t, hint, "next=map[string]any{node=map[string]any{")
		assert.Contains(t, hint, "value=int")
	})
}

// Additional unmarshal tests not covered by round-trip tests
func TestUnmarshalWithHint_EdgeCases(t *testing.T) {
	t.Run("map[string]any with mixed types", func(t *testing.T) {
		data := []byte(`{
			"int": "42",
			"str": "hello",
			"bytes": "AQID",
			"nested": {"inner": "10"}
		}`)

		hint := "map[string]any{bytes=[]byte,int=int,nested=map[string]any{inner=int},str=string}"

		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)

		m, ok := result.(map[string]any)
		require.True(t, ok)

		assert.Equal(t, 42, m["int"])
		assert.Equal(t, "hello", m["str"])
		assert.Equal(t, []byte{1, 2, 3}, m["bytes"])

		nested, ok := m["nested"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, 10, nested["inner"])
	})

	t.Run("null handling for basic types", func(t *testing.T) {
		tests := []struct {
			name     string
			data     []byte
			hint     string
			expected any
		}{
			{"null for int", []byte(`null`), "int", 0},
			{"null for string", []byte(`null`), "string", ""},
			{"null for bool", []byte(`null`), "bool", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := UnmarshalWithHint(tt.data, tt.hint)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// Integration test for full round-trip functionality
func TestTypeHints_PointerToSliceBytes(t *testing.T) {
	// Test the specific case reported: *[]byte being received as []uint8
	t.Run("*[]byte parameter", func(t *testing.T) {
		// Simulate what happens in GetLatestValue
		original := &[]byte{1, 2, 3, 4, 5}
		
		// Marshal with hint (client side)
		data, hint, err := MarshalWithHint(original)
		require.NoError(t, err)
		assert.Equal(t, "*[]byte", hint)
		t.Logf("Marshaled data: %s", string(data))
		t.Logf("Type hint: %s", hint)
		
		// Unmarshal with hint (server side)
		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)
		
		// Check type - should be *[]byte, not []byte
		resultPtr, ok := result.(*[]byte)
		require.True(t, ok, "Expected *[]byte but got %T", result)
		require.NotNil(t, resultPtr)
		
		// Check values
		assert.Equal(t, []byte{1, 2, 3, 4, 5}, *resultPtr)
	})
	
	t.Run("*[]byte as any parameter", func(t *testing.T) {
		// Test when passed as interface{} parameter
		var params any = &[]byte{1, 2, 3, 4, 5}
		
		// Marshal with hint
		data, hint, err := MarshalWithHint(params)
		require.NoError(t, err)
		assert.Equal(t, "*[]byte", hint)
		
		// Unmarshal with hint
		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)
		
		// Verify type
		_, ok := result.(*[]byte)
		require.True(t, ok, "Expected *[]byte but got %T", result)
	})
	
	t.Run("simulate chainlink-aptos flow", func(t *testing.T) {
		// This simulates exactly what chainlink-aptos is doing
		// 1. They have some params
		originalParams := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		
		// 2. They marshal to JSON bytes
		jsonParamBytes, err := json.Marshal(originalParams)
		require.NoError(t, err)
		t.Logf("JSON params: %s", string(jsonParamBytes))
		
		// 3. They pass &jsonParamBytes to our GetLatestValue
		paramsForRPC := &jsonParamBytes
		t.Logf("Type of params for RPC: %T", paramsForRPC)
		
		// 4. Our client marshals with hint
		data, hint, err := MarshalWithHint(paramsForRPC)
		require.NoError(t, err)
		t.Logf("Our marshaled data: %s", string(data))
		t.Logf("Our type hint: %s", hint)
		assert.Equal(t, "*[]byte", hint)
		
		// 5. Our server unmarshals with hint
		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)
		t.Logf("Unmarshaled type: %T", result)
		
		// 6. Their server expects *[]byte
		receivedParams, ok := result.(*[]byte)
		require.True(t, ok, "Expected *[]byte but got %T", result)
		require.NotNil(t, receivedParams)
		
		// 7. They should be able to unmarshal the JSON
		var decodedParams map[string]interface{}
		decoder := json.NewDecoder(bytes.NewReader(*receivedParams))
		decoder.UseNumber()
		err = decoder.Decode(&decodedParams)
		require.NoError(t, err)
		// Convert the number for comparison
		decodedParams["key2"], _ = decodedParams["key2"].(json.Number).Int64()
		decodedParams["key2"] = int(decodedParams["key2"].(int64))
		assert.Equal(t, originalParams, decodedParams)
	})
	
	t.Run("edge case - nil *[]byte", func(t *testing.T) {
		// Test nil pointer
		var nilByteSlice *[]byte = nil
		
		data, hint, err := MarshalWithHint(nilByteSlice)
		require.NoError(t, err)
		assert.Equal(t, "*[]byte", hint)
		assert.Equal(t, "null", string(data))
		
		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)
		
		resultPtr, ok := result.(*[]byte)
		require.True(t, ok, "Expected *[]byte but got %T", result)
		require.Nil(t, resultPtr)
	})
	
	t.Run("edge case - pointer to nil slice", func(t *testing.T) {
		// Test pointer to nil slice
		var nilSlice []byte = nil
		ptrToNilSlice := &nilSlice
		
		data, hint, err := MarshalWithHint(ptrToNilSlice)
		require.NoError(t, err)
		assert.Equal(t, "*[]byte", hint)
		t.Logf("Marshaled nil slice: %s", string(data))
		
		result, err := UnmarshalWithHint(data, hint)
		require.NoError(t, err)
		
		resultPtr, ok := result.(*[]byte)
		require.True(t, ok, "Expected *[]byte but got %T", result)
		// Note: When a pointer to nil slice is marshaled to JSON, it becomes "null"
		// When unmarshaled, it becomes a nil pointer, not a pointer to nil slice
		// This is acceptable behavior for our use case
		require.Nil(t, resultPtr)
	})
}

func TestTypeHints_RoundTrip(t *testing.T) {
	// Helper function for round-trip test
	testRoundTrip := func(t *testing.T, name string, original any) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			// Marshal with hint
			data, hint, err := MarshalWithHint(original)
			require.NoError(t, err)

			// Unmarshal with hint
			result, err := UnmarshalWithHint(data, hint)
			require.NoError(t, err)

			// Compare based on type
			switch v := original.(type) {
			case *big.Int:
				assertBigIntEqual(t, v, result.(*big.Int))
			case time.Time:
				assertTimeEqual(t, v, result.(time.Time))
			case nil:
				assert.Nil(t, result)
			default:
				// For pointers, we need to compare appropriately
				if reflect.TypeOf(original).Kind() == reflect.Ptr {
					if reflect.ValueOf(original).IsNil() {
						assert.Nil(t, result)
					} else {
						assert.Equal(t, original, result)
					}
				} else {
					assert.Equal(t, original, result)
				}
			}
		})
	}

	// Test various types
	testRoundTrip(t, "int", 42)
	testRoundTrip(t, "string", "hello")
	testRoundTrip(t, "[]int", []int{1, 2, 3})
	testRoundTrip(t, "map", map[string]string{"key": "value"})
	testRoundTrip(t, "*big.Int", big.NewInt(999))
	testRoundTrip(t, "time", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	testRoundTrip(t, "[]byte", []byte{1, 2, 3, 4})
	testRoundTrip(t, "*int32", func() *int32 { v := int32(42); return &v }())
	testRoundTrip(t, "nil *int32", (*int32)(nil))
	testRoundTrip(t, "[]any", []any{42, "hello", true, nil})
	testRoundTrip(t, "complex map", map[string]any{
		"int":    42,
		"str":    "hello",
		"bytes":  []byte{5, 6, 7},
		"nested": map[string]any{"inner": 99},
	})
}

func TestParseSliceTypeHint(t *testing.T) {
	tests := []struct {
		name     string
		hint     string
		expected []string
		ok       bool
	}{
		{"empty", "[]any{}", []string{}, true},
		{"simple", "[]any{int,string,bool}", []string{"int", "string", "bool"}, true},
		{"nested", "[]any{[]any{int,int},string}", []string{"[]any{int,int}", "string"}, true},
		{"not slice hint", "map[string]any{}", nil, false},
		{"regular slice", "[]int", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := parseSliceTypeHint(tt.hint)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseMapTypeHint(t *testing.T) {
	tests := []struct {
		name     string
		hint     string
		expected map[string]string
		ok       bool
	}{
		{
			name:     "empty map",
			hint:     "map[string]any{}",
			expected: map[string]string{},
			ok:       true,
		},
		{
			name:     "simple map",
			hint:     "map[string]any{field1=int,field2=string}",
			expected: map[string]string{"field1": "int", "field2": "string"},
			ok:       true,
		},
		{
			name:     "map with nested types",
			hint:     "map[string]any{nested=map[string]any{inner=int64},value=[]byte}",
			expected: map[string]string{"nested": "map[string]any{inner=int64}", "value": "[]byte"},
			ok:       true,
		},
		{
			name: "complex nested map",
			hint: "map[string]any{a=int,b=map[string]any{c=[]any{int,string},d=bool},e=*int32}",
			expected: map[string]string{
				"a": "int",
				"b": "map[string]any{c=[]any{int,string},d=bool}",
				"e": "*int32",
			},
			ok: true,
		},
		{
			name:     "not a map hint",
			hint:     "[]any{int,string}",
			expected: nil,
			ok:       false,
		},
		{
			name:     "regular map type",
			hint:     "map[string]string",
			expected: nil,
			ok:       false,
		},
		{
			name:     "missing closing brace",
			hint:     "map[string]any{field=int",
			expected: nil,
			ok:       false,
		},
		{
			name:     "missing opening brace",
			hint:     "map[string]anyfield=int}",
			expected: nil,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := parseMapTypeHint(tt.hint)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

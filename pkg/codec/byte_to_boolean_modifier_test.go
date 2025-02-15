package codec_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// ptrUint8 is a helper that returns a pointer to a uint8.
func ptrUint8(u uint8) *uint8 {
	return &u
}

// deepEqualPtrTestStruct compares two values of type ptrTestStruct.
func deepEqualPtrTestStruct(a any, b any) bool {
	return reflect.DeepEqual(a, b)
}

func TestBoolToByteModifier(t *testing.T) {
	t.Parallel()

	// on-chain struct: field B is uint8.
	type testStruct struct {
		A string
		B uint8
	}

	// on-chain struct with pointer field.
	type ptrTestStruct struct {
		A string
		B *uint8
	}

	// A struct with an invalid type for conversion.
	type testInvalidStruct struct {
		B string
	}

	// ─── RETYPE TO OFF-CHAIN ─────────────────────────────────────────────

	t.Run("RetypeToOffChain returns error if field type is not convertible", func(t *testing.T) {
		converter := codec.NewByteToBooleanModifier([]string{"B"})
		_, err := converter.RetypeToOffChain(reflect.TypeOf(testInvalidStruct{}), "")
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("RetypeToOffChain converts uint8 to bool", func(t *testing.T) {
		converter := codec.NewByteToBooleanModifier([]string{"B"})
		convertedType, err := converter.RetypeToOffChain(reflect.TypeOf(testStruct{}), "")
		require.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(""), convertedType.Field(0).Type)
		assert.Equal(t, reflect.TypeOf(true), convertedType.Field(1).Type)
	})

	t.Run("RetypeToOffChain converts pointer to uint8 to pointer to bool", func(t *testing.T) {
		converter := codec.NewByteToBooleanModifier([]string{"B"})
		convertedType, err := converter.RetypeToOffChain(reflect.TypeOf(ptrTestStruct{}), "")
		require.NoError(t, err)

		assert.Equal(t, reflect.TypeOf(""), convertedType.Field(0).Type)
		assert.Equal(t, reflect.PointerTo(reflect.TypeOf(true)), convertedType.Field(1).Type)
	})

	t.Run("RetypeToOffChain converts nested field", func(t *testing.T) {
		// inner struct with a uint8 field.
		type inner struct {
			C uint8
		}
		// outer struct with a nested inner.
		type outer struct {
			A string
			B inner
		}
		converter := codec.NewByteToBooleanModifier([]string{"B.C"})
		convertedType, err := converter.RetypeToOffChain(reflect.TypeOf(outer{}), "")
		require.NoError(t, err)
		// Expect a pointer to outer.
		outStruct := convertedType
		require.Equal(t, 2, outStruct.NumField())
		// Field A remains string.
		assert.Equal(t, reflect.TypeOf(""), outStruct.Field(0).Type)
		// Field B is a nested struct; its field C should be retyped from uint8 to bool.
		nestedType := outStruct.Field(1).Type
		assert.Equal(t, reflect.Struct, nestedType.Kind())
		require.Equal(t, 1, nestedType.NumField())
		assert.Equal(t, "C", nestedType.Field(0).Name)
		assert.Equal(t, reflect.TypeOf(true), nestedType.Field(0).Type)
	})

	t.Run("TransformToOnChain converts off-chain bool to on-chain uint8", func(t *testing.T) {
		converter := codec.NewByteToBooleanModifier([]string{"B"})
		convertedType, err := converter.RetypeToOffChain(reflect.TypeOf(testStruct{}), "")
		require.NoError(t, err)

		offchain := reflect.New(convertedType).Elem()
		offchain.FieldByName("A").SetString("example")
		offchain.FieldByName("B").SetBool(true)

		onChainVal, err := converter.TransformToOnChain(offchain.Interface(), "")
		require.NoError(t, err)

		expected := testStruct{A: "example", B: 1}
		assert.Equal(t, expected, onChainVal)
	})

	t.Run("TransformToOnChain converts nested field bool to on-chain uint8", func(t *testing.T) {
		type inner struct {
			C uint8
		}
		type outer struct {
			A string
			B inner
		}
		converter := codec.NewByteToBooleanModifier([]string{"B.C"})
		convertedType, err := converter.RetypeToOffChain(reflect.TypeOf(outer{}), "")
		require.NoError(t, err)

		// Create an off-chain instance.
		offchain := reflect.New(convertedType).Elem()
		offchain.FieldByName("A").SetString("nested")
		nestedStruct := offchain.FieldByName("B")
		// Set nested field C (off-chain) to bool true.
		nestedStruct.FieldByName("C").SetBool(true)

		onChainVal, err := converter.TransformToOnChain(offchain.Interface(), "")
		require.NoError(t, err)
		expected := outer{A: "nested", B: inner{C: 1}}
		assert.Equal(t, expected, onChainVal)
	})

	t.Run("TransformToOffChain converts on-chain uint8 to off-chain bool", func(t *testing.T) {
		converter := codec.NewByteToBooleanModifier([]string{"B"})
		_, err := converter.RetypeToOffChain(reflect.TypeOf(testStruct{}), "")
		require.NoError(t, err)

		// On-chain instance with B=0 (false).
		onChain := testStruct{A: "hello", B: 0}
		offChainVal, err := converter.TransformToOffChain(onChain, "")
		require.NoError(t, err)

		offchain := reflect.ValueOf(offChainVal)
		assert.Equal(t, "hello", offchain.FieldByName("A").String())
		// Field B should now be false.
		assert.False(t, offchain.FieldByName("B").Bool())
	})

	t.Run("TransformToOffChain converts nested field uint8 to off-chain bool", func(t *testing.T) {
		type inner struct {
			F uint8
		}
		type outer struct {
			D string
			E inner
		}
		converter := codec.NewByteToBooleanModifier([]string{"E.F"})
		_, err := converter.RetypeToOffChain(reflect.TypeOf(outer{}), "")
		require.NoError(t, err)

		// On-chain instance with nested field B.C set to 1 (true).
		onChain := outer{D: "off", E: inner{F: 0}}
		offChainVal, err := converter.TransformToOffChain(onChain, "")
		require.NoError(t, err)
		offchain := reflect.ValueOf(offChainVal)
		assert.Equal(t, "off", offchain.FieldByName("D").String())
		// Nested field B.C should now be false.
		nestedStruct := offchain.FieldByName("E")
		assert.False(t, nestedStruct.FieldByName("F").Bool())
	})

	t.Run("TransformToOnChain and TransformToOffChain work on pointer fields", func(t *testing.T) {
		converter := codec.NewByteToBooleanModifier([]string{"B"})
		convertedType, err := converter.RetypeToOffChain(reflect.TypeOf(ptrTestStruct{}), "")
		require.NoError(t, err)

		// Create an off-chain instance.
		offchainPtr := reflect.New(convertedType)
		offchain := offchainPtr.Elem()
		offchain.FieldByName("A").SetString("pointer")
		boolVal := true
		offchain.FieldByName("B").Set(reflect.ValueOf(&boolVal))

		onChainVal, err := converter.TransformToOnChain(offchainPtr.Elem().Interface(), "")
		require.NoError(t, err)

		expected := ptrTestStruct{
			A: "pointer",
			B: ptrUint8(1),
		}
		assert.True(t, deepEqualPtrTestStruct(onChainVal, expected))

		offChainVal2, err := converter.TransformToOffChain(expected, "")
		require.NoError(t, err)
		offchain2 := reflect.ValueOf(offChainVal2)
		assert.Equal(t, "pointer", offchain2.FieldByName("A").String())

		b, ok := offchain2.FieldByName("B").Interface().(*bool)
		require.True(t, ok)
		assert.True(t, *b)
	})
}

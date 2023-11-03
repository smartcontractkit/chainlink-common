package utils_test

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
)

const anyItemTypeForMapDecoder = "anything"

var anyRawBytes = []byte("raw")

func TestMapDecoder(t *testing.T) {
	t.Parallel()
	t.Run("Decode works on a single item", func(t *testing.T) {
		item := &mapTestType{}
		field1 := "Field1"
		anyValue1 := "value1"
		field2 := "Field2"
		anyValue2 := 122
		anySingleResult := map[string]any{field1: anyValue1, field2: anyValue2}
		tmd := &testMapDecoder{resultSingle: anySingleResult}
		decoder, err := utils.DecoderFromMapDecoder(tmd)
		assert.NoError(t, err)

		err = decoder.Decode(context.Background(), anyRawBytes, item, anyItemTypeForMapDecoder)

		assert.NoError(t, err)
		assert.Equal(t, anyValue1, item.Field1)
		assert.Equal(t, anyValue2, item.Field2)
		assertTestMapDecoder(t, tmd)
	})

	t.Run("Decode works on a slice", func(t *testing.T) { runSliceArrayDecodeTest(t, &[]mapTestType{}) })

	t.Run("Decode works on an array", func(t *testing.T) { runSliceArrayDecodeTest(t, &[2]mapTestType{}) })

	t.Run("Decode returns an error if the type is not a pointer", func(t *testing.T) {
		item := mapTestType{}
		field1 := "Field1"
		anyValue1 := "value1"
		field2 := "Field2"
		anyValue2 := 13
		anySingleResult := map[string]any{field1: anyValue1, field2: anyValue2}
		tmd := &testMapDecoder{resultSingle: anySingleResult}
		decoder, err := utils.DecoderFromMapDecoder(tmd)
		assert.NoError(t, err)

		err = decoder.Decode(context.Background(), anyRawBytes, item, anyItemTypeForMapDecoder)

		assert.IsType(t, types.InvalidTypeError{}, err)
	})

	t.Run("Decode returns an error if map is too big for an array", func(t *testing.T) {
		testWrongArraySize(t, &[3]mapTestType{})
	})

	t.Run("Decode returns an error if map is too small for an array", func(t *testing.T) {
		testWrongArraySize(t, &[1]mapTestType{})
	})

	t.Run("Decode returns an error for nil argument", func(t *testing.T) {
		_, err := utils.DecoderFromMapDecoder(nil)
		assert.Error(t, err)
	})
}

func TestVerifyFieldMaps(t *testing.T) {
	t.Parallel()
	anyKey1 := "anything"
	anyKey2 := "different"
	input := map[string]any{
		anyKey1: 1,
		anyKey2: 2,
	}
	t.Run("returns nil if fields match", func(t *testing.T) {
		assert.NoError(t, utils.VerifyFieldMaps([]string{anyKey1, anyKey2}, input))
	})

	t.Run("returns error if field is missing", func(t *testing.T) {
		assert.IsType(t, types.InvalidEncodingError{}, utils.VerifyFieldMaps([]string{anyKey1, anyKey2, "missing"}, input))
	})

	t.Run("returns error for extra key", func(t *testing.T) {
		assert.IsType(t, types.InvalidEncodingError{}, utils.VerifyFieldMaps([]string{anyKey1}, input))
	})

	t.Run("returns error if keys do not match", func(t *testing.T) {
		assert.IsType(t, types.InvalidEncodingError{}, utils.VerifyFieldMaps([]string{anyKey1, "new key"}, input))
	})
}

func TestBigIntHook(t *testing.T) {
	intTypes := []struct {
		Type reflect.Type
		Max  *big.Int
		Min  *big.Int
	}{
		{Type: reflect.TypeOf(0), Min: big.NewInt(math.MinInt), Max: big.NewInt(math.MaxInt)},
		{Type: reflect.TypeOf(uint(0)), Min: big.NewInt(0), Max: new(big.Int).SetUint64(math.MaxUint)},
		{Type: reflect.TypeOf(int8(0)), Min: big.NewInt(math.MinInt8), Max: big.NewInt(math.MaxInt8)},
		{Type: reflect.TypeOf(uint8(0)), Min: big.NewInt(0), Max: new(big.Int).SetUint64(math.MaxUint8)},
		{Type: reflect.TypeOf(int16(0)), Min: big.NewInt(math.MinInt16), Max: big.NewInt(math.MaxInt16)},
		{Type: reflect.TypeOf(uint16(0)), Min: big.NewInt(0), Max: new(big.Int).SetUint64(math.MaxUint16)},
		{Type: reflect.TypeOf(int32(0)), Min: big.NewInt(math.MinInt32), Max: big.NewInt(math.MaxInt32)},
		{Type: reflect.TypeOf(uint32(0)), Min: big.NewInt(0), Max: new(big.Int).SetUint64(math.MaxUint32)},
		{Type: reflect.TypeOf(int64(0)), Min: big.NewInt(math.MinInt64), Max: big.NewInt(math.MaxInt64)},
		{Type: reflect.TypeOf(uint64(0)), Min: big.NewInt(0), Max: new(big.Int).SetUint64(math.MaxUint64)},
	}
	for _, intType := range intTypes {
		t.Run(fmt.Sprintf("Fits conversion %v", intType.Type), func(t *testing.T) {
			anyValidNumber := big.NewInt(5)
			result, err := utils.BigIntHook(reflect.TypeOf((*big.Int)(nil)), intType.Type, anyValidNumber)
			require.NoError(t, err)
			require.IsType(t, reflect.New(intType.Type).Elem().Interface(), result)
			if intType.Min.Cmp(big.NewInt(0)) == 0 {
				u64 := reflect.ValueOf(result).Convert(reflect.TypeOf(uint64(0))).Interface().(uint64)
				actual := new(big.Int).SetUint64(u64)
				require.Equal(t, anyValidNumber, actual)
			} else {
				i64 := reflect.ValueOf(result).Convert(reflect.TypeOf(int64(0))).Interface().(int64)
				actual := big.NewInt(i64)
				require.Equal(t, 0, anyValidNumber.Cmp(actual))
			}
		})

		t.Run("Overflow return an error "+intType.Type.String(), func(t *testing.T) {
			bigger := new(big.Int).Add(intType.Max, big.NewInt(1))
			_, err := utils.BigIntHook(reflect.TypeOf((*big.Int)(nil)), intType.Type, bigger)
			assert.IsType(t, types.InvalidTypeError{}, err)
		})

		t.Run("Underflow return an error "+intType.Type.String(), func(t *testing.T) {
			smaller := new(big.Int).Sub(intType.Min, big.NewInt(1))
			_, err := utils.BigIntHook(reflect.TypeOf((*big.Int)(nil)), intType.Type, smaller)
			assert.IsType(t, types.InvalidTypeError{}, err)
		})

		t.Run("Converts from "+intType.Type.String(), func(t *testing.T) {
			anyValidNumber := int64(5)
			asType := reflect.ValueOf(anyValidNumber).Convert(intType.Type).Interface()
			result, err := utils.BigIntHook(intType.Type, reflect.TypeOf((*big.Int)(nil)), asType)
			require.NoError(t, err)
			bi, ok := result.(*big.Int)
			require.True(t, ok)
			assert.Equal(t, anyValidNumber, bi.Int64())
		})
	}

	t.Run("Converts from string", func(t *testing.T) {
		anyNumber := int64(5)
		result, err := utils.BigIntHook(reflect.TypeOf(""), reflect.TypeOf((*big.Int)(nil)), strconv.FormatInt(anyNumber, 10))
		require.NoError(t, err)
		bi, ok := result.(*big.Int)
		require.True(t, ok)
		assert.Equal(t, anyNumber, bi.Int64())
	})

	t.Run("Converts to string", func(t *testing.T) {
		anyNumber := int64(5)
		result, err := utils.BigIntHook(reflect.TypeOf((*big.Int)(nil)), reflect.TypeOf(""), big.NewInt(anyNumber))
		require.NoError(t, err)
		assert.Equal(t, strconv.FormatInt(anyNumber, 10), result)
	})

	t.Run("Errors for invalid string", func(t *testing.T) {
		_, err := utils.BigIntHook(reflect.TypeOf(""), reflect.TypeOf((*big.Int)(nil)), "Not a number :(")
		require.IsType(t, types.InvalidTypeError{}, err)
	})
}

func runSliceArrayDecodeTest(t *testing.T, item any) {
	field1 := "Field1"
	anyValue11 := "value1"
	anyValue12 := "value12"
	field2 := "Field2"
	anyValue21 := 23
	anyValue22 := 33
	anyManyResult := []map[string]any{
		{field1: anyValue11, field2: anyValue21},
		{field1: anyValue12, field2: anyValue22},
	}
	// template is being used to provide the types
	tmd := &testMapDecoder{resultMany: anyManyResult}
	decoder, err := utils.DecoderFromMapDecoder(tmd)
	assert.NoError(t, err)

	err = decoder.Decode(context.Background(), anyRawBytes, item, anyItemTypeForMapDecoder)
	assert.NoError(t, err)

	rItem := reflect.ValueOf(item).Elem()
	assert.Equal(t, 2, rItem.Len())
	assert.Equal(t, mapTestType{Field1: anyValue11, Field2: anyValue21}, rItem.Index(0).Interface())
	assert.Equal(t, mapTestType{Field1: anyValue12, Field2: anyValue22}, rItem.Index(1).Interface())
}

func testWrongArraySize(t *testing.T, item any) {
	field1 := "Field1"
	anyValue1 := "value1"
	field2 := "Field2"
	anyValue2 := "value2"
	anyManyResult := []map[string]any{
		{field1: anyValue1, field2: anyValue2},
		{field1: anyValue2, field2: anyValue1},
	}
	tmd := &testMapDecoder{resultMany: anyManyResult}
	decoder, err := utils.DecoderFromMapDecoder(tmd)
	assert.NoError(t, err)

	err = decoder.Decode(context.Background(), anyRawBytes, item, anyItemTypeForMapDecoder)
	assert.IsType(t, types.WrongNumberOfElements{}, err)
}

func assertTestMapDecoder(t *testing.T, md *testMapDecoder) {
	assert.True(t, md.correctRaw)
	assert.True(t, md.correctItem)
}

type testMapDecoder struct {
	resultSingle map[string]any
	resultMany   []map[string]any
	correctRaw   bool
	correctItem  bool
}

func (t *testMapDecoder) DecodeSingle(ctx context.Context, raw []byte, itemType string) (map[string]any, error) {
	t.correctRaw = reflect.DeepEqual(raw, anyRawBytes)
	t.correctItem = itemType == anyItemTypeForMapDecoder
	return t.resultSingle, nil
}

func (t *testMapDecoder) DecodeMany(ctx context.Context, raw []byte, itemType string) ([]map[string]any, error) {
	t.correctRaw = reflect.DeepEqual(raw, anyRawBytes)
	t.correctItem = itemType == anyItemTypeForMapDecoder
	return t.resultMany, nil
}

type mapTestType struct {
	Field1 string
	Field2 int
}

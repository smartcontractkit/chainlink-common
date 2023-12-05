package codec_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var anyTestBytes = []byte("any test bytes")

const anyMaxEncodingSize = 100
const anyMaxDecodingSize = 50
const anyItemType = "any item type"
const anySliceItemType = "any slice item type"
const anyNonPointerSliceItemType = "any non pointer slice item type"
const anyValue = 5
const anyForEncoding = true

func TestModifierCodec(t *testing.T) {
	ctx := context.Background()
	mod, err := codec.NewModifierCodec(&testCodec{}, &testTypeProvider{}, testModifier{})
	require.NoError(t, err)

	t.Run("Nil codec returns error", func(t *testing.T) {
		_, err := codec.NewModifierCodec(nil, &testTypeProvider{}, testModifier{})
		assert.Error(t, err)
	})

	t.Run("Nil type provider returns error", func(t *testing.T) {
		_, err := codec.NewModifierCodec(&testCodec{}, nil, testModifier{})
		assert.Error(t, err)
	})

	t.Run("Nil modifier returns error", func(t *testing.T) {
		_, err := codec.NewModifierCodec(&testCodec{}, &testTypeProvider{}, nil)
		assert.Error(t, err)
	})

	t.Run("Encode calls modifiers then encodes", func(t *testing.T) {
		encoded, err := mod.Encode(ctx, &modifierCodecOffChainType{Z: anyValue}, anyItemType)

		require.NoError(t, err)
		assert.Equal(t, anyTestBytes, encoded)
	})

	t.Run("Encode works on compatible types", func(t *testing.T) {
		encoded, err := mod.Encode(ctx, modifierCodecOffChainCompatibleType{Z: anyValue}, anyItemType)

		require.NoError(t, err)
		assert.Equal(t, anyTestBytes, encoded)
	})

	t.Run("Encode works on slices", func(t *testing.T) {
		encoded, err := mod.Encode(ctx, &[]modifierCodecOffChainType{{Z: anyValue}, {Z: anyValue + 1}}, anySliceItemType)

		require.NoError(t, err)
		assert.Equal(t, anyTestBytes, encoded)
	})

	t.Run("Encode works on slices without a pointer", func(t *testing.T) {
		encoded, err := mod.Encode(ctx, []modifierCodecOffChainType{{Z: anyValue}, {Z: anyValue + 1}}, anyNonPointerSliceItemType)

		require.NoError(t, err)
		assert.Equal(t, anyTestBytes, encoded)
	})

	t.Run("Encode works on compatible slices", func(t *testing.T) {
		encoded, err := mod.Encode(ctx, &[]modifierCodecOffChainCompatibleType{{Z: anyValue}, {Z: anyValue + 1}}, anySliceItemType)

		require.NoError(t, err)
		assert.Equal(t, anyTestBytes, encoded)
	})

	t.Run("Decode calls modifiers then decodes", func(t *testing.T) {
		decoded := &modifierCodecOffChainCompatibleType{}
		require.NoError(t, mod.Decode(ctx, anyTestBytes, decoded, anyItemType))
		assert.Equal(t, anyValue, decoded.Z)
	})

	t.Run("Decode works on slices", func(t *testing.T) {
		decoded := &[]modifierCodecOffChainCompatibleType{}
		require.NoError(t, mod.Decode(ctx, anyTestBytes, decoded, anySliceItemType))
		assert.Equal(t, len(*decoded), anyValue)
		for i, d := range *decoded {
			assert.Equal(t, anyValue+i, d.Z)
		}
	})

	t.Run("Encode returns errors from modifiers", func(t *testing.T) {
		_, err = mod.Encode(ctx, &modifierCodecOffChainType{}, "differentType")
		assert.Error(t, err)
	})

	t.Run("Encode returns errors from codec", func(t *testing.T) {
		// test encoder returns error if the value isn't what's expected
		_, err = mod.Encode(ctx, &modifierCodecOffChainType{Z: anyValue + 1}, anyItemType)
		assert.Error(t, err)
	})

	t.Run("Encode returns errors from type converter", func(t *testing.T) {
		_, err = mod.Encode(ctx, &modifierCodecOffChainType{Z: anyValue}, "invalid type")
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("Decode returns errors from codec", func(t *testing.T) {
		assert.Error(t, mod.Decode(ctx, []byte("different"), &modifierCodecOffChainType{}, anyItemType))
	})

	t.Run("Decode returns error for non pointer types", func(t *testing.T) {
		decoded := modifierCodecOffChainType{}
		require.True(t, errors.Is(mod.Decode(ctx, anyTestBytes, decoded, anyItemType), types.ErrInvalidType))
	})

	t.Run("Decode returns error arrays with wrong number of elements", func(t *testing.T) {
		decoded := &[3]modifierCodecOffChainType{}
		require.True(t, errors.Is(mod.Decode(ctx, anyTestBytes, decoded, anySliceItemType), types.ErrWrongNumberOfElements))
	})

	t.Run("Decode returns error for incompatible type", func(t *testing.T) {
		decoded := &modifierCodecOffChainType{}
		require.True(t, errors.Is(mod.Decode(ctx, anyTestBytes, decoded, anySliceItemType), types.ErrInvalidType))
	})

	t.Run("Encode returns errors from type converter", func(t *testing.T) {
		err = mod.Decode(ctx, anyTestBytes, &modifierCodecOffChainType{}, "invalid type")
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("CreateType returns modified type", func(t *testing.T) {
		actual, err := mod.(types.TypeProvider).CreateType(anyItemType, anyForEncoding)
		require.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(&modifierCodecOffChainType{}), reflect.TypeOf(actual))
	})

	t.Run("Create type returns errors from type provides", func(t *testing.T) {
		_, err = mod.(types.TypeProvider).CreateType("differentType", anyForEncoding)
		assert.Equal(t, types.ErrInvalidType, err)
	})

	t.Run("GetMaxEncodingSize delegates", func(t *testing.T) {
		size, err := mod.GetMaxEncodingSize(ctx, anyValue, anyItemType)
		require.NoError(t, err)
		assert.Equal(t, anyMaxEncodingSize, size)
	})

	t.Run("GetMaxDecodingSize delegates", func(t *testing.T) {
		size, err := mod.GetMaxDecodingSize(ctx, anyValue, anyItemType)
		require.NoError(t, err)
		assert.Equal(t, anyMaxDecodingSize, size)
	})
}

type testTypeProvider struct{}

func (t *testTypeProvider) CreateType(itemType string, _ bool) (any, error) {
	switch itemType {
	case anyItemType:
		return &modifierCodecChainType{}, nil
	case anySliceItemType:
		return &[]modifierCodecChainType{}, nil
	case anyNonPointerSliceItemType:
		return []modifierCodecChainType{}, nil
	default:
		return nil, types.ErrInvalidType
	}
}

type modifierCodecChainType struct {
	A int
}

type modifierCodecOffChainType struct {
	Z int
}

type modifierCodecOffChainCompatibleType struct {
	Z int
}

type testCodec struct{}

func (t *testCodec) Encode(_ context.Context, item any, itemType string) ([]byte, error) {
	switch itemType {
	case anyItemType:
		if item.(*modifierCodecChainType).A != anyValue {
			return nil, types.ErrInvalidType
		}
		return anyTestBytes, nil
	case anySliceItemType, anyNonPointerSliceItemType:
		items := item.(*[]modifierCodecChainType)
		for i := 0; i < len(*items); i++ {
			if (*items)[i].A != anyValue+i {
				return nil, types.ErrInvalidType
			}
		}
		return anyTestBytes, nil
	default:
		return nil, types.ErrInvalidType
	}
}

func (t *testCodec) GetMaxEncodingSize(_ context.Context, n int, itemType string) (int, error) {
	if itemType != anyItemType {
		return 0, types.ErrInvalidType
	}

	if n != anyValue {
		return 0, types.ErrUnknown
	}

	return anyMaxEncodingSize, nil
}

func (t *testCodec) Decode(_ context.Context, raw []byte, into any, itemType string) error {
	switch itemType {
	case anyItemType:
		into.(*modifierCodecChainType).A = anyValue
	case anySliceItemType:
		items := make([]modifierCodecChainType, anyValue)
		reflect.Indirect(reflect.ValueOf(into)).Set(reflect.ValueOf(items))
		for i := 0; i < anyValue; i++ {
			items[i].A = anyValue + i
		}
	default:
		return types.ErrInvalidType
	}

	if len(raw) != len(anyTestBytes) {
		return types.ErrInvalidEncoding
	}

	for i, b := range raw {
		if b != anyTestBytes[i] {
			return types.ErrInvalidEncoding
		}
	}

	return nil
}

func (t *testCodec) GetMaxDecodingSize(_ context.Context, n int, itemType string) (int, error) {
	if itemType != anyItemType {
		return 0, types.ErrInvalidType
	}

	if n != anyValue {
		return 0, types.ErrUnknown
	}

	return anyMaxDecodingSize, nil
}

type testModifier struct{}

func (testModifier) RetypeForOffChain(onChainType reflect.Type) (reflect.Type, error) {
	switch onChainType {
	case reflect.TypeOf(&modifierCodecChainType{}):
		return reflect.TypeOf(&modifierCodecOffChainType{}), nil
	case reflect.TypeOf(&[]modifierCodecChainType{}):
		return reflect.TypeOf(&[]modifierCodecOffChainType{}), nil
	case reflect.TypeOf([]modifierCodecChainType{}):
		return reflect.TypeOf([]modifierCodecOffChainType{}), nil
	default:
		return nil, types.ErrInvalidType
	}
}

func (t testModifier) TransformForOnChain(offChainValue any) (any, error) {
	offChain, ok := offChainValue.(*modifierCodecOffChainType)
	if !ok {
		slice, ok := offChainValue.(*[]modifierCodecOffChainType)
		if !ok {
			return nil, types.ErrInvalidType
		}
		if slice == nil {
			return nil, nil
		}
		onChain := make([]modifierCodecChainType, len(*slice))
		for i, v := range *slice {
			onChain[i] = modifierCodecChainType{A: v.Z}
		}
		return &onChain, nil
	}
	return &modifierCodecChainType{A: offChain.Z}, nil
}

func (t testModifier) TransformForOffChain(onChainValue any) (any, error) {
	onChain, ok := onChainValue.(*modifierCodecChainType)
	if !ok {
		slice, ok := onChainValue.(*[]modifierCodecChainType)
		if !ok {
			return nil, types.ErrInvalidType
		}
		if slice == nil {
			return nil, nil
		}
		offChain := make([]modifierCodecOffChainType, len(*slice))
		for i, v := range *slice {
			offChain[i] = modifierCodecOffChainType{Z: v.A}
		}
		return &offChain, nil
	}
	return &modifierCodecOffChainType{Z: onChain.A}, nil
}

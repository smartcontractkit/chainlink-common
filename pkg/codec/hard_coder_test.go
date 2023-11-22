package codec_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
)

var onChainType = reflect.TypeOf(hardCodedTestStruct{})

func TestHardCoder(t *testing.T) {
	hardCoder := codec.NewHardCoder(map[string]any{"A": "Foo", "C": []int32{2, 3}}, map[string]any{"Z": "Bar", "Q": []struct {
		A int
		B string
	}{{1, "a"}, {2, "b"}}})
	replacingHardCoder := codec.NewHardCoder(map[string]any{}, map[string]any{"A": int64(0), "Q": []int32{4, 5}})
	t.Run("RetypeForOffChain adds fields to struct", func(t *testing.T) {
		offChainType, err := hardCoder.RetypeForOffChain(onChainType)
		require.NoError(t, err)
		assertBasicHardCodedType(t, offChainType)
	})

	t.Run("RetypeForOffChain adds fields to pointers", func(t *testing.T) {
		offChainType, err := hardCoder.RetypeForOffChain(reflect.PointerTo(onChainType))
		require.NoError(t, err)
		assert.Equal(t, reflect.Ptr, offChainType.Kind())
		assertBasicHardCodedType(t, offChainType.Elem())
	})

	t.Run("RetypeForOffChain adds fields to pointers of non-structs", func(t *testing.T) {
		offChainType, err := hardCoder.RetypeForOffChain(reflect.PointerTo(reflect.SliceOf(onChainType)))
		require.NoError(t, err)
		assert.Equal(t, reflect.Pointer, offChainType.Kind())
		assert.Equal(t, reflect.Slice, offChainType.Elem().Kind())
		assertBasicHardCodedType(t, offChainType.Elem().Elem())
	})

	t.Run("RetypeForOffChain adds fields to slices", func(t *testing.T) {
		offChainType, err := hardCoder.RetypeForOffChain(reflect.SliceOf(onChainType))
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, offChainType.Kind())
		assertBasicHardCodedType(t, offChainType.Elem())
	})

	t.Run("RetypeForOffChain adds fields to arrays", func(t *testing.T) {
		anyArrayLen := 3
		offChainType, err := hardCoder.RetypeForOffChain(reflect.ArrayOf(anyArrayLen, onChainType))
		require.NoError(t, err)
		assert.Equal(t, reflect.Array, offChainType.Kind())
		assert.Equal(t, anyArrayLen, offChainType.Len())
		assertBasicHardCodedType(t, offChainType.Elem())
	})

	t.Run("RetypeForOffChain replaces already existing field", func(t *testing.T) {
		offChainType, err := replacingHardCoder.RetypeForOffChain(onChainType)
		require.NoError(t, err)
		require.Equal(t, onChainType.NumField()+1, offChainType.NumField())
		for i := 0; i < onChainType.NumField(); i++ {
			if onChainType.Field(i).Name == "A" {
				continue
			}
			require.Equal(t, cleanStructField(onChainType.Field(i)), cleanStructField(offChainType.Field(i)))
		}

		a, ok := offChainType.FieldByName("A")
		require.True(t, ok)
		assert.Equal(t, reflect.TypeOf(int64(0)), a.Type)

		extra := offChainType.Field(onChainType.NumField())
		assert.Equal(t, reflect.StructField{Name: "Q", Type: reflect.TypeOf([]int32{})}, cleanStructField(extra))
	})

	t.Run("TransformForOnChain and TransformForOffChain works on structs", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("TransformForOnChain and TransformForOffChain returns error if input type was not from TransformForOnChain", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("TransformForOnChain and TransformForOffChain works on pointers", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("TransformForOnChain and TransformForOffChain works on slices by creating a new slice and converting elements", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("TransformForOnChain and TransformForOffChain works on pointers to non structs", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("TransformForOnChain and TransformForOffChain works on arrays", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("TransformForOnChain and TransformForOffChain works on nested fields even if the field itself is renamed", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("TransformForOnChain and TransformForOffChain works for replaced type", func(t *testing.T) {
		t.Skip("TODO")
	})
}

func assertBasicHardCodedType(t *testing.T, offChainType reflect.Type) {
	require.Equal(t, onChainType.NumField()+2, offChainType.NumField())
	for i := 0; i < onChainType.NumField(); i++ {
		require.Equal(t, onChainType.Field(i), offChainType.Field(i))
	}

	fn1 := offChainType.Field(onChainType.NumField())
	fn2 := offChainType.Field(onChainType.NumField() + 1)
	var z, q *reflect.StructField
	switch fn1.Name {
	case "Z":
		z = &fn1
	case "Q":
		q = &fn1
	}
	switch fn2.Name {
	case "Z":
		z = &fn2
	case "Q":
		q = &fn2
	}
	require.NotNil(t, z)
	assert.Equal(t, reflect.TypeOf("string"), z.Type)
	require.NotNil(t, q)
	require.Equal(t, reflect.Slice, q.Type.Kind())
	qe := q.Type.Elem()
	require.Equal(t, reflect.Struct, qe.Kind())
	assert.Equal(t, 2, qe.NumField())
	a, ok := qe.FieldByName("A")
	require.True(t, ok)
	assert.Equal(t, reflect.TypeOf(0), a.Type)
	b, ok := qe.FieldByName("B")
	require.True(t, ok)
	assert.Equal(t, reflect.TypeOf("string"), b.Type)
}

type hardCodedTestStruct struct {
	A string
	B int32
	C []int32
}

func cleanStructField(field reflect.StructField) reflect.StructField {
	field.Index = nil
	field.Offset = uintptr(0)
	return field
}

type nestedHardCodedTestStruct struct {
	A string
	B hardCodedTestStruct
	C []hardCodedTestStruct
	D int32
}

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

func TestWrapper(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		A string
		B int64
		C int64
	}

	type nestedTestStruct struct {
		A string
		B testStruct
		C []testStruct
		D string
	}

	wrapper := codec.NewWrapperModifier(map[string]string{"A": "X", "C": "Z"})
	invalidWrapper := codec.NewWrapperModifier(map[string]string{"W": "X", "C": "Z"})
	nestedWrapper := codec.NewWrapperModifier(map[string]string{"A": "X", "B.A": "X", "B.C": "Z", "C.A": "X", "C.C": "Z"})
	t.Run("RetypeToOffChain works on slices", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf([]testStruct{}), "")
		require.NoError(t, err)

		assert.Equal(t, reflect.Slice, offChainType.Kind())
		assertBasicWrapperTransform(t, offChainType.Elem())
	})

	t.Run("RetypeToOffChain works on pointers", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf(&testStruct{}), "")
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, offChainType.Kind())
		assertBasicWrapperTransform(t, offChainType.Elem())
	})

	t.Run("RetypeToOffChain works on pointers to non structs", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf(&[]testStruct{}), "")
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, offChainType.Kind())
		assert.Equal(t, reflect.Slice, offChainType.Elem().Kind())
		assertBasicWrapperTransform(t, offChainType.Elem().Elem())
	})

	t.Run("RetypeToOffChain works on arrays", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf([2]testStruct{}), "")
		require.NoError(t, err)

		assert.Equal(t, reflect.Array, offChainType.Kind())
		assert.Equal(t, 2, offChainType.Len())
		assertBasicWrapperTransform(t, offChainType.Elem())
	})

	t.Run("RetypeToOffChain returns exception if a field is not on the type", func(t *testing.T) {
		_, err := invalidWrapper.RetypeToOffChain(reflect.TypeOf(testStruct{}), "")
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("RetypeToOffChain works on nested fields even if the field itself is already wrapped", func(t *testing.T) {
		offChainType, err := nestedWrapper.RetypeToOffChain(reflect.TypeOf(nestedTestStruct{}), "")
		require.NoError(t, err)
		assert.Equal(t, 4, offChainType.NumField())

		f0 := offChainType.Field(0)
		f0PreRetype := reflect.TypeOf(nestedTestStruct{}).Field(0)
		assert.Equal(t, wrapType("X", f0PreRetype.Type).String(), f0.Type.String())
		assert.Equal(t, "struct { A struct { X string }; B int64; C struct { Z int64 } }", offChainType.Field(1).Type.String())

		f2 := offChainType.Field(2)
		assert.Equal(t, reflect.Slice, f2.Type.Kind())
		assertBasicWrapperTransform(t, f2.Type.Elem())
		f3 := offChainType.Field(3)
		assert.Equal(t, reflect.TypeOf(""), f3.Type)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on structs", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf(testStruct{}), "")
		require.NoError(t, err)
		iOffchain := reflect.Indirect(reflect.New(offChainType))
		iOffchain.FieldByName("A").FieldByName("X").SetString("foo")
		iOffchain.FieldByName("B").SetInt(10)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(20)

		output, err := wrapper.TransformToOnChain(iOffchain.Interface(), "")
		require.NoError(t, err)

		expected := testStruct{
			A: "foo",
			B: 10,
			C: 20,
		}
		assert.Equal(t, expected, output)
		newInput, err := wrapper.TransformToOffChain(expected, "")
		require.NoError(t, err)
		assert.Equal(t, iOffchain.Interface(), newInput)
	})

	t.Run("TransformToOnChain and TransformToOffChain returns error if input type was not from TransformToOnChain", func(t *testing.T) {
		_, err := invalidWrapper.TransformToOnChain(testStruct{}, "")
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformToOnChain and TransformToOffChain works on pointers, but doesn't maintain same addresses", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf(&testStruct{}), "")
		require.NoError(t, err)
		rInput := reflect.New(offChainType.Elem())
		iOffchain := reflect.Indirect(rInput)
		iOffchain.FieldByName("A").FieldByName("X").SetString("foo")
		iOffchain.FieldByName("B").SetInt(10)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(20)

		output, err := wrapper.TransformToOnChain(rInput.Interface(), "")
		require.NoError(t, err)

		expected := &testStruct{
			A: "foo",
			B: 10,
			C: 20,
		}
		assert.Equal(t, expected, output)

		newInput, err := wrapper.TransformToOffChain(output, "")
		require.NoError(t, err)
		assert.Equal(t, rInput.Interface(), newInput)

	})

	t.Run("TransformToOnChain and TransformToOffChain works on slices", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf([]testStruct{}), "")
		require.NoError(t, err)
		rInput := reflect.MakeSlice(offChainType, 2, 2)
		iOffchain := rInput.Index(0)
		iOffchain.FieldByName("A").FieldByName("X").SetString("foo")
		iOffchain.FieldByName("B").SetInt(10)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(20)
		iOffchain = rInput.Index(1)
		iOffchain.FieldByName("A").FieldByName("X").SetString("baz")
		iOffchain.FieldByName("B").SetInt(15)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(25)

		output, err := wrapper.TransformToOnChain(rInput.Interface(), "")

		require.NoError(t, err)

		expected := []testStruct{
			{
				A: "foo",
				B: 10,
				C: 20,
			},
			{
				A: "baz",
				B: 15,
				C: 25,
			},
		}
		assert.Equal(t, expected, output)

		newInput, err := wrapper.TransformToOffChain(expected, "")
		require.NoError(t, err)
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on nested slices", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf([][]testStruct{}), "")
		require.NoError(t, err)
		rInput := reflect.MakeSlice(offChainType, 2, 2)
		rOuter := rInput.Index(0)
		rOuter.Set(reflect.MakeSlice(rOuter.Type(), 2, 2))
		iOffchain := rOuter.Index(0)
		iOffchain.FieldByName("A").FieldByName("X").SetString("foo")
		iOffchain.FieldByName("B").SetInt(10)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(20)
		iOffchain = rOuter.Index(1)
		iOffchain.FieldByName("A").FieldByName("X").SetString("baz")
		iOffchain.FieldByName("B").SetInt(15)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(25)
		rOuter = rInput.Index(1)
		rOuter.Set(reflect.MakeSlice(rOuter.Type(), 2, 2))
		iOffchain = rOuter.Index(0)
		iOffchain.FieldByName("A").FieldByName("X").SetString("fooz")
		iOffchain.FieldByName("B").SetInt(100)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(200)
		iOffchain = rOuter.Index(1)
		iOffchain.FieldByName("A").FieldByName("X").SetString("bazz")
		iOffchain.FieldByName("B").SetInt(150)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(250)

		output, err := wrapper.TransformToOnChain(rInput.Interface(), "")

		require.NoError(t, err)

		expected := [][]testStruct{
			{
				{
					A: "foo",
					B: 10,
					C: 20,
				},
				{
					A: "baz",
					B: 15,
					C: 25,
				},
			},
			{
				{
					A: "fooz",
					B: 100,
					C: 200,
				},
				{
					A: "bazz",
					B: 150,
					C: 250,
				},
			},
		}
		assert.Equal(t, expected, output)

		newInput, err := wrapper.TransformToOffChain(expected, "")
		require.NoError(t, err)
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on pointers to non structs", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf(&[]testStruct{}), "")
		require.NoError(t, err)
		rInput := reflect.New(offChainType.Elem())
		rElm := reflect.MakeSlice(offChainType.Elem(), 2, 2)
		iElm := rElm.Index(0)
		iElm.FieldByName("A").FieldByName("X").SetString("foo")
		iElm.FieldByName("B").SetInt(10)
		iElm.FieldByName("C").FieldByName("Z").SetInt(20)
		iElm = rElm.Index(1)
		iElm.FieldByName("A").FieldByName("X").SetString("baz")
		iElm.FieldByName("B").SetInt(15)
		iElm.FieldByName("C").FieldByName("Z").SetInt(25)
		reflect.Indirect(rInput).Set(rElm)

		output, err := wrapper.TransformToOnChain(rInput.Interface(), "")

		require.NoError(t, err)

		expected := &[]testStruct{
			{
				A: "foo",
				B: 10,
				C: 20,
			},
			{
				A: "baz",
				B: 15,
				C: 25,
			},
		}
		assert.Equal(t, expected, output)

		newInput, err := wrapper.TransformToOffChain(expected, "")
		require.NoError(t, err)
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on arrays", func(t *testing.T) {
		offChainType, err := wrapper.RetypeToOffChain(reflect.TypeOf([2]testStruct{}), "")
		require.NoError(t, err)
		rInput := reflect.New(offChainType).Elem()
		iOffchain := rInput.Index(0)
		iOffchain.FieldByName("A").FieldByName("X").SetString("foo")
		iOffchain.FieldByName("B").SetInt(10)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(20)
		iOffchain = rInput.Index(1)
		iOffchain.FieldByName("A").FieldByName("X").SetString("baz")
		iOffchain.FieldByName("B").SetInt(15)
		iOffchain.FieldByName("C").FieldByName("Z").SetInt(25)

		output, err := wrapper.TransformToOnChain(rInput.Interface(), "")

		require.NoError(t, err)

		expected := [2]testStruct{
			{
				A: "foo",
				B: 10,
				C: 20,
			},
			{
				A: "baz",
				B: 15,
				C: 25,
			},
		}
		assert.Equal(t, expected, output)

		newInput, err := wrapper.TransformToOffChain(expected, "")
		require.NoError(t, err)
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on nested fields even if the field itself was wrapped", func(t *testing.T) {
		offChainType, err := nestedWrapper.RetypeToOffChain(reflect.TypeOf(nestedTestStruct{}), "")
		require.NoError(t, err)

		iOffchain := reflect.Indirect(reflect.New(offChainType))
		iOffchain.FieldByName("A").FieldByName("X").SetString("foo")
		rB := iOffchain.FieldByName("B")
		assert.Equal(t, "struct { A struct { X string }; B int64; C struct { Z int64 } }", offChainType.Field(1).Type.String())

		rB.FieldByName("A").FieldByName("X").SetString("foo")
		rB.FieldByName("B").SetInt(10)
		rB.FieldByName("C").FieldByName("Z").SetInt(20)

		rC := iOffchain.FieldByName("C")
		rC.Set(reflect.MakeSlice(rC.Type(), 2, 2))
		iElm := rC.Index(0)
		iElm.FieldByName("A").FieldByName("X").SetString("foo")
		iElm.FieldByName("B").SetInt(10)
		iElm.FieldByName("C").FieldByName("Z").SetInt(20)
		iElm = rC.Index(1)
		iElm.FieldByName("A").FieldByName("X").SetString("baz")
		iElm.FieldByName("B").SetInt(15)
		iElm.FieldByName("C").FieldByName("Z").SetInt(25)

		iOffchain.FieldByName("D").SetString("bar")

		output, err := nestedWrapper.TransformToOnChain(iOffchain.Interface(), "")
		require.NoError(t, err)

		expected := nestedTestStruct{
			A: "foo",
			B: testStruct{
				A: "foo",
				B: 10,
				C: 20,
			},
			C: []testStruct{
				{
					A: "foo",
					B: 10,
					C: 20,
				},
				{
					A: "baz",
					B: 15,
					C: 25,
				},
			},
			D: "bar",
		}
		assert.Equal(t, expected, output)
		newInput, err := nestedWrapper.TransformToOffChain(expected, "")
		require.NoError(t, err)
		assert.Equal(t, iOffchain.Interface(), newInput)
	})
}

func assertBasicWrapperTransform(t *testing.T, offChainType reflect.Type) {
	require.Equal(t, 3, offChainType.NumField())

	f0 := offChainType.Field(0).Type.Field(0)
	assert.Equal(t, wrapType(f0.Name, f0.Type).String(), offChainType.Field(0).Type.String())

	f1 := offChainType.Field(1)
	assert.Equal(t, reflect.TypeOf(int64(0)), f1.Type)

	f2 := offChainType.Field(2).Type.Field(0)
	assert.Equal(t, wrapType(f2.Name, f2.Type).String(), offChainType.Field(2).Type.String())
}

func wrapType(name string, typ reflect.Type) reflect.Type {
	wrapped := reflect.StructOf([]reflect.StructField{{
		Name: name,
		Type: typ,
	}})
	return wrapped
}

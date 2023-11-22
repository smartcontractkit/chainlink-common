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

func TestRenamer(t *testing.T) {
	renamer := codec.NewRenamer(map[string]string{"A": "X", "C": "Z"})
	invalidRenamer := codec.NewRenamer(map[string]string{"W": "X", "C": "Z"})
	nestedRenamer := codec.NewRenamer(map[string]string{"A": "X", "B.A": "X", "B.C": "Z", "C.A": "X", "C.C": "Z", "B": "Y"})
	t.Run("RetypeForOffChain renames fields keeping structure", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf(renamerTestStruct{}))
		require.NoError(t, err)

		assertBasicRenameTransform(t, inputType)
	})

	t.Run("RetypeForOffChain works on slices", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf([]renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Slice, inputType.Kind())
		assertBasicRenameTransform(t, inputType.Elem())
	})

	t.Run("RetypeForOffChain works on pointers", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf(&renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, inputType.Kind())
		assertBasicRenameTransform(t, inputType.Elem())
	})

	t.Run("RetypeForOffChain works on pointers to non structs", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf(&[]renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, inputType.Kind())
		assert.Equal(t, reflect.Slice, inputType.Elem().Kind())
		assertBasicRenameTransform(t, inputType.Elem().Elem())
	})

	t.Run("RetypeForOffChain works on arrays", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf([2]renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Array, inputType.Kind())
		assert.Equal(t, 2, inputType.Len())
		assertBasicRenameTransform(t, inputType.Elem())
	})

	t.Run("RetypeForOffChain returns exception if a field is not on the type", func(t *testing.T) {
		_, err := invalidRenamer.RetypeForOffChain(reflect.TypeOf(renamerTestStruct{}))
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("RetypeForOffChain works on nested fields even if the field itself is renamed", func(t *testing.T) {
		inputType, err := nestedRenamer.RetypeForOffChain(reflect.TypeOf(renameNesting{}))
		require.NoError(t, err)
		assert.Equal(t, 4, inputType.NumField())
		f0 := inputType.Field(0)
		assert.Equal(t, "X", f0.Name)
		assert.Equal(t, reflect.TypeOf(""), f0.Type)
		f1 := inputType.Field(1)
		assert.Equal(t, "Y", f1.Name)
		assertBasicRenameTransform(t, f1.Type)
		f2 := inputType.Field(2)
		assert.Equal(t, "C", f2.Name)
		assert.Equal(t, reflect.Slice, f2.Type.Kind())
		assertBasicRenameTransform(t, f2.Type.Elem())
		f3 := inputType.Field(3)
		assert.Equal(t, "D", f3.Name)
		assert.Equal(t, reflect.TypeOf(""), f3.Type)
	})

	t.Run("TransformForOnChain and TransformForOffChain works on structs", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf(renamerTestStruct{}))
		require.NoError(t, err)
		iInput := reflect.Indirect(reflect.New(inputType))
		iInput.FieldByName("X").SetString("foo")
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("Z").SetInt(20)

		output, err := renamer.TransformForOnChain(iInput.Interface())

		require.NoError(t, err)

		expected := renamerTestStruct{
			A: "foo",
			B: 10,
			C: 20,
		}
		assert.Equal(t, expected, output)
		newInput, err := renamer.TransformForOffChain(output)
		require.NoError(t, err)
		assert.Equal(t, iInput.Interface(), newInput)
	})

	t.Run("TransformForOnChain and TransformForOffChain returns error if input type was not from TransformForOnChain", func(t *testing.T) {
		_, err := invalidRenamer.TransformForOnChain(renamerTestStruct{})
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformForOnChain and TransformForOffChain works on pointers", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf(&renamerTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType.Elem())
		iInput := reflect.Indirect(rInput)
		iInput.FieldByName("X").SetString("foo")
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("Z").SetInt(20)

		output, err := renamer.TransformForOnChain(rInput.Interface())

		require.NoError(t, err)

		expected := &renamerTestStruct{
			A: "foo",
			B: 10,
			C: 20,
		}
		assert.Equal(t, expected, output)

		// Optimization to avoid creating objects unnecessarily
		iInput.FieldByName("X").SetString("Z")
		expected.A = "Z"
		assert.Equal(t, expected, output)
		newInput, err := renamer.TransformForOffChain(output)
		require.NoError(t, err)
		assert.Same(t, rInput.Interface(), newInput)
	})

	t.Run("TransformForOnChain and TransformForOffChain works on slices by creating a new slice and converting elements", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf([]renamerTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.MakeSlice(inputType, 2, 2)
		iInput := rInput.Index(0)
		iInput.FieldByName("X").SetString("foo")
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("Z").SetInt(20)
		iInput = rInput.Index(1)
		iInput.FieldByName("X").SetString("baz")
		iInput.FieldByName("B").SetInt(15)
		iInput.FieldByName("Z").SetInt(25)

		output, err := renamer.TransformForOnChain(rInput.Interface())

		require.NoError(t, err)

		expected := []renamerTestStruct{
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

		newInput, err := renamer.TransformForOffChain(output)
		require.NoError(t, err)
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformForOnChain and TransformForOffChain works on pointers to non structs", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf(&[]renamerTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType.Elem())
		rElm := reflect.MakeSlice(inputType.Elem(), 2, 2)
		iElm := rElm.Index(0)
		iElm.FieldByName("X").SetString("foo")
		iElm.FieldByName("B").SetInt(10)
		iElm.FieldByName("Z").SetInt(20)
		iElm = rElm.Index(1)
		iElm.FieldByName("X").SetString("baz")
		iElm.FieldByName("B").SetInt(15)
		iElm.FieldByName("Z").SetInt(25)
		reflect.Indirect(rInput).Set(rElm)

		output, err := renamer.TransformForOnChain(rInput.Interface())

		require.NoError(t, err)

		expected := &[]renamerTestStruct{
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

		newInput, err := renamer.TransformForOffChain(output)
		require.NoError(t, err)
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformForOnChain and TransformForOffChain works on arrays", func(t *testing.T) {
		inputType, err := renamer.RetypeForOffChain(reflect.TypeOf([2]renamerTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType).Elem()
		iInput := rInput.Index(0)
		iInput.FieldByName("X").SetString("foo")
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("Z").SetInt(20)
		iInput = rInput.Index(1)
		iInput.FieldByName("X").SetString("baz")
		iInput.FieldByName("B").SetInt(15)
		iInput.FieldByName("Z").SetInt(25)

		output, err := renamer.TransformForOnChain(rInput.Interface())

		require.NoError(t, err)

		expected := [2]renamerTestStruct{
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

		newInput, err := renamer.TransformForOffChain(output)
		require.NoError(t, err)
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformForOnChain and TransformForOffChain works on nested fields even if the field itself is renamed", func(t *testing.T) {
		inputType, err := nestedRenamer.RetypeForOffChain(reflect.TypeOf(renameNesting{}))
		require.NoError(t, err)
		iInput := reflect.Indirect(reflect.New(inputType))

		iInput.FieldByName("X").SetString("foo")
		rY := iInput.FieldByName("Y")
		rY.FieldByName("X").SetString("foo")
		rY.FieldByName("B").SetInt(10)
		rY.FieldByName("Z").SetInt(20)

		rC := iInput.FieldByName("C")
		rC.Set(reflect.MakeSlice(rC.Type(), 2, 2))
		iElm := rC.Index(0)
		iElm.FieldByName("X").SetString("foo")
		iElm.FieldByName("B").SetInt(10)
		iElm.FieldByName("Z").SetInt(20)
		iElm = rC.Index(1)
		iElm.FieldByName("X").SetString("baz")
		iElm.FieldByName("B").SetInt(15)
		iElm.FieldByName("Z").SetInt(25)

		iInput.FieldByName("D").SetString("bar")

		output, err := nestedRenamer.TransformForOnChain(iInput.Interface())

		require.NoError(t, err)

		expected := renameNesting{
			A: "foo",
			B: renamerTestStruct{
				A: "foo",
				B: 10,
				C: 20,
			},
			C: []renamerTestStruct{
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
		newInput, err := nestedRenamer.TransformForOffChain(output)
		require.NoError(t, err)
		assert.Equal(t, iInput.Interface(), newInput)
	})
}

func assertBasicRenameTransform(t *testing.T, inputType reflect.Type) {
	require.Equal(t, 3, inputType.NumField())
	f0 := inputType.Field(0)
	assert.Equal(t, "X", f0.Name)
	assert.Equal(t, reflect.TypeOf(""), f0.Type)
	f1 := inputType.Field(1)
	assert.Equal(t, "B", f1.Name)
	assert.Equal(t, reflect.TypeOf(int64(0)), f1.Type)
	f2 := inputType.Field(2)
	assert.Equal(t, "Z", f2.Name)
	assert.Equal(t, reflect.TypeOf(int64(0)), f2.Type)
}

type renamerTestStruct struct {
	A string
	B int64
	C int64
}

type renameNesting struct {
	A string
	B renamerTestStruct
	C []renamerTestStruct
	D string
}

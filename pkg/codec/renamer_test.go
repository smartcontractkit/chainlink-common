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
	renamer := codec.Renamer{Fields: map[string]string{"A": "X", "C": "Z"}}
	invalidRenamer := codec.Renamer{Fields: map[string]string{"W": "X", "C": "Z"}}
	t.Run("RetypeForInput renames fields keeping structure", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf(renamerTestStruct{}))
		require.NoError(t, err)

		assertBasicTransform(t, inputType)
	})

	t.Run("RetypeForInput works on slices", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf([]renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Slice, inputType.Kind())
		assertBasicTransform(t, inputType.Elem())
	})

	t.Run("RetypeForInput works on pointers", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf(&renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, inputType.Kind())
		assertBasicTransform(t, inputType.Elem())
	})

	t.Run("RetypeForInput works on pointers to non structs", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf(&[]renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, inputType.Kind())
		assert.Equal(t, reflect.Slice, inputType.Elem().Kind())
		assertBasicTransform(t, inputType.Elem().Elem())
	})

	t.Run("RetypeForInput works on arrays", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf([2]renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Array, inputType.Kind())
		assert.Equal(t, 2, inputType.Len())
		assertBasicTransform(t, inputType.Elem())
	})

	t.Run("RetypeForInput returns exception if a field is not on the type", func(t *testing.T) {
		_, err := invalidRenamer.RetypeForInput(reflect.TypeOf(renamerTestStruct{}))
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformInput works on structs", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf(renamerTestStruct{}))
		require.NoError(t, err)
		iInput := reflect.Indirect(reflect.New(inputType))
		iInput.FieldByName("X").SetString("foo")
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("Z").SetInt(20)

		output, err := renamer.TransformInput(iInput.Interface())

		require.NoError(t, err)

		expected := renamerTestStruct{
			A: "foo",
			B: 10,
			C: 20,
		}
		assert.Equal(t, expected, output)
	})

	t.Run("TransformInput returns error if input type was not from TransformInput", func(t *testing.T) {
		_, err := invalidRenamer.TransformInput(renamerTestStruct{})
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformInput works on pointers", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf(renamerTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType)
		iInput := reflect.Indirect(rInput)
		iInput.FieldByName("X").SetString("foo")
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("Z").SetInt(20)

		output, err := renamer.TransformInput(rInput.Interface())

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
	})

	t.Run("TransformInput works on slices by creating a new slice and converting elements", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf([]renamerTestStruct{}))
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

		output, err := renamer.TransformInput(rInput.Interface())

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
	})

	t.Run("TransformInput works on pointers to non structs", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf([]renamerTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType)
		rElm := reflect.MakeSlice(inputType, 2, 2)
		iElm := rElm.Index(0)
		iElm.FieldByName("X").SetString("foo")
		iElm.FieldByName("B").SetInt(10)
		iElm.FieldByName("Z").SetInt(20)
		iElm = rElm.Index(1)
		iElm.FieldByName("X").SetString("baz")
		iElm.FieldByName("B").SetInt(15)
		iElm.FieldByName("Z").SetInt(25)
		reflect.Indirect(rInput).Set(rElm)

		output, err := renamer.TransformInput(rInput.Interface())

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
	})

	t.Run("TransformInput works on arrays", func(t *testing.T) {
		inputType, err := renamer.RetypeForInput(reflect.TypeOf([2]renamerTestStruct{}))
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

		output, err := renamer.TransformInput(rInput.Interface())

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
	})
}

func assertBasicTransform(t *testing.T, inputType reflect.Type) {
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

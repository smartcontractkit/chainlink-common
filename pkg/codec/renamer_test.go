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
	t.Run("AdjustForInput renames fields keeping structure", func(t *testing.T) {
		inputType, err := renamer.AdjustForInput(reflect.TypeOf(renamerTestStruct{}))
		require.NoError(t, err)

		assertBasicTransform(t, inputType)
	})

	t.Run("AdjustForInput works on slices", func(t *testing.T) {
		inputType, err := renamer.AdjustForInput(reflect.TypeOf([]renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Slice, inputType.Kind())
		assertBasicTransform(t, inputType.Elem())
	})

	t.Run("AdjustForInput works on pointers", func(t *testing.T) {
		inputType, err := renamer.AdjustForInput(reflect.TypeOf(&renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, inputType.Kind())
		assertBasicTransform(t, inputType.Elem())
	})

	t.Run("AdjustForInput works on arrays", func(t *testing.T) {
		inputType, err := renamer.AdjustForInput(reflect.TypeOf([2]renamerTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Array, inputType.Kind())
		assert.Equal(t, 2, inputType.Len())
		assertBasicTransform(t, inputType.Elem())
	})

	t.Run("AdjustForInput returns exception if a field is not on the type", func(t *testing.T) {
		_, err := invalidRenamer.AdjustForInput(reflect.TypeOf(renamerTestStruct{}))
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformInput works on structs", func(t *testing.T) {
		inputType, err := renamer.AdjustForInput(reflect.TypeOf(renamerTestStruct{}))
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

	t.Run("TransformInput returns error if input type does not have all the fields", func(t *testing.T) {
		_, err := invalidRenamer.TransformInput(renamerTestStruct{})
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformInput returns struct with renamed fields", func(t *testing.T) {
		inputType, err := renamer.AdjustForInput(reflect.TypeOf(renamerTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType)
		iInput := reflect.Indirect(rInput)
		iInput.FieldByName("X").SetString("foo")
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("Z").SetInt(20)

		output, err := renamer.TransformInput(rInput)

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

	t.Run("TransformInput works on slices", func(t *testing.T) {
		assert.Fail(t, "Not written yet")
	})

	t.Run("TransformInput works on arrays", func(t *testing.T) {
		assert.Fail(t, "Not written yet")
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

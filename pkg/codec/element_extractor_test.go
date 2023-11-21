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

func TestElementExtractor(t *testing.T) {
	extractor := codec.NewElementExtractor(map[string]codec.ElementExtractorLocation{"A": codec.FirstElementLocation, "C": codec.MiddleElementLocation, "D": codec.LastElementLocation})
	invalidExtractor := codec.NewElementExtractor(map[string]codec.ElementExtractorLocation{"A": codec.FirstElementLocation, "W": codec.MiddleElementLocation})
	t.Run("RetypeForInput gets non-slice type", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf(elementExtractorTestStruct{}))
		require.NoError(t, err)

		assertBasicElementExtractTransform(t, inputType)
	})

	t.Run("RetypeForInput works on slices", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf([]elementExtractorTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Slice, inputType.Kind())
		assertBasicElementExtractTransform(t, inputType.Elem())
	})

	t.Run("RetypeForInput works on pointers", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf(&elementExtractorTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, inputType.Kind())
		assertBasicElementExtractTransform(t, inputType.Elem())
	})

	t.Run("RetypeForInput works on pointers to non structs", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf(&[]elementExtractorTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, inputType.Kind())
		assert.Equal(t, reflect.Slice, inputType.Elem().Kind())
		assertBasicElementExtractTransform(t, inputType.Elem().Elem())
	})

	t.Run("RetypeForInput works on arrays", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf([2]elementExtractorTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Array, inputType.Kind())
		assert.Equal(t, 2, inputType.Len())
		assertBasicElementExtractTransform(t, inputType.Elem())
	})

	t.Run("RetypeForInput returns exception if a field is not on the type", func(t *testing.T) {
		_, err := invalidExtractor.RetypeForInput(reflect.TypeOf(elementExtractorTestStruct{}))
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformInput and TransformOutput works on structs", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf(elementExtractorTestStruct{}))
		require.NoError(t, err)
		iInput := reflect.Indirect(reflect.New(inputType))
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 20, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 30}))

		output, err := extractor.TransformInput(iInput.Interface())

		require.NoError(t, err)

		expected := elementExtractorTestStruct{
			A: "A",
			B: 10,
			C: 20,
			D: 30,
		}
		assert.Equal(t, expected, output)
		newInput, err := extractor.TransformOutput(output)
		require.NoError(t, err)
		// Lossy modification
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))
		assert.Equal(t, iInput.Interface(), newInput)
	})

	t.Run("TransformInput and TransformOutput returns error if input type was not from TransformInput", func(t *testing.T) {
		_, err := invalidExtractor.TransformInput(elementExtractorTestStruct{})
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformInput and TransformOutput works on pointers", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf(&elementExtractorTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType.Elem())
		iInput := reflect.Indirect(rInput)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 20, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 30}))

		output, err := extractor.TransformInput(rInput.Interface())

		require.NoError(t, err)

		expected := &elementExtractorTestStruct{
			A: "A",
			B: 10,
			C: 20,
			D: 30,
		}
		assert.Equal(t, expected, output)
		newInput, err := extractor.TransformOutput(output)
		require.NoError(t, err)
		// Lossy modification
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformInput and TransformOutput works on slices by creating a new slice and converting elements", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf([]elementExtractorTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.MakeSlice(inputType, 2, 2)
		iInput := rInput.Index(0)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 20, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 30}))
		iInput = rInput.Index(1)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"Az", "Bz", "Cz"}))
		iInput.FieldByName("B").SetInt(15)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 25, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 35}))

		output, err := extractor.TransformInput(rInput.Interface())

		require.NoError(t, err)

		expected := []elementExtractorTestStruct{
			{
				A: "A",
				B: 10,
				C: 20,
				D: 30,
			},
			{
				A: "Az",
				B: 15,
				C: 25,
				D: 35,
			},
		}
		assert.Equal(t, expected, output)

		newInput, err := extractor.TransformOutput(output)
		require.NoError(t, err)
		// Lossy modification
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"Az"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{25}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{35}))
		iInput = rInput.Index(0)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformInput and TransformOutput works on pointers to non structs", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf([]elementExtractorTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType)
		rElm := reflect.MakeSlice(inputType, 2, 2)
		iInput := rElm.Index(0)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 20, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 30}))
		iInput = rElm.Index(1)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"Az", "Bz", "Cz"}))
		iInput.FieldByName("B").SetInt(15)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 25, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 35}))
		reflect.Indirect(rInput).Set(rElm)

		output, err := extractor.TransformInput(rInput.Interface())

		require.NoError(t, err)

		expected := []elementExtractorTestStruct{
			{
				A: "A",
				B: 10,
				C: 20,
				D: 30,
			},
			{
				A: "Az",
				B: 15,
				C: 25,
				D: 35,
			},
		}
		assert.Equal(t, expected, output)

		newInput, err := extractor.TransformOutput(output)
		require.NoError(t, err)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"Az"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{25}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{35}))
		iInput = rInput.Index(0)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformInput and TransformOutput works on arrays", func(t *testing.T) {
		inputType, err := extractor.RetypeForInput(reflect.TypeOf([2]elementExtractorTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType).Elem()
		iInput := rInput.Index(0)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 20, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 30}))
		iInput = rInput.Index(1)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"Az", "Bz", "Cz"}))
		iInput.FieldByName("B").SetInt(15)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 25, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 35}))

		output, err := extractor.TransformInput(rInput.Interface())

		require.NoError(t, err)

		expected := [2]elementExtractorTestStruct{
			{
				A: "A",
				B: 10,
				C: 20,
				D: 30,
			},
			{
				A: "Az",
				B: 15,
				C: 25,
				D: 35,
			},
		}
		assert.Equal(t, expected, output)

		newInput, err := extractor.TransformOutput(output)
		require.NoError(t, err)
		// Lossy modification
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"Az"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{25}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{35}))
		iInput = rInput.Index(0)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))
		assert.Equal(t, rInput.Interface(), newInput)
	})
}

func assertBasicElementExtractTransform(t *testing.T, inputType reflect.Type) {
	require.Equal(t, 4, inputType.NumField())
	f0 := inputType.Field(0)
	assert.Equal(t, "A", f0.Name)
	assert.Equal(t, reflect.TypeOf([]string{}), f0.Type)
	f1 := inputType.Field(1)
	assert.Equal(t, "B", f1.Name)
	assert.Equal(t, reflect.TypeOf(int64(0)), f1.Type)
	f2 := inputType.Field(2)
	assert.Equal(t, "C", f2.Name)
	assert.Equal(t, reflect.TypeOf([]int64{}), f2.Type)
	f3 := inputType.Field(3)
	assert.Equal(t, "D", f3.Name)
	assert.Equal(t, reflect.TypeOf([]uint64{}), f3.Type)
}

type elementExtractorTestStruct struct {
	A string
	B int64
	C int64
	D uint64
}

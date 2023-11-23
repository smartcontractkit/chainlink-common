package codec_test

import (
	"errors"
	"fmt"
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
	nestedExtractor := codec.NewElementExtractor(map[string]codec.ElementExtractorLocation{"A": codec.FirstElementLocation, "B.A": codec.FirstElementLocation, "B.C": codec.MiddleElementLocation, "B.D": codec.LastElementLocation, "C.A": codec.FirstElementLocation, "C.C": codec.MiddleElementLocation, "C.D": codec.LastElementLocation, "B": codec.LastElementLocation})
	t.Run("RetypeForOffChain gets non-slice type", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf(elementExtractorTestStruct{}))
		require.NoError(t, err)

		assertBasicElementExtractTransform(t, inputType)
	})

	t.Run("RetypeForOffChain works on slices", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf([]elementExtractorTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Slice, inputType.Kind())
		assertBasicElementExtractTransform(t, inputType.Elem())
	})

	t.Run("RetypeForOffChain works on pointers", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf(&elementExtractorTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, inputType.Kind())
		assertBasicElementExtractTransform(t, inputType.Elem())
	})

	t.Run("RetypeForOffChain works on pointers to non structs", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf(&[]elementExtractorTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Pointer, inputType.Kind())
		assert.Equal(t, reflect.Slice, inputType.Elem().Kind())
		assertBasicElementExtractTransform(t, inputType.Elem().Elem())
	})

	t.Run("RetypeForOffChain works on arrays", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf([2]elementExtractorTestStruct{}))
		require.NoError(t, err)

		assert.Equal(t, reflect.Array, inputType.Kind())
		assert.Equal(t, 2, inputType.Len())
		assertBasicElementExtractTransform(t, inputType.Elem())
	})

	t.Run("RetypeForOffChain returns exception if a field is not on the type", func(t *testing.T) {
		_, err := invalidExtractor.RetypeForOffChain(reflect.TypeOf(elementExtractorTestStruct{}))
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("RetypeForOffChain works on nested fields even if the field itself is also extracted", func(t *testing.T) {
		inputType, err := nestedExtractor.RetypeForOffChain(reflect.TypeOf(nestedElementExtractorTestStruct{}))
		fmt.Printf("%+v\n", inputType)
		require.NoError(t, err)
		assert.Equal(t, 4, inputType.NumField())
		f0 := inputType.Field(0)
		assert.Equal(t, "A", f0.Name)
		assert.Equal(t, reflect.TypeOf([]string{}), f0.Type)
		f1 := inputType.Field(1)
		assert.Equal(t, "B", f1.Name)
		require.Equal(t, reflect.Slice, f1.Type.Kind())
		assertBasicElementExtractTransform(t, f1.Type.Elem())
		f2 := inputType.Field(2)
		require.Equal(t, reflect.Slice, f2.Type.Kind())
		assertBasicElementExtractTransform(t, f2.Type.Elem())
		f3 := inputType.Field(3)
		assert.Equal(t, "D", f3.Name)
		assert.Equal(t, reflect.TypeOf(""), f3.Type)
	})

	t.Run("TransformForOnChain and TransformForOffChain works on structs", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf(elementExtractorTestStruct{}))
		require.NoError(t, err)
		iInput := reflect.Indirect(reflect.New(inputType))
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 20, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 30}))

		output, err := extractor.TransformForOnChain(iInput.Interface())

		require.NoError(t, err)

		expected := elementExtractorTestStruct{
			A: "A",
			B: 10,
			C: 20,
			D: 30,
		}
		assert.Equal(t, expected, output)
		newInput, err := extractor.TransformForOffChain(expected)
		require.NoError(t, err)
		// Lossy modification
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))
		assert.Equal(t, iInput.Interface(), newInput)
	})

	t.Run("TransformForOnChain and TransformForOffChain returns error if input type was not from TransformForOnChain", func(t *testing.T) {
		_, err := invalidExtractor.TransformForOnChain(elementExtractorTestStruct{})
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformForOnChain and TransformForOffChain works on pointers", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf(&elementExtractorTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType.Elem())
		iInput := reflect.Indirect(rInput)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))
		iInput.FieldByName("B").SetInt(10)
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{15, 20, 35}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 30}))

		output, err := extractor.TransformForOnChain(rInput.Interface())

		require.NoError(t, err)

		expected := &elementExtractorTestStruct{
			A: "A",
			B: 10,
			C: 20,
			D: 30,
		}
		assert.Equal(t, expected, output)
		newInput, err := extractor.TransformForOffChain(expected)
		require.NoError(t, err)
		// Lossy modification
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformForOnChain and TransformForOffChain works on slices by creating a new slice and converting elements", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf([]elementExtractorTestStruct{}))
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

		output, err := extractor.TransformForOnChain(rInput.Interface())

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

		newInput, err := extractor.TransformForOffChain(expected)
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

	t.Run("TransformForOnChain and TransformForOffChain works on pointers to non structs", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf(&[]elementExtractorTestStruct{}))
		require.NoError(t, err)
		rInput := reflect.New(inputType.Elem())
		rElm := reflect.MakeSlice(inputType.Elem(), 2, 2)
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

		output, err := extractor.TransformForOnChain(rInput.Interface())

		require.NoError(t, err)

		expected := &[]elementExtractorTestStruct{
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

		newInput, err := extractor.TransformForOffChain(expected)
		require.NoError(t, err)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"Az"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{25}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{35}))
		iInput = rElm.Index(0)
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		iInput.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		iInput.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))
		assert.Equal(t, rInput.Interface(), newInput)
	})

	t.Run("TransformForOnChain and TransformForOffChain works on arrays", func(t *testing.T) {
		inputType, err := extractor.RetypeForOffChain(reflect.TypeOf([2]elementExtractorTestStruct{}))
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

		output, err := extractor.TransformForOnChain(rInput.Interface())

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

		newInput, err := extractor.TransformForOffChain(expected)
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

	t.Run("TransformForOnChain and TransformForOffChain works on nested fields even if the field itself is also extracted", func(t *testing.T) {
		inputType, err := nestedExtractor.RetypeForOffChain(reflect.TypeOf(nestedElementExtractorTestStruct{}))
		require.NoError(t, err)

		iInput := reflect.Indirect(reflect.New(inputType))
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))

		rB := iInput.FieldByName("B")
		rB.Set(reflect.MakeSlice(rB.Type(), 2, 2))

		rElm := rB.Index(0)
		rElm.FieldByName("A").Set(reflect.ValueOf([]string{"Z", "W", "Z"}))
		rElm.FieldByName("B").SetInt(99)
		rElm.FieldByName("C").Set(reflect.ValueOf([]int64{44, 44, 44}))
		rElm.FieldByName("D").Set(reflect.ValueOf([]uint64{42, 62, 99}))

		rElm = rB.Index(1)
		rElm.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))
		rElm.FieldByName("B").SetInt(10)
		rElm.FieldByName("C").Set(reflect.ValueOf([]int64{15, 20, 35}))
		rElm.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 30}))

		rC := iInput.FieldByName("C")
		rC.Set(reflect.MakeSlice(rC.Type(), 2, 2))
		iElm := rC.Index(0)
		iElm.FieldByName("A").Set(reflect.ValueOf([]string{"A", "B", "C"}))
		iElm.FieldByName("B").SetInt(10)
		iElm.FieldByName("C").Set(reflect.ValueOf([]int64{15, 20, 35}))
		iElm.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 30}))
		iElm = rC.Index(1)
		iElm.FieldByName("A").Set(reflect.ValueOf([]string{"Az", "Bz", "Cz"}))
		iElm.FieldByName("B").SetInt(15)
		iElm.FieldByName("C").Set(reflect.ValueOf([]int64{15, 25, 35}))
		iElm.FieldByName("D").Set(reflect.ValueOf([]uint64{10, 20, 35}))

		iInput.FieldByName("D").SetString("bar")

		output, err := nestedExtractor.TransformForOnChain(iInput.Interface())
		require.NoError(t, err)

		expected := nestedElementExtractorTestStruct{
			A: "A",
			B: elementExtractorTestStruct{
				A: "A",
				B: 10,
				C: 20,
				D: 30,
			},
			C: []elementExtractorTestStruct{
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
			},
			D: "bar",
		}

		assert.Equal(t, expected, output)

		newInput, err := nestedExtractor.TransformForOffChain(expected)
		require.NoError(t, err)

		// Lossy modification
		iInput.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		rB.Set(rB.Slice(1, 2))
		rElm = rB.Index(0)
		rElm.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		rElm.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		rElm.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))

		rElm = rC.Index(0)
		rElm.FieldByName("A").Set(reflect.ValueOf([]string{"A"}))
		rElm.FieldByName("C").Set(reflect.ValueOf([]int64{20}))
		rElm.FieldByName("D").Set(reflect.ValueOf([]uint64{30}))
		rElm = rC.Index(1)
		rElm.FieldByName("A").Set(reflect.ValueOf([]string{"Az"}))
		rElm.FieldByName("C").Set(reflect.ValueOf([]int64{25}))
		rElm.FieldByName("D").Set(reflect.ValueOf([]uint64{35}))
		assert.Equal(t, iInput.Interface(), newInput)
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

type nestedElementExtractorTestStruct struct {
	A string
	B elementExtractorTestStruct
	C []elementExtractorTestStruct
	D string
}

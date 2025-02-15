package codec_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func TestElementExtractorFromOnchain(t *testing.T) {
	// Use different location options
	first := codec.ElementExtractorLocationFirst
	middle := codec.ElementExtractorLocationMiddle
	last := codec.ElementExtractorLocationLast

	// Basic struct: on-chain the fields are slices; off-chain they become single values.
	type testStruct struct {
		A []int64
		B []string
		C uint64
	}

	type nestedTestStruct struct {
		A []int64
		B testStruct
		C [][]int64
		D string
	}

	extractor := codec.NewElementExtractorFromOnchain(map[string]*codec.ElementExtractorLocation{
		"A": &first,
		"B": &last,
	})

	t.Run("RetypeToOffChain modifies slice fields to single values", func(t *testing.T) {
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf(testStruct{}), "")
		require.NoError(t, err)
		require.Equal(t, 3, newType.NumField())

		fA := newType.Field(0)
		assert.Equal(t, "A", fA.Name)
		assert.Equal(t, reflect.Int64, fA.Type.Kind())

		fB := newType.Field(1)
		assert.Equal(t, "B", fB.Name)
		assert.Equal(t, reflect.String, fB.Type.Kind())

		fC := newType.Field(2)
		assert.Equal(t, "C", fC.Name)
		assert.Equal(t, reflect.Uint64, fC.Type.Kind())
	})

	t.Run("RetypeToOffChain works on slices", func(t *testing.T) {
		// When the input type is a slice, the element type is retyped.
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf([]testStruct{}), "")
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, newType.Kind())
		elemType := newType.Elem()
		require.Equal(t, 3, elemType.NumField())
		fA := elemType.Field(0)
		assert.Equal(t, "A", fA.Name)
		assert.Equal(t, reflect.Int64, fA.Type.Kind())
		fB := elemType.Field(1)
		assert.Equal(t, "B", fB.Name)
		assert.Equal(t, reflect.String, fB.Type.Kind())
		fC := elemType.Field(2)
		assert.Equal(t, "C", fC.Name)
		assert.Equal(t, reflect.Uint64, fC.Type.Kind())
	})

	t.Run("RetypeToOffChain works on pointers", func(t *testing.T) {
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf(&testStruct{}), "")
		require.NoError(t, err)
		assert.Equal(t, reflect.Pointer, newType.Kind())
		elemType := newType.Elem()
		require.Equal(t, 3, elemType.NumField())
		fA := elemType.Field(0)
		assert.Equal(t, "A", fA.Name)
		assert.Equal(t, reflect.Int64, fA.Type.Kind())
	})

	t.Run("RetypeToOffChain works on pointers to slices", func(t *testing.T) {
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf(&[]testStruct{}), "")
		require.NoError(t, err)
		assert.Equal(t, reflect.Pointer, newType.Kind())
		// Underlying type is a slice; retype its element.
		sliceType := newType.Elem()
		assert.Equal(t, reflect.Slice, sliceType.Kind())
		elemType := sliceType.Elem()
		require.Equal(t, 3, elemType.NumField())
		fA := elemType.Field(0)
		assert.Equal(t, "A", fA.Name)
		assert.Equal(t, reflect.Int64, fA.Type.Kind())
	})

	t.Run("RetypeToOffChain works on arrays", func(t *testing.T) {
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf([2]testStruct{}), "")
		require.NoError(t, err)
		assert.Equal(t, reflect.Array, newType.Kind())
		assert.Equal(t, 2, newType.Len())
		elemType := newType.Elem()
		require.Equal(t, 3, elemType.NumField())
		fA := elemType.Field(0)
		assert.Equal(t, "A", fA.Name)
		assert.Equal(t, reflect.Int64, fA.Type.Kind())
	})

	t.Run("RetypeToOffChain returns error if a field is not on the type", func(t *testing.T) {
		// onchain struct should have a field that is a slice
		type badStruct struct {
			X int // X is not a slice.
		}
		badWrapper := codec.NewElementExtractorFromOnchain(map[string]*codec.ElementExtractorLocation{"X": &first})
		_, err := badWrapper.RetypeToOffChain(reflect.TypeOf(badStruct{}), "")
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("RetypeToOffChain works on nested fields even if the field itself is also wrapped", func(t *testing.T) {
		// In nestedTestStruct:
		// - Field A is []int64, so off-chain becomes int64.
		// - Field B is a testStruct; we want to wrap its field A as well.
		// - Field C is [][]int64; off-chain we pick a single int64 from the inner slice.
		// - Field D remains unchanged.
		nestedWrapper := codec.NewElementExtractorFromOnchain(map[string]*codec.ElementExtractorLocation{
			"A":   &first,
			"B.A": &first,
			"C":   &middle,
		})
		newType, err := nestedWrapper.RetypeToOffChain(reflect.TypeOf(nestedTestStruct{}), "")
		require.NoError(t, err)
		require.Equal(t, 4, newType.NumField())
		// Field A retyped: []int64 → int64.
		f0 := newType.Field(0)
		assert.Equal(t, "A", f0.Name)
		assert.Equal(t, reflect.Int64, f0.Type.Kind())
		// Field B is a struct; check its retyped fields.
		f1 := newType.Field(1)
		assert.Equal(t, "B", f1.Name)
		require.Equal(t, reflect.Struct, f1.Type.Kind())
		nestedB := f1.Type
		require.Equal(t, 3, nestedB.NumField()) // testStruct retyped.
		fBA := nestedB.Field(0)
		assert.Equal(t, "A", fBA.Name)
		assert.Equal(t, reflect.Int64, fBA.Type.Kind())
		// Field C: originally [][]int64, now becomes int64.
		f2 := newType.Field(2)
		assert.Equal(t, "C", f2.Name)
		assert.Equal(t, reflect.Slice, f2.Type.Kind())
		// Field D remains string.
		f3 := newType.Field(3)
		assert.Equal(t, "D", f3.Name)
		assert.Equal(t, reflect.TypeOf(""), f3.Type)
	})

	// ─── TRANSFORMATION TESTS ─────────────────────────────────────────

	t.Run("TransformToOnChain wraps single values in a slice", func(t *testing.T) {
		// Off-chain type: testStruct retyped so that A is int64 and B is string.
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf(testStruct{}), "")
		require.NoError(t, err)
		offChainVal := reflect.New(newType).Elem()
		offChainVal.FieldByName("A").SetInt(99)       // off-chain: int64
		offChainVal.FieldByName("B").SetString("xyz") // off-chain: string
		offChainVal.FieldByName("C").SetUint(123)     // unchanged

		// Transform to on-chain: off-chain value is wrapped into slices.
		onChainVal, err := extractor.TransformToOnChain(offChainVal.Interface(), "")
		require.NoError(t, err)
		// Expect on-chain testStruct: A becomes []int64{99}, B becomes []string{"xyz"}, C remains 123.
		expected := testStruct{
			A: []int64{99},
			B: []string{"xyz"},
			C: 123,
		}
		assert.Equal(t, expected, onChainVal)

		// And round-trip back off-chain.
		newOff, err := extractor.TransformToOffChain(onChainVal, "")
		require.NoError(t, err)
		rv := reflect.Indirect(reflect.ValueOf(newOff))
		assert.EqualValues(t, 99, rv.FieldByName("A").Int())
		assert.Equal(t, "xyz", rv.FieldByName("B").String())
		assert.EqualValues(t, 123, rv.FieldByName("C").Uint())
	})

	t.Run("TransformToOnChain and TransformToOffChain works on pointers", func(t *testing.T) {
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf(&testStruct{}), "")
		require.NoError(t, err)
		rInput := reflect.New(newType.Elem())
		iInput := reflect.Indirect(rInput)
		iInput.FieldByName("A").SetInt(50)
		iInput.FieldByName("B").SetString("hello")
		iInput.FieldByName("C").SetUint(999)
		onChainVal, err := extractor.TransformToOnChain(rInput.Interface(), "")
		require.NoError(t, err)
		expected := &testStruct{
			A: []int64{50},
			B: []string{"hello"},
			C: 999,
		}
		assert.Equal(t, expected, onChainVal)
		newOff, err := extractor.TransformToOffChain(expected, "")
		require.NoError(t, err)
		rv := reflect.Indirect(reflect.ValueOf(newOff))
		assert.EqualValues(t, 50, rv.FieldByName("A").Int())
		assert.Equal(t, "hello", rv.FieldByName("B").String())
		assert.EqualValues(t, 999, rv.FieldByName("C").Uint())
	})

	t.Run("TransformToOnChain and TransformToOffChain works on slices", func(t *testing.T) {
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf([]testStruct{}), "")
		require.NoError(t, err)
		rInput := reflect.MakeSlice(newType, 2, 2)
		// Element 0.
		e0 := rInput.Index(0)
		e0.FieldByName("A").SetInt(10)
		e0.FieldByName("B").SetString("first")
		e0.FieldByName("C").SetUint(111)
		// Element 1.
		e1 := rInput.Index(1)
		e1.FieldByName("A").SetInt(20)
		e1.FieldByName("B").SetString("second")
		e1.FieldByName("C").SetUint(222)
		onChainVal, err := extractor.TransformToOnChain(rInput.Interface(), "")
		require.NoError(t, err)
		expected := []testStruct{
			{A: []int64{10}, B: []string{"first"}, C: 111},
			{A: []int64{20}, B: []string{"second"}, C: 222},
		}
		assert.Equal(t, expected, onChainVal)
		newOff, err := extractor.TransformToOffChain(expected, "")
		require.NoError(t, err)
		// The round-trip should recover the original off-chain slice.
		assert.Equal(t, rInput.Interface(), newOff)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on arrays", func(t *testing.T) {
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf([2]testStruct{}), "")
		require.NoError(t, err)
		rInput := reflect.New(newType).Elem()
		e0 := rInput.Index(0)
		e0.FieldByName("A").SetInt(15)
		e0.FieldByName("B").SetString("alpha")
		e0.FieldByName("C").SetUint(333)
		e1 := rInput.Index(1)
		e1.FieldByName("A").SetInt(25)
		e1.FieldByName("B").SetString("beta")
		e1.FieldByName("C").SetUint(444)
		onChainVal, err := extractor.TransformToOnChain(rInput.Interface(), "")
		require.NoError(t, err)
		expected := [2]testStruct{
			{
				A: []int64{15},
				B: []string{"alpha"},
				C: 333,
			},
			{
				A: []int64{25},
				B: []string{"beta"},
				C: 444,
			},
		}
		assert.Equal(t, expected, onChainVal)
		newOff, err := extractor.TransformToOffChain(expected, "")
		require.NoError(t, err)
		// Adjust rInput if needed for lossy round-trip.
		e0.FieldByName("A").SetInt(15)
		e0.FieldByName("B").SetString("alpha")
		e0.FieldByName("C").SetUint(333)
		e1.FieldByName("A").SetInt(25)
		e1.FieldByName("B").SetString("beta")
		e1.FieldByName("C").SetUint(444)
		assert.Equal(t, rInput.Interface(), newOff)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on nested fields", func(t *testing.T) {
		// Create a nested extractor to wrap fields at different depths.
		// For nestedTestStruct:
		//   - "A": []int64 becomes int64 (picking the first element)
		//   - "B.A": inside field B (of type testStruct), the field A (which is []int64)
		//           becomes int64.
		//   - "C": [][]int64 becomes int64 (picking the last element from the inner slice)
		nestedExtractor := codec.NewElementExtractorFromOnchain(map[string]*codec.ElementExtractorLocation{
			"A":   &first,
			"B.A": &first,
			"C":   &last,
		})
		newType, err := nestedExtractor.RetypeToOffChain(reflect.TypeOf(nestedTestStruct{}), "")
		require.NoError(t, err)
		offVal := reflect.New(newType).Elem()
		// Set off-chain values:
		offVal.FieldByName("A").SetInt(100)
		// For nested field B, whose type is retyped from testStruct.
		nestedB := offVal.FieldByName("B")
		nestedB.FieldByName("A").SetInt(50)
		// IMPORTANT: Field B.B is originally []string and remains unchanged (since no mapping was provided).
		// So we must set it with a slice of string.
		nestedB.FieldByName("B").Set(reflect.ValueOf([]string{"nested"}))
		nestedB.FieldByName("C").SetUint(555)
		// For field C (originally [][]int64), off-chain becomes []int64.
		offVal.FieldByName("C").Set(reflect.ValueOf([]int64{200}))
		// Field D is unchanged.
		offVal.FieldByName("D").SetString("end")

		// Transform off-chain → on-chain.
		onChainVal, err := nestedExtractor.TransformToOnChain(offVal.Interface(), "")
		require.NoError(t, err)
		expected := nestedTestStruct{
			A: []int64{100},
			B: testStruct{
				A: []int64{50},
				B: []string{"nested"},
				C: 555,
			},
			C: [][]int64{{200}},
			D: "end",
		}
		assert.Equal(t, expected, onChainVal)

		// Now, transform back from on-chain → off-chain.
		newOff, err := nestedExtractor.TransformToOffChain(expected, "")
		require.NoError(t, err)
		// Due to the lossy nature of the transformation, we adjust offVal to match.
		nestedB.FieldByName("A").SetInt(50)
		nestedB.FieldByName("B").Set(reflect.ValueOf([]string{"nested"}))
		nestedB.FieldByName("C").SetUint(555)
		offVal.FieldByName("A").SetInt(100)
		offVal.FieldByName("C").Set(reflect.ValueOf([]int64{200}))
		offVal.FieldByName("D").SetString("end")
		assert.Equal(t, offVal.Interface(), newOff)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on pointers to non structs", func(t *testing.T) {
		newType, err := extractor.RetypeToOffChain(reflect.TypeOf(&[]testStruct{}), "")
		require.NoError(t, err)
		rInput := reflect.New(newType.Elem())
		rSlice := reflect.MakeSlice(newType.Elem(), 2, 2)
		e0 := rSlice.Index(0)
		e0.FieldByName("A").SetInt(77)
		e0.FieldByName("B").SetString("ptr1")
		e0.FieldByName("C").SetUint(777)
		e1 := rSlice.Index(1)
		e1.FieldByName("A").SetInt(88)
		e1.FieldByName("B").SetString("ptr2")
		e1.FieldByName("C").SetUint(888)
		reflect.Indirect(rInput).Set(rSlice)
		onChainVal, err := extractor.TransformToOnChain(rInput.Interface(), "")
		require.NoError(t, err)
		expected := []testStruct{
			{
				A: []int64{77},
				B: []string{"ptr1"},
				C: 777,
			},
			{
				A: []int64{88},
				B: []string{"ptr2"},
				C: 888,
			},
		}
		assert.Equal(t, &expected, onChainVal)
		newOff, err := extractor.TransformToOffChain(expected, "")
		require.NoError(t, err)
		e0.FieldByName("A").SetInt(77)
		e0.FieldByName("B").SetString("ptr1")
		e0.FieldByName("C").SetUint(777)
		e1.FieldByName("A").SetInt(88)
		e1.FieldByName("B").SetString("ptr2")
		e1.FieldByName("C").SetUint(888)
		assert.Equal(t, rSlice.Interface(), newOff)
	})

	for _, test := range []struct {
		location codec.ElementExtractorLocation
	}{
		{location: codec.ElementExtractorLocationFirst},
		{location: codec.ElementExtractorLocationMiddle},
		{location: codec.ElementExtractorLocationLast},
	} {
		t.Run("Json encoding works", func(t *testing.T) {
			b, err := json.Marshal(test.location)
			require.NoError(t, err)
			var actual codec.ElementExtractorLocation
			require.NoError(t, json.Unmarshal(b, &actual))
			assert.Equal(t, test.location, actual)
		})
	}
}

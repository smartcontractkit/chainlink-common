package codec_test

import (
	"reflect"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropertyExtractor(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		A bool
		B int64
	}

	type nestedTestStruct struct {
		A string
		B testStruct
	}

	extractor := codec.NewPropertyExtractor("A")
	invalidExtractor := codec.NewPropertyExtractor("A.B")
	nestedExtractor := codec.NewPropertyExtractor("B.B")

	t.Run("RetypeToOffChain keeps equal structure to onchain", func(t *testing.T) {
		offChainType, err := extractor.RetypeToOffChain(reflect.TypeOf(nestedTestStruct{}), "")
		require.NoError(t, err)

		require.Equal(t, 2, offChainType.NumField())
		f0 := offChainType.Field(0)
		assert.Equal(t, "A", f0.Name)
		assert.Equal(t, reflect.TypeOf(""), f0.Type)
		f1 := offChainType.Field(1)
		assert.Equal(t, "B", f1.Name)
		assert.Equal(t, reflect.TypeOf(testStruct{}), f1.Type)
	})

	t.Run("RetypeToOffChain returns error for missing field", func(t *testing.T) {
		_, err := invalidExtractor.RetypeToOffChain(reflect.TypeOf(nestedTestStruct{}), "")
		require.NotNil(t, err)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on structs", func(t *testing.T) {
		_, err := extractor.RetypeToOffChain(reflect.TypeOf(nestedTestStruct{}), "")
		require.NoError(t, err)

		onChainValue := nestedTestStruct{
			A: "test",
			B: testStruct{
				A: true,
				B: 42,
			},
		}

		offChainValue, err := extractor.TransformToOffChain(onChainValue, "")
		require.NoError(t, err)
		require.Equal(t, "test", offChainValue)

		reducedOnChain, err := extractor.TransformToOnChain("value", "")
		require.NoError(t, err)

		expected := struct {
			A string
		}{
			A: "value",
		}

		assert.Equal(t, reducedOnChain, expected)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on nested structs", func(t *testing.T) {
		_, err := nestedExtractor.RetypeToOffChain(reflect.TypeOf(nestedTestStruct{}), "")
		require.NoError(t, err)

		onChainValue := nestedTestStruct{
			A: "test",
			B: testStruct{
				A: true,
				B: 42,
			},
		}

		offChainValue, err := nestedExtractor.TransformToOffChain(onChainValue, "")
		require.NoError(t, err)
		require.Equal(t, int64(42), offChainValue)

		reducedOnChain, err := nestedExtractor.TransformToOnChain(int64(3), "")
		require.NoError(t, err)

		expected := struct {
			B int64
		}{
			B: 3,
		}

		assert.Equal(t, reducedOnChain, expected)
	})

	t.Run("TransformToOnChain and TransformToOffChain works on pointers", func(t *testing.T) {
		_, err := nestedExtractor.RetypeToOffChain(reflect.TypeOf(&nestedTestStruct{}), "")
		require.NoError(t, err)

		onChainValue := &nestedTestStruct{
			A: "test",
			B: testStruct{
				A: true,
				B: 42,
			},
		}

		offChainValue, err := nestedExtractor.TransformToOffChain(onChainValue, "")
		require.NoError(t, err)
		require.Equal(t, int64(42), offChainValue)

		reducedOnChain, err := nestedExtractor.TransformToOnChain(int64(3), "")
		require.NoError(t, err)

		expected := struct {
			B int64
		}{
			B: 3,
		}

		assert.Equal(t, reducedOnChain, expected)
	})
}

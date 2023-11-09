package codec_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-relay/pkg/codec"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

func TestFitsInNBitsSigned(t *testing.T) {
	t.Parallel()
	t.Run("Fits", func(t *testing.T) {
		bi := big.NewInt(math.MaxInt16)
		assert.True(t, codec.FitsInNBitsSigned(16, bi))
	})

	t.Run("Too large", func(t *testing.T) {
		bi := big.NewInt(math.MaxInt16 + 1)
		assert.False(t, codec.FitsInNBitsSigned(16, bi))
	})

	t.Run("Too small", func(t *testing.T) {
		bi := big.NewInt(math.MinInt16 - 1)
		assert.False(t, codec.FitsInNBitsSigned(16, bi))
	})
}

func TestMergeValueFields(t *testing.T) {
	t.Parallel()
	t.Run("Merges fields", func(t *testing.T) {
		input := []map[string]any{
			{"Foo": int32(1), "Bar": "Hi"},
			{"Foo": int32(2), "Bar": "How"},
			{"Foo": int32(3), "Bar": "Are"},
			{"Foo": int32(4), "Bar": "You?"},
		}

		output, err := codec.MergeValueFields(input)
		require.NoError(t, err)

		expected := map[string]any{
			"Foo": []int32{1, 2, 3, 4},
			"Bar": []string{"Hi", "How", "Are", "You?"},
		}
		assert.Equal(t, expected, output)
	})

	t.Run("Returns error if keys are not the same", func(t *testing.T) {
		input := []map[string]any{
			{"Foo": int32(1), "Bar": "Hi"},
			{"Zap": 2, "Foo": int32(2), "Bar": "How"},
		}

		_, err := codec.MergeValueFields(input)

		assert.IsType(t, types.InvalidTypeError{}, err)
	})

	t.Run("Returns error if values are not compatible types", func(t *testing.T) {
		input := []map[string]any{
			{"Foo": int32(1), "Bar": "Hi"},
			{"Foo": int32(2), "Bar": int32(3)},
		}

		_, err := codec.MergeValueFields(input)

		assert.IsType(t, types.InvalidTypeError{}, err)
	})
}

func TestSplitValueField(t *testing.T) {
	t.Parallel()
	t.Run("Returns slit field values", func(t *testing.T) {
		input := map[string]any{
			"Foo": []int32{1, 2, 3, 4},
			"Bar": [4]string{"Hi", "How", "Are", "You?"},
		}

		output, err := codec.SplitValueFields(input)
		require.NoError(t, err)

		expected := []map[string]any{
			{"Foo": int32(1), "Bar": "Hi"},
			{"Foo": int32(2), "Bar": "How"},
			{"Foo": int32(3), "Bar": "Are"},
			{"Foo": int32(4), "Bar": "You?"},
		}
		assert.Equal(t, expected, output)
	})

	t.Run("Returns error if lengths do not match", func(t *testing.T) {
		input := map[string]any{
			"Foo": []int32{1, 2, 3},
			"Bar": []string{"Hi", "How", "Are", "You?"},
		}

		_, err := codec.SplitValueFields(input)
		assert.IsType(t, types.InvalidTypeError{}, err)
	})

	t.Run("Returns error if item is not an array or slice", func(t *testing.T) {
		input := map[string]any{
			"Foo": int32(3),
			"Bar": []string{"Hi", "How", "Are", "You?"},
		}

		_, err := codec.SplitValueFields(input)
		assert.IsType(t, types.NotASliceError{}, err)
	})
}

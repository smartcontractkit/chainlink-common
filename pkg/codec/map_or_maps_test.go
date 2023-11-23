package codec_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func TestFlattenToMapOrMaps(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result, err := codec.FlattenToMapOrMaps(nil)
		require.NoError(t, err)
		assertMapOrMaps(t, []map[string]any{}, result)
	})

	t.Run("map[string]any input flattens fields", func(t *testing.T) {
		input := map[string]any{
			"Z": "w",
			"Q": mapOrMapsTestStruct{
				A: 1,
				B: "2",
				C: nestedMapOrMapsTestStruct{
					A: "3",
					X: 4,
				},
			},
		}

		result, err := codec.FlattenToMapOrMaps(input)
		require.NoError(t, err)

		expected := map[string]any{
			"Z": "w",
			"Q": map[string]any{
				"A": 1,
				"B": "2",
				"C": map[string]any{
					"A": "3",
					"X": 4,
				},
			},
		}
		assertMapOrMaps(t, []map[string]any{expected}, result)
	})

	t.Run("map[string]type input", func(t *testing.T) {
		input := map[string]mapOrMapsTestStruct{
			"Q": {
				A: 1,
				B: "2",
				C: nestedMapOrMapsTestStruct{
					A: "3",
					X: 4,
				},
			},
		}

		result, err := codec.FlattenToMapOrMaps(input)
		require.NoError(t, err)

		expected := map[string]any{
			"Q": map[string]any{
				"A": 1,
				"B": "2",
				"C": map[string]any{
					"A": "3",
					"X": 4,
				},
			},
		}
		assertMapOrMaps(t, []map[string]any{expected}, result)
	})

	t.Run("map with non-string key returns errors", func(t *testing.T) {
		input := map[any]any{1: "A"}
		_, err := codec.FlattenToMapOrMaps(input)
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("slice of maps[string]any input", func(t *testing.T) {
		input := []map[string]any{
			{
				"Q": mapOrMapsTestStruct{
					A: 1,
					B: "2",
					C: nestedMapOrMapsTestStruct{
						A: "3",
						X: 4,
					},
				},
			},
			{
				"Z": mapOrMapsTestStruct{
					A: 5,
					B: "6",
					C: nestedMapOrMapsTestStruct{
						A: "7",
						X: 8,
					},
				},
			},
		}

		result, err := codec.FlattenToMapOrMaps(input)
		require.NoError(t, err)

		expected := []map[string]any{
			{
				"Q": map[string]any{
					"A": 1,
					"B": "2",
					"C": map[string]any{
						"A": "3",
						"X": 4,
					},
				},
			},
			{
				"Z": map[string]any{
					"A": 5,
					"B": "6",
					"C": map[string]any{
						"A": "7",
						"X": 8,
					},
				},
			},
		}
		assertMapOrMaps(t, expected, result)
	})

	t.Run("array of maps[string]type input", func(t *testing.T) {
		input := []map[string]mapOrMapsTestStruct{
			{
				"Q": {
					A: 1,
					B: "2",
					C: nestedMapOrMapsTestStruct{
						A: "3",
						X: 4,
					},
				},
			},
			{
				"Z": {
					A: 5,
					B: "6",
					C: nestedMapOrMapsTestStruct{
						A: "7",
						X: 8,
					},
				},
			},
		}

		result, err := codec.FlattenToMapOrMaps(input)
		require.NoError(t, err)

		expected := []map[string]any{
			{
				"Q": map[string]any{
					"A": 1,
					"B": "2",
					"C": map[string]any{
						"A": "3",
						"X": 4,
					},
				},
			},
			{
				"Z": map[string]any{
					"A": 5,
					"B": "6",
					"C": map[string]any{
						"A": "7",
						"X": 8,
					},
				},
			},
		}
		assertMapOrMaps(t, expected, result)
	})

	t.Run("slice of map with non-string key returns errors", func(t *testing.T) {
		input := []map[any]any{{1: "A"}}
		_, err := codec.FlattenToMapOrMaps(input)
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("struct input", func(t *testing.T) {
		input := mapOrMapsTestStruct{
			A: 1,
			B: "2",
			C: nestedMapOrMapsTestStruct{
				A: "3",
				X: 4,
			},
		}

		result, err := codec.FlattenToMapOrMaps(input)
		require.NoError(t, err)

		expected := map[string]any{
			"A": 1,
			"B": "2",
			"C": map[string]any{
				"A": "3",
				"X": 4,
			},
		}
		assertMapOrMaps(t, []map[string]any{expected}, result)
	})

	t.Run("pointer input", func(t *testing.T) {
		input := &mapOrMapsTestStruct{
			A: 1,
			B: "2",
			C: nestedMapOrMapsTestStruct{
				A: "3",
				X: 4,
			},
		}

		result, err := codec.FlattenToMapOrMaps(input)
		require.NoError(t, err)

		expected := map[string]any{
			"A": 1,
			"B": "2",
			"C": map[string]any{
				"A": "3",
				"X": 4,
			},
		}
		assertMapOrMaps(t, []map[string]any{expected}, result)
	})

	t.Run("nested pointer input", func(t *testing.T) {
		input := &mapOrMapsTestStruct{
			A: 1,
			B: "2",
			C: nestedMapOrMapsTestStruct{
				A: "3",
				X: 4,
			},
		}

		result, err := codec.FlattenToMapOrMaps(&input)
		require.NoError(t, err)

		expected := map[string]any{
			"A": 1,
			"B": "2",
			"C": map[string]any{
				"A": "3",
				"X": 4,
			},
		}
		assertMapOrMaps(t, []map[string]any{expected}, result)
	})

	t.Run("slice input", func(t *testing.T) {
		input := []mapOrMapsTestStruct{
			{
				A: 1,
				B: "2",
				C: nestedMapOrMapsTestStruct{
					A: "3",
					X: 4,
				},
			},
			{
				A: 5,
				B: "6",
				C: nestedMapOrMapsTestStruct{
					A: "7",
					X: 8,
				},
			},
		}

		result, err := codec.FlattenToMapOrMaps(input)
		require.NoError(t, err)

		expected := []map[string]any{
			{
				"A": 1,
				"B": "2",
				"C": map[string]any{
					"A": "3",
					"X": 4,
				},
			},
			{
				"A": 5,
				"B": "6",
				"C": map[string]any{
					"A": "7",
					"X": 8,
				},
			},
		}
		assertMapOrMaps(t, expected, result)
	})

	t.Run("array input", func(t *testing.T) {
		input := [2]mapOrMapsTestStruct{
			{
				A: 1,
				B: "2",
				C: nestedMapOrMapsTestStruct{
					A: "3",
					X: 4,
				},
			},
			{
				A: 5,
				B: "6",
				C: nestedMapOrMapsTestStruct{
					A: "7",
					X: 8,
				},
			},
		}

		result, err := codec.FlattenToMapOrMaps(input)
		require.NoError(t, err)

		expected := []map[string]any{
			{
				"A": 1,
				"B": "2",
				"C": map[string]any{
					"A": "3",
					"X": 4,
				},
			},
			{
				"A": 5,
				"B": "6",
				"C": map[string]any{
					"A": "7",
					"X": 8,
				},
			},
		}
		assertMapOrMaps(t, expected, result)
	})

	t.Run("nested slice input", func(t *testing.T) {
		input := [][][]mapOrMapsTestStruct{
			{
				{
					{
						A: 1,
						B: "2",
						C: nestedMapOrMapsTestStruct{
							A: "3",
							X: 4,
						},
					},
					{
						A: 5,
						B: "6",
						C: nestedMapOrMapsTestStruct{
							A: "7",
							X: 8,
						},
					},
				},
			},
			{
				{
					{
						A: 9,
						B: "10",
						C: nestedMapOrMapsTestStruct{
							A: "11",
							X: 12,
						},
					},
					{
						A: 13,
						B: "14",
						C: nestedMapOrMapsTestStruct{
							A: "15",
							X: 16,
						},
					},
				},
			},
		}

		result, err := codec.FlattenToMapOrMaps(input)
		require.NoError(t, err)

		expected := []map[string]any{
			{
				"A": 1,
				"B": "2",
				"C": map[string]any{
					"A": "3",
					"X": 4,
				},
			},
			{
				"A": 5,
				"B": "6",
				"C": map[string]any{
					"A": "7",
					"X": 8,
				},
			},
			{
				"A": 9,
				"B": "10",
				"C": map[string]any{
					"A": "11",
					"X": 12,
				},
			},
			{
				"A": 13,
				"B": "14",
				"C": map[string]any{
					"A": "15",
					"X": 16,
				},
			},
		}
		assertMapOrMaps(t, expected, result)
	})

	t.Run("invalid input returns errors", func(t *testing.T) {
		_, err := codec.FlattenToMapOrMaps(1)
		assert.True(t, errors.Is(err, types.ErrInvalidType))

		_, err = codec.FlattenToMapOrMaps(make(chan int))
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func assertMapOrMaps(t *testing.T, expected []map[string]any, actual *codec.MapOrMaps) {
	seen := make([]map[string]any, 0, len(expected))
	assert.NoError(t, actual.ForEachNestedMap(func(m map[string]any) error {
		seen = append(seen, m)
		return nil
	}))
	assert.Equal(t, expected, seen)
}

type mapOrMapsTestStruct struct {
	A int
	B string
	C nestedMapOrMapsTestStruct
}

type nestedMapOrMapsTestStruct struct {
	A string
	X int
}

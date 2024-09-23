package values

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IntUnwrapTo(t *testing.T) {
	expected := int64(100)
	v := NewInt64(expected)

	var got int
	err := v.UnwrapTo(&got)
	require.NoError(t, err)

	assert.Equal(t, expected, int64(got))

	var gotInt64 int64
	err = v.UnwrapTo(&gotInt64)
	require.NoError(t, err)

	assert.Equal(t, expected, gotInt64)

	var varAny any
	err = v.UnwrapTo(&varAny)
	require.NoError(t, err)
	assert.Equal(t, expected, varAny)

	in := (*Int64)(nil)
	_, err = in.Unwrap()
	assert.ErrorContains(t, err, "cannot unwrap nil")

	var i int64
	err = in.UnwrapTo(&i)
	assert.ErrorContains(t, err, "cannot unwrap nil")
}

func Test_IntUnwrapping(t *testing.T) {
	t.Run("int64 -> int32", func(st *testing.T) {
		expected := int64(100)
		v := NewInt64(expected)
		got := int32(0)
		err := v.UnwrapTo(&got)
		require.NoError(t, err)
		assert.Equal(t, int32(expected), got)
	})

	t.Run("int64 -> int32; overflow", func(st *testing.T) {
		expected := int64(math.MaxInt64)
		v := NewInt64(expected)
		got := int32(0)
		err := v.UnwrapTo(&got)
		assert.NotNil(t, err)
	})

	t.Run("int64 -> int32; underflow", func(st *testing.T) {
		expected := int64(math.MinInt64)
		v := NewInt64(expected)
		got := int32(0)
		err := v.UnwrapTo(&got)
		assert.NotNil(t, err)
	})

	t.Run("int64 -> uint32", func(st *testing.T) {
		expected := int64(100)
		v := NewInt64(expected)
		got := uint32(0)
		err := v.UnwrapTo(&got)
		require.NoError(t, err)
		assert.Equal(t, uint32(expected), got)
	})

	t.Run("int64 -> uint32; overflow", func(st *testing.T) {
		expected := int64(math.MaxInt64)
		v := NewInt64(expected)
		got := uint32(0)
		err := v.UnwrapTo(&got)
		assert.NotNil(t, err)
	})

	t.Run("int64 -> uint32; underflow", func(st *testing.T) {
		expected := int64(math.MinInt64)
		v := NewInt64(expected)
		got := uint32(0)
		err := v.UnwrapTo(&got)
		assert.NotNil(t, err)
	})

	t.Run("int64 -> uint64; underflow", func(st *testing.T) {
		expected := int64(math.MinInt64)
		v := NewInt64(expected)
		got := uint64(0)
		err := v.UnwrapTo(&got)
		assert.NotNil(t, err)
	})
}

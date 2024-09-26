package values

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Float64UnwrapTo(t *testing.T) {
	expected := 1.1
	v := NewFloat64(expected)

	var got float64
	err := v.UnwrapTo(&got)
	require.NoError(t, err)

	assert.Equal(t, expected, got)

	gotn := (*float64)(nil)
	err = v.UnwrapTo(gotn)
	assert.ErrorContains(t, err, "cannot unwrap to nil pointer")

	var varAny any
	err = v.UnwrapTo(&varAny)
	require.NoError(t, err)
	assert.Equal(t, expected, varAny)

	fn := (*Float64)(nil)
	_, err = fn.Unwrap()
	assert.ErrorContains(t, err, "cannot unwrap nil")

	var f float64
	err = fn.UnwrapTo(&f)
	assert.ErrorContains(t, err, "cannot unwrap nil")

	// handle alias
	type myFloat64 float64
	var mf myFloat64
	err = v.UnwrapTo(&mf)
	require.NoError(t, err)
	assert.Equal(t, myFloat64(expected), mf)
}

// Test_Float64 tests that Float64 values can converted to and from protobuf representations.
func Test_Float64(t *testing.T) {
	testCases := []struct {
		name string
		f    float64
	}{
		{
			name: "positive",
			f:    1.1,
		},
		{
			name: "0",
			f:    0,
		},
		{
			name: "negative",
			f:    -1.1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			v := NewFloat64(tc.f)

			vp := Proto(v)
			got, err := FromProto(vp)
			assert.NoError(t, err)
			assert.Equal(t, tc.f, got.(*Float64).Underlying)
		})
	}
}

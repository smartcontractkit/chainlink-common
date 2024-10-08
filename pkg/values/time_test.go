package values

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_TimeUnwrapTo(t *testing.T) {
	expected, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	assert.NoError(t, err)

	// Unwraps to a time.Time pointer
	v := NewTime(expected)
	got := new(time.Time)
	err = v.UnwrapTo(got)
	assert.NoError(t, err)
	assert.Equal(t, expected, *got)

	// Fails to unwrap to nil time.Time pointer
	gotTime := (*time.Time)(nil)
	err = v.UnwrapTo(gotTime)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "cannot unwrap to nil pointer")

	// Unwraps to an any pointer
	var varAny any
	err = v.UnwrapTo(&varAny)
	assert.NoError(t, err)
	assert.Equal(t, expected, varAny)

	// Fails to unwrap to a string pointer
	var varStr string
	err = v.UnwrapTo(&varStr)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "cannot unwrap to value of type: *string")

	// Fails to unwrap nil value of Time
	nilVal := (*Time)(nil)
	_, err = nilVal.Unwrap()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "could not unwrap nil")

	// Unwraps zero value of Time
	zeroTime := &Time{}
	unwrapped, err := zeroTime.Unwrap()
	assert.NoError(t, err)
	assert.Equal(t, time.Time{}, unwrapped)

	// Unwraps an alias
	type aliasTime time.Time
	alias := aliasTime(time.Time{})
	err = v.UnwrapTo(&alias)
	assert.NoError(t, err)
	assert.Equal(t, expected, time.Time(alias))
}

// Test_Time tests that Time values can be converted to and from protobuf representations.
func Test_Time(t *testing.T) {
	testCases := []struct {
		name string
		t    time.Time
	}{
		{
			name: "zero",
			t:    time.Time{},
		},
		{
			name: "some time",
			t: func() time.Time {
				someTime, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
				assert.NoError(t, err)
				return someTime
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := NewTime(tc.t)

			vp := Proto(v)
			got, err := FromProto(vp)
			assert.NoError(t, err)
			assert.Equal(t, tc.t, got.(*Time).Underlying)
		})
	}
}

package sqlutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewInterval(t *testing.T) {
	t.Parallel()

	duration := 33 * time.Second
	interval := NewInterval(duration)

	require.Equal(t, duration, interval.Duration())
}

func TestInterval_IsZero(t *testing.T) {
	t.Parallel()

	i := NewInterval(0)
	require.NotNil(t, i)
	require.True(t, i.IsZero())

	i = NewInterval(1)
	require.NotNil(t, i)
	require.False(t, i.IsZero())
}

func TestInterval_Scan_Value(t *testing.T) {
	t.Parallel()

	i := NewInterval(100)
	require.NotNil(t, i)

	val, err := i.Value()
	require.NoError(t, err)

	iNew := NewInterval(0)
	err = iNew.Scan(val)
	require.NoError(t, err)

	require.Equal(t, i, iNew)
}

func TestInterval_MarshalText_UnmarshalText(t *testing.T) {
	t.Parallel()

	i := NewInterval(100)
	require.NotNil(t, i)

	txt, err := i.MarshalText()
	require.NoError(t, err)

	iNew := NewInterval(0)
	err = iNew.UnmarshalText(txt)
	require.NoError(t, err)

	require.Equal(t, i, iNew)
}

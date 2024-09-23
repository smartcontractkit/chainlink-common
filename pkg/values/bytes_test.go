package values

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BytesUnwrapTo(t *testing.T) {
	hs := []byte("hello")
	tr := NewBytes(hs)

	got := []byte{}
	err := tr.UnwrapTo(&got)
	require.NoError(t, err)

	assert.Equal(t, hs, got)

	var b bool
	err = tr.UnwrapTo(&b)
	require.Error(t, err)

	str := ""
	err = tr.UnwrapTo(&str)
	require.NoError(t, err)
	assert.Equal(t, []byte(str), tr.Underlying)

	gotB := (*[]byte)(nil)
	err = tr.UnwrapTo(gotB)
	assert.ErrorContains(t, err, "cannot unwrap to nil pointer")

	var varAny any
	err = tr.UnwrapTo(&varAny)
	require.NoError(t, err)
	assert.Equal(t, hs, varAny)

	bn := (*Bytes)(nil)
	_, err = bn.Unwrap()
	assert.ErrorContains(t, err, "cannot unwrap nil")

	var bp []byte
	err = bn.UnwrapTo(bp)
	assert.ErrorContains(t, err, "cannot unwrap nil")

	bn = &Bytes{}
	err = bn.UnwrapTo(&bp)
	require.NoError(t, err)
	assert.Nil(t, bp)
}

type alias uint8

func Test_BytesUnwrapToAlias(t *testing.T) {
	underlying := []byte("hello")
	bn := &Bytes{Underlying: underlying}
	bp := []alias{}
	err := bn.UnwrapTo(&bp)
	require.NoError(t, err)

	got := []byte{}
	for _, b := range bp {
		got = append(got, byte(b))
	}
	assert.Equal(t, underlying, got)

	var oracleIDs [5]alias
	underlying = []byte("hello")
	bn = &Bytes{Underlying: underlying}
	err = bn.UnwrapTo(&oracleIDs)
	require.NoError(t, err)
	got = []byte{}
	for _, b := range oracleIDs {
		got = append(got, byte(b))
	}
	assert.Equal(t, underlying, got)
}

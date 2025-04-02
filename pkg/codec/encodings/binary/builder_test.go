package binary

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForInputBufferCorruption(t *testing.T) {
	builder := LittleEndian()

	codec, err := builder.BigInt(16, true)
	require.NoError(t, err)

	myBigInt := big.NewInt(3)

	buf := make([]byte, 0, 16)
	buf, err = codec.Encode(myBigInt, buf)
	require.NoError(t, err)

	decoded, _, err := codec.Decode(buf)
	require.NoError(t, err)
	require.Equal(t, myBigInt, decoded)

	// decoding from the same buffer again should give same answer
	decoded, _, err = codec.Decode(buf)
	require.NoError(t, err)
	assert.Equal(t, myBigInt, decoded)
}

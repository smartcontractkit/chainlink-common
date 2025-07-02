package ccipocr3

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCurseInfo_NonCursedSourceChains(t *testing.T) {
	chains := []ChainSelector{1, 2, 3, 4}
	ci := CurseInfo{
		CursedSourceChains: map[ChainSelector]bool{
			2: true,
			4: true,
		},
		CursedDestination: false,
		GlobalCurse:       false,
	}
	result := ci.NonCursedSourceChains(chains)
	require.Equal(t, []ChainSelector{1, 3}, result)

	ci.GlobalCurse = true
	result = ci.NonCursedSourceChains(chains)
	require.Nil(t, result)
}

func TestTimeStampedBigFromUnix(t *testing.T) {
	unixValue := big.NewInt(12345)
	unixTimestamp := uint32(1700000000)
	input := TimestampedUnixBig{
		Value:     unixValue,
		Timestamp: unixTimestamp,
	}
	result := TimeStampedBigFromUnix(input)
	require.Equal(t, NewBigInt(unixValue), result.Value)
	require.Equal(t, time.Unix(int64(unixTimestamp), 0), result.Timestamp)
}

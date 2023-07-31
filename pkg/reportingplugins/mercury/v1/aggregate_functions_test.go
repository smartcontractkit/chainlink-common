package mercury_v1

import (
	"math/big"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/pkg/reportingplugins/mercury"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewValidParsedAttributedObservations() []ParsedAttributedObservation {
	return []ParsedAttributedObservation{
		parsedAttributedObservation{
			Timestamp: 1689648456,

			BenchmarkPrice: big.NewInt(123),
			Bid:            big.NewInt(120),
			Ask:            big.NewInt(130),
			PricesValid:    true,

			MaxFinalizedTimestamp:      1679648456,
			MaxFinalizedTimestmapValid: true,

			LinkFee:        big.NewInt(1),
			LinkFeeValid:   true,
			NativeFee:      big.NewInt(1),
			NativeFeeValid: true,
		},
		parsedAttributedObservation{
			Timestamp: 1689648456,

			BenchmarkPrice: big.NewInt(456),
			Bid:            big.NewInt(450),
			Ask:            big.NewInt(460),
			PricesValid:    true,

			MaxFinalizedTimestamp:      1679648456,
			MaxFinalizedTimestmapValid: true,

			LinkFee:        big.NewInt(2),
			LinkFeeValid:   true,
			NativeFee:      big.NewInt(2),
			NativeFeeValid: true,
		},
		parsedAttributedObservation{
			Timestamp: 1689648789,

			BenchmarkPrice: big.NewInt(789),
			Bid:            big.NewInt(780),
			Ask:            big.NewInt(800),
			PricesValid:    true,

			MaxFinalizedTimestamp:      1679648456,
			MaxFinalizedTimestmapValid: true,

			LinkFee:        big.NewInt(3),
			LinkFeeValid:   true,
			NativeFee:      big.NewInt(3),
			NativeFeeValid: true,
		},
		parsedAttributedObservation{
			Timestamp: 1689648789,

			BenchmarkPrice: big.NewInt(456),
			Bid:            big.NewInt(450),
			Ask:            big.NewInt(460),
			PricesValid:    true,

			MaxFinalizedTimestamp:      1679513477,
			MaxFinalizedTimestmapValid: true,

			LinkFee:        big.NewInt(4),
			LinkFeeValid:   true,
			NativeFee:      big.NewInt(4),
			NativeFeeValid: true,
		},
	}
}

func NewInvalidParsedAttributedObservations() []ParsedAttributedObservation {
	return []ParsedAttributedObservation{
		parsedAttributedObservation{
			Timestamp: 1,

			BenchmarkPrice: big.NewInt(123),
			Bid:            big.NewInt(120),
			Ask:            big.NewInt(130),
			PricesValid:    false,

			MaxFinalizedTimestamp:      1679648456,
			MaxFinalizedTimestmapValid: false,

			LinkFee:        big.NewInt(1),
			LinkFeeValid:   false,
			NativeFee:      big.NewInt(1),
			NativeFeeValid: false,
		},
		parsedAttributedObservation{
			Timestamp: 2,

			BenchmarkPrice: big.NewInt(456),
			Bid:            big.NewInt(450),
			Ask:            big.NewInt(460),
			PricesValid:    false,

			MaxFinalizedTimestamp:      1679648456,
			MaxFinalizedTimestmapValid: false,

			LinkFee:        big.NewInt(2),
			LinkFeeValid:   false,
			NativeFee:      big.NewInt(2),
			NativeFeeValid: false,
		},
		parsedAttributedObservation{
			Timestamp: 2,

			BenchmarkPrice: big.NewInt(789),
			Bid:            big.NewInt(780),
			Ask:            big.NewInt(800),
			PricesValid:    false,

			MaxFinalizedTimestamp:      1679648456,
			MaxFinalizedTimestmapValid: false,

			LinkFee:        big.NewInt(3),
			LinkFeeValid:   false,
			NativeFee:      big.NewInt(3),
			NativeFeeValid: false,
		},
		parsedAttributedObservation{
			Timestamp: 3,

			BenchmarkPrice: big.NewInt(456),
			Bid:            big.NewInt(450),
			Ask:            big.NewInt(460),
			PricesValid:    true,

			MaxFinalizedTimestamp:      1679513477,
			MaxFinalizedTimestmapValid: true,

			LinkFee:        big.NewInt(4),
			LinkFeeValid:   true,
			NativeFee:      big.NewInt(4),
			NativeFeeValid: true,
		},
	}
}

func Test_AggregateFunctions(t *testing.T) {
	f := 1
	validPaos := NewValidParsedAttributedObservations()
	invalidPaos := NewInvalidParsedAttributedObservations()

	t.Run("GetConsensusTimestamp", func(t *testing.T) {
		validMPaos := Convert(validPaos)
		ts := mercury.GetConsensusTimestamp(validMPaos)

		assert.Equal(t, 1689648789, int(ts))
	})

	t.Run("GetConsensusBenchmarkPrice", func(t *testing.T) {
		t.Run("gets consensus price when prices are valid", func(t *testing.T) {
			validMPaos := Convert(validPaos)
			bp, err := mercury.GetConsensusBenchmarkPrice(validMPaos, f)
			require.NoError(t, err)
			assert.Equal(t, "456", bp.String())
		})

		t.Run("fails when fewer than f+1 prices are valid", func(t *testing.T) {
			invalidMPaos := Convert(invalidPaos)
			_, err := mercury.GetConsensusBenchmarkPrice(invalidMPaos, f)
			assert.EqualError(t, err, "fewer than f+1 observations have a valid price (got: 1/4)")
		})
	})

	t.Run("GetConsensusBid", func(t *testing.T) {
		t.Run("gets consensus bid when prices are valid", func(t *testing.T) {
			validMPaos := Convert(validPaos)
			bid, err := mercury.GetConsensusBid(validMPaos, f)
			require.NoError(t, err)
			assert.Equal(t, "450", bid.String())
		})

		t.Run("fails when fewer than f+1 prices are valid", func(t *testing.T) {
			invalidMPaos := Convert(invalidPaos)
			_, err := mercury.GetConsensusBid(invalidMPaos, f)
			assert.EqualError(t, err, "fewer than f+1 observations have a valid price (got: 1/4)")
		})
	})

	t.Run("GetConsensusAsk", func(t *testing.T) {
		t.Run("gets consensus ask when prices are valid", func(t *testing.T) {
			validMPaos := Convert(validPaos)
			bid, err := mercury.GetConsensusAsk(validMPaos, f)
			require.NoError(t, err)
			assert.Equal(t, "460", bid.String())
		})

		t.Run("fails when fewer than f+1 prices are valid", func(t *testing.T) {
			invalidMPaos := Convert(invalidPaos)
			_, err := mercury.GetConsensusAsk(invalidMPaos, f)
			assert.EqualError(t, err, "fewer than f+1 observations have a valid price (got: 1/4)")
		})
	})

	t.Run("GetConsensusMaxFinalizedTimestamp", func(t *testing.T) {
		t.Run("gets consensus on maxFinalizedTimestamp when valid", func(t *testing.T) {
			ts, err := GetConsensusMaxFinalizedTimestamp(validPaos, f)
			require.NoError(t, err)
			assert.Equal(t, uint32(1679648456), ts)
		})

		t.Run("fails when fewer than f+1 maxFinalizedTimestamps are valid", func(t *testing.T) {
			_, err := GetConsensusMaxFinalizedTimestamp(invalidPaos, f)
			assert.EqualError(t, err, "fewer than f+1 observations have a valid maxFinalizedTimestamp (got: 1/4)")
		})
	})

	t.Run("GetConsensusLinkFee", func(t *testing.T) {
		t.Run("gets consensus on linkFee when valid", func(t *testing.T) {
			validMPaos := Convert(validPaos)
			linkFee, err := mercury.GetConsensusLinkFee(validMPaos, f)
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(3), linkFee)
		})

		t.Run("fails when fewer than f+1 linkFees are valid", func(t *testing.T) {
			invalidMPaos := Convert(invalidPaos)
			_, err := mercury.GetConsensusLinkFee(invalidMPaos, f)
			assert.EqualError(t, err, "fewer than f+1 observations have a valid linkFee (got: 1/4)")
		})
	})

	t.Run("GetConsensusNativeFee", func(t *testing.T) {
		t.Run("gets consensus on nativeFee when valid", func(t *testing.T) {
			validMPaos := Convert(validPaos)
			nativeFee, err := mercury.GetConsensusNativeFee(validMPaos, f)
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(3), nativeFee)
		})

		t.Run("fails when fewer than f+1 nativeFees are valid", func(t *testing.T) {
			invalidMPaos := Convert(invalidPaos)
			_, err := mercury.GetConsensusNativeFee(invalidMPaos, f)
			assert.EqualError(t, err, "fewer than f+1 observations have a valid nativeFee (got: 1/4)")
		})
	})
}

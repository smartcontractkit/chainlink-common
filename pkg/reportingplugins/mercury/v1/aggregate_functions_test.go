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
			PricesValid:    true,

			MaxFinalizedTimestamp: 1679648456,
		},
		parsedAttributedObservation{
			Timestamp: 1689648456,

			BenchmarkPrice: big.NewInt(456),
			PricesValid:    true,

			MaxFinalizedTimestamp: 1679648456,
		},
		parsedAttributedObservation{
			Timestamp: 1689648789,

			BenchmarkPrice: big.NewInt(789),
			PricesValid:    true,

			MaxFinalizedTimestamp: 1679648456,
		},
		parsedAttributedObservation{
			Timestamp: 1689648789,

			BenchmarkPrice:        big.NewInt(456),
			PricesValid:           true,
			MaxFinalizedTimestamp: 1679513477,
		},
	}
}

func NewInvalidParsedAttributedObservations() []ParsedAttributedObservation {
	return []ParsedAttributedObservation{
		parsedAttributedObservation{
			Timestamp: 1,

			BenchmarkPrice: big.NewInt(123),
			PricesValid:    false,

			MaxFinalizedTimestamp: 1679648456,
		},
		parsedAttributedObservation{
			Timestamp: 2,

			BenchmarkPrice: big.NewInt(456),
			PricesValid:    false,

			MaxFinalizedTimestamp: 1679648456,
		},
		parsedAttributedObservation{
			Timestamp: 2,

			BenchmarkPrice: big.NewInt(789),
			PricesValid:    false,

			MaxFinalizedTimestamp: 1679648456,
		},
		parsedAttributedObservation{
			Timestamp: 3,

			BenchmarkPrice: big.NewInt(456),
			PricesValid:    true,

			MaxFinalizedTimestamp: 1679513477,
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
			assert.EqualError(t, err, "fewer than f+1 observations have a valid price")
		})
	})

	t.Run("GetConsensusBid", func(t *testing.T) {
		// t.Run("gets consensus bid when prices are valid", func(t *testing.T) {
		// 	validMPaos := Convert(validPaos)
		// 	bid, err := mercury.GetConsensusBid(validMPaos, f)
		// 	require.NoError(t, err)
		// 	assert.Equal(t, "345", bid.String())
		// })

		// t.Run("fails when fewer than f+1 prices are valid", func(t *testing.T) {
		// 	invalidMPaos := Convert(invalidPaos)
		// 	_, err := mercury.GetConsensusBid(invalidMPaos, f)
		// 	assert.EqualError(t, err, "fewer than f+1 observations have a valid price")
		// })
	})

	t.Run("GetConsensusMaxFinalizedTimestamp", func(t *testing.T) {
		ts := GetConsensusMaxFinalizedTimestamp(validPaos)

		assert.Equal(t, 1679648456, int(ts))
	})
}

package mercury_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewValidParsedAttributedObservations() []IParsedAttributedObservation {
	return []IParsedAttributedObservation{
		ParsedAttributedObservation{
			Timestamp:             1689648456,
			MaxFinalizedTimestamp: 1679648456,
		},
		ParsedAttributedObservation{
			Timestamp:             1689648456,
			MaxFinalizedTimestamp: 1679648456,
		},
		ParsedAttributedObservation{
			Timestamp:             1689648789,
			MaxFinalizedTimestamp: 1679648456,
		},
		ParsedAttributedObservation{
			Timestamp:             1689648789,
			MaxFinalizedTimestamp: 1679513477,
		},
	}
}

func NewInvalidParsedAttributedObservations() []IParsedAttributedObservation {
	return []IParsedAttributedObservation{
		ParsedAttributedObservation{
			Timestamp:             1,
			MaxFinalizedTimestamp: 1679648456,
		},
		ParsedAttributedObservation{
			Timestamp:             2,
			MaxFinalizedTimestamp: 1679648456,
		},
		ParsedAttributedObservation{
			Timestamp:             2,
			MaxFinalizedTimestamp: 1679648456,
		},
		ParsedAttributedObservation{
			Timestamp:             3,
			MaxFinalizedTimestamp: 1679513477,
		},
	}
}

func Test_AggregateFunctions(t *testing.T) {
	// f := 1
	validPaos := NewValidParsedAttributedObservations()

	t.Run("GetConsensusMaxFinalizedTimestamp", func(t *testing.T) {
		ts := GetConsensusMaxFinalizedTimestamp(validPaos)

		assert.Equal(t, 1679648456, int(ts))
	})
}

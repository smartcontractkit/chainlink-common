package ccipocr3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasNonEmptyDAGasParams(t *testing.T) {
	cases := []struct {
		name   string
		config FeeQuoterDestChainConfig
		want   bool
	}{
		{
			name:   "all DA gas params non-zero",
			config: FeeQuoterDestChainConfig{DestDataAvailabilityOverheadGas: 1, DestGasPerDataAvailabilityByte: 1, DestDataAvailabilityMultiplierBps: 1},
			want:   true,
		},
		{
			name:   "one DA gas param zero",
			config: FeeQuoterDestChainConfig{DestDataAvailabilityOverheadGas: 0, DestGasPerDataAvailabilityByte: 1, DestDataAvailabilityMultiplierBps: 1},
			want:   false,
		},
		{
			name:   "all DA gas params zero",
			config: FeeQuoterDestChainConfig{DestDataAvailabilityOverheadGas: 0, DestGasPerDataAvailabilityByte: 0, DestDataAvailabilityMultiplierBps: 0},
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.HasNonEmptyDAGasParams()
			assert.Equal(t, tc.want, got)
		})
	}
}

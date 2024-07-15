package llo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ChannelDefinitions_Serialization(t *testing.T) {
	inputJSON := `
{
  "0": {
    "reportFormat": "json",
    "streamIDs": [
      0,
      1
    ],
    "aggregators": [
      "median",
      "mode"
    ],
    "opts": null
  },
  "1": {
    "reportFormat": "evm_premium_legacy",
    "streamIDs": [
      0,
      1,
      2
    ],
    "aggregators": [
      "median",
      "median",
      "quote"
    ],
    "opts": {
      "expirationWindow": 86400,
      "multiplier": "1000000000000000000",
      "feedId": "0x0003aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "baseUSDFee": "0.1"
    }
  }
}`
	var channelDefinitions ChannelDefinitions
	err := json.Unmarshal([]byte(inputJSON), &channelDefinitions)
	require.NoError(t, err)

	marshaledJSON, err := json.Marshal(channelDefinitions)

	assert.JSONEq(t, inputJSON, string(marshaledJSON))

	assert.Equal(t, `{"0":{"reportFormat":"json","streamIDs":[0,1],"aggregators":["median","mode"],"opts":null},"1":{"reportFormat":"evm_premium_legacy","streamIDs":[0,1,2],"aggregators":["median","median","quote"],"opts":{"baseUSDFee":"0.1","expirationWindow":86400,"feedId":"0x0003aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","multiplier":"1000000000000000000"}}}`, string(marshaledJSON))
}

func Test_ChannelDefinition_Equals(t *testing.T) {
	t.Run("different ReportFormat", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			StreamIDs:    []StreamID{0, 1},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorMode},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatEVMPremiumLegacy,
			StreamIDs:    []StreamID{0, 1},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorMode},
			Opts:         nil,
		}
		assert.False(t, a.Equals(b))
	})
	t.Run("different StreamIDs", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			StreamIDs:    []StreamID{0, 1},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorMode},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			StreamIDs:    []StreamID{0, 2},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorMode},
			Opts:         nil,
		}
		assert.False(t, a.Equals(b))
	})
	t.Run("different Aggregators", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			StreamIDs:    []StreamID{0, 1},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorMode},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			StreamIDs:    []StreamID{0, 1},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorQuote},
			Opts:         nil,
		}
		assert.False(t, a.Equals(b))
	})
	t.Run("different Opts", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			StreamIDs:    []StreamID{0, 1},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorMode},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			StreamIDs:    []StreamID{0, 1},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorMode},
			Opts:         []byte{0x01},
		}
		assert.False(t, a.Equals(b))
	})
	t.Run("equal", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			StreamIDs:    []StreamID{0, 1},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorMode},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			StreamIDs:    []StreamID{0, 1},
			Aggregators:  []Aggregator{AggregatorMedian, AggregatorMode},
			Opts:         nil,
		}
		assert.True(t, a.Equals(b))
	})
}

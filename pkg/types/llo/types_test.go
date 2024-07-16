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
    "streams": [
	  {"streamID": 1, "aggregator": "median"},
	  {"streamID": 2, "aggregator": "mode"}
    ],
    "opts": null
  },
  "1": {
    "reportFormat": "evm_premium_legacy",
    "streams": [
	  {"streamID": 1, "aggregator": "median"},
	  {"streamID": 2, "aggregator": "median"},
	  {"streamID": 3, "aggregator": "quote"}
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

	assert.Equal(t, `{"0":{"reportFormat":"json","streams":[{"streamID":1,"aggregator":"median"},{"streamID":2,"aggregator":"mode"}],"opts":null},"1":{"reportFormat":"evm_premium_legacy","streams":[{"streamID":1,"aggregator":"median"},{"streamID":2,"aggregator":"median"},{"streamID":3,"aggregator":"quote"}],"opts":{"baseUSDFee":"0.1","expirationWindow":86400,"feedId":"0x0003aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","multiplier":"1000000000000000000"}}}`, string(marshaledJSON))
}

func Test_ChannelDefinition_Equals(t *testing.T) {
	t.Run("different ReportFormat", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatEVMPremiumLegacy,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
		}
		assert.False(t, a.Equals(b))
	})
	t.Run("different StreamIDs", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {2, AggregatorMode}},
			Opts:         nil,
		}
		assert.False(t, a.Equals(b))
	})
	t.Run("different Aggregators", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorQuote}},
			Opts:         nil,
		}
		assert.False(t, a.Equals(b))
	})
	t.Run("different Opts", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         []byte{0x01},
		}
		assert.False(t, a.Equals(b))
	})
	t.Run("equal", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
		}
		assert.True(t, a.Equals(b))
	})
}

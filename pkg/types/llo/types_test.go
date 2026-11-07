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
	  {"streamId": 1, "aggregator": "median"},
	  {"streamId": 2, "aggregator": "mode"}
    ],
    "opts": null,
    "tombstone": false,
	"source": 1,
	"disableNilStreamValues": true
  },
  "1": {
    "reportFormat": "evm_premium_legacy",
    "streams": [
	  {"streamId": 1, "aggregator": "median"},
	  {"streamId": 2, "aggregator": "median"},
	  {"streamId": 3, "aggregator": "quote"}
    ],
    "opts": {
      "expirationWindow": 86400,
      "multiplier": "1000000000000000000",
      "feedId": "0x0003aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "baseUSDFee": "0.1"
    },
    "tombstone": false,
	"source": 2,
	"disableNilStreamValues": false
  }
}`
	var channelDefinitions ChannelDefinitions
	err := json.Unmarshal([]byte(inputJSON), &channelDefinitions)
	require.NoError(t, err)

	marshaledJSON, err := json.Marshal(channelDefinitions)
	require.NoError(t, err)

	assert.JSONEq(t, inputJSON, string(marshaledJSON))
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
	t.Run("different Tombstone", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
			Tombstone:    true,
		}
		assert.False(t, a.Equals(b))
	})

	t.Run("different Source", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
			Source:       1,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatJSON,
			Streams:      []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:         nil,
			Source:       2,
		}
		assert.False(t, a.Equals(b))
	})

	t.Run("different DisableNilStreamValues", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat:           ReportFormatJSON,
			Streams:                []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:                   nil,
			DisableNilStreamValues: true,
		}
		b := ChannelDefinition{
			ReportFormat:           ReportFormatJSON,
			Streams:                []Stream{{0, AggregatorMedian}, {1, AggregatorMode}},
			Opts:                   nil,
			DisableNilStreamValues: false,
		}
		assert.False(t, a.Equals(b))
	})

	t.Run("different SigningScheme", func(t *testing.T) {
		a := ChannelDefinition{
			ReportFormat:  ReportFormatEVMABIEncodeUnpacked,
			Streams:       []Stream{{0, AggregatorMedian}},
			SigningScheme: SigningSchemeMerkle,
		}
		b := ChannelDefinition{
			ReportFormat: ReportFormatEVMABIEncodeUnpacked,
			Streams:      []Stream{{0, AggregatorMedian}},
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

func Test_ChannelDefinitions_Scan(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var c ChannelDefinitions
		err := c.Scan(([]byte)(nil))
		require.NoError(t, err)
		assert.Empty(t, c)
	})
	t.Run("empty", func(t *testing.T) {
		var c ChannelDefinitions
		err := c.Scan([]byte{})
		require.NoError(t, err)
		assert.Empty(t, c)
	})
	t.Run("invalid JSON", func(t *testing.T) {
		var c ChannelDefinitions
		err := c.Scan([]byte(`{`))
		require.Error(t, err)
		assert.Empty(t, c)
	})
	t.Run("valid JSON", func(t *testing.T) {
		var c ChannelDefinitions
		err := c.Scan([]byte(`{"0":{"reportFormat":"json","streams":[{"streamId":1,"aggregator":"median"},{"streamId":2,"aggregator":"mode"}],"opts":null},"1":{"reportFormat":"evm_premium_legacy","streams":[{"streamId":1,"aggregator":"median"},{"streamId":2,"aggregator":"median"},{"streamId":3,"aggregator":"quote"}],"opts":{"baseUSDFee":"0.1","expirationWindow":86400,"feedId":"0x0003aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","multiplier":"1000000000000000000"}}}`))
		require.NoError(t, err)
		assert.Len(t, c, 2)
	})
}

func Test_ChannelDefinitions_Value(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		c := ChannelDefinitions{}
		v, err := c.Value()
		require.NoError(t, err)
		assert.Equal(t, []byte(`{}`), v)
	})
	t.Run("empty", func(t *testing.T) {
		c := ChannelDefinitions{}
		v, err := c.Value()
		require.NoError(t, err)
		assert.Equal(t, []byte(`{}`), v)
	})
	t.Run("valid JSON", func(t *testing.T) {
		c := ChannelDefinitions{
			0: {
				ReportFormat:           ReportFormatJSON,
				Streams:                []Stream{{1, AggregatorMedian}, {2, AggregatorMode}},
				Opts:                   nil,
				Source:                 1,
				DisableNilStreamValues: true,
			},
			1: {
				ReportFormat:           ReportFormatEVMPremiumLegacy,
				Streams:                []Stream{{1, AggregatorMedian}, {2, AggregatorMedian}, {3, AggregatorQuote}},
				Opts:                   []byte(`{"baseUSDFee":"0.1","expirationWindow":86400,"feedId":"0x0003aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","multiplier":"1000000000000000000"}`),
				Source:                 2,
				DisableNilStreamValues: false,
			},
		}
		v, err := c.Value()
		require.NoError(t, err)
		expectedJSON := `{
  "0": {
    "reportFormat": "json",
    "streams": [
      {"streamId": 1, "aggregator": "median"},
      {"streamId": 2, "aggregator": "mode"}
    ],
    "opts": null,
    "tombstone": false,
	"source": 1,
	"disableNilStreamValues": true
  },
  "1": {
    "reportFormat": "evm_premium_legacy",
    "streams": [
      {"streamId": 1, "aggregator": "median"},
      {"streamId": 2, "aggregator": "median"},
      {"streamId": 3, "aggregator": "quote"}
    ],
    "opts": {
      "baseUSDFee": "0.1",
      "expirationWindow": 86400,
      "feedId": "0x0003aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "multiplier": "1000000000000000000"
    },
    "tombstone": false,
	"source": 2,
	"disableNilStreamValues": false
  }
}`
		assert.JSONEq(t, expectedJSON, string(v.([]byte)))
	})
}

func Test_ChannelDefinition_SigningScheme_JSON(t *testing.T) {
	t.Run("empty scheme is omitted", func(t *testing.T) {
		b, err := json.Marshal(ChannelDefinition{ReportFormat: ReportFormatJSON})
		require.NoError(t, err)
		assert.NotContains(t, string(b), "signingScheme")
	})
	t.Run("merkle scheme round trips", func(t *testing.T) {
		var cd ChannelDefinition
		require.NoError(t, json.Unmarshal([]byte(`{"reportFormat":"evm_abi_encode_unpacked","streams":[{"streamId":1,"aggregator":"median"}],"opts":null,"signingScheme":"merkle"}`), &cd))
		assert.Equal(t, SigningSchemeMerkle, cd.SigningScheme)

		b, err := json.Marshal(cd)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"signingScheme":"merkle"`)
	})
}

func Test_ReportFormat_String(t *testing.T) {
	for _, rf := range ReportFormats {
		parsed, err := ReportFormatFromString(rf.String())
		require.NoError(t, err)
		assert.Equal(t, rf, parsed)
	}
	assert.Equal(t, "evm_merkle", ReportFormatEVMMerkle.String())
}

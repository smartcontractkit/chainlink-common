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
    "ReportFormat": "json",
    "StreamIDs": [
      0,
      1
    ],
    "Aggregators": [
      "median",
      "mode"
    ],
    "Opts": null
  },
  "1": {
    "ReportFormat": "evm_premium_legacy",
    "StreamIDs": [
      0,
      1,
      2
    ],
    "Aggregators": [
      "median",
      "median",
      "quote"
    ],
    "Opts": {
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

	assert.Equal(t, `{"0":{"ReportFormat":"json","StreamIDs":[0,1],"Aggregators":["median","mode"],"Opts":null},"1":{"ReportFormat":"evm_premium_legacy","StreamIDs":[0,1,2],"Aggregators":["median","median","quote"],"Opts":{"baseUSDFee":"0.1","expirationWindow":86400,"feedId":"0x0003aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","multiplier":"1000000000000000000"}}}`, string(marshaledJSON))
}

package evm_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/types/evm"
)

const (
	jsonCompact = `{"contracts":{"MinimalConsumer":{"contractABI":"[{\"inputs\":[],\"name\":\"getFactor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]","configs":{"getFactor":{"chainSpecificName":"getFactor"}}}}}`
	jsonPretty  = `{
  "contracts": {
    "MinimalConsumer": {
      "configs": {
        "getFactor": {"chainSpecificName": "getFactor"}
      },
      "contractABI": "[{\"inputs\":[],\"name\":\"getFactor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
    }
  }
}`

	jsonFullCompact = `{"contracts":{"MinimalConsumer":{"contractABI":"[{\"inputs\":[],\"name\":\"getFactor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]","contractPollingFilter":{"genericEventNames":["foo"],"pollingFilter":{"topic2":["bar","baz"],"topic3":["chain","link"],"topic4":["red","blue"],"retention":"1h0m0s","maxLogsKept":123,"logsPerBlock":456}},"configs":{"getFactor":{"chainSpecificName":"getFactor"}}}}}`
	jsonFullPretty  = `{
  "contracts": {
    "MinimalConsumer": {
      "configs": {
        "getFactor": {"chainSpecificName": "getFactor"}
      },
      "contractABI": "[{\"inputs\":[],\"name\":\"getFactor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
      "contractPollingFilter": {
        "genericEventNames": ["foo"],
        "pollingFilter": {
          "topic2": ["bar", "baz"],
          "topic3": ["chain", "link"],
          "topic4": ["red", "blue"],
          "retention": "1h",
          "maxLogsKept": 123,
          "logsPerBlock": 456
        }
      }
    }
  }
}`
)

var (
	crConfig = evm.ContractReaderConfig{
		Contracts: map[string]evm.ChainContractReader{
			"MinimalConsumer": {
				ContractABI: `[{"inputs":[],"name":"getFactor","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`,
				Configs: map[string]*evm.ChainReaderDefinition{
					"getFactor": {ChainSpecificName: "getFactor"},
				},
			},
		},
	}
	crConfigFull = evm.ContractReaderConfig{
		Contracts: map[string]evm.ChainContractReader{
			"MinimalConsumer": {
				ContractABI: `[{"inputs":[],"name":"getFactor","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`,
				ContractPollingFilter: evm.ContractPollingFilter{
					GenericEventNames: []string{"foo"},
					PollingFilter: evm.PollingFilter{
						Topic2:       []string{"bar", "baz"},
						Topic3:       []string{"chain", "link"},
						Topic4:       []string{"red", "blue"},
						Retention:    *config.MustNewDuration(time.Hour),
						MaxLogsKept:  123,
						LogsPerBlock: 456,
					},
				},
				Configs: map[string]*evm.ChainReaderDefinition{
					"getFactor": {ChainSpecificName: "getFactor"},
				},
			},
		},
	}
)

func TestContractReaderConfig(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		jsonCompact, jsonPretty string
		want                    evm.ContractReaderConfig
	}{
		{"minimal", jsonCompact, jsonPretty, crConfig},
		{"full", jsonFullCompact, jsonFullPretty, crConfigFull},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got evm.ContractReaderConfig
			require.NoError(t, json.Unmarshal([]byte(tt.jsonPretty), &got))
			t.Logf("%#v", got)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}

			b, err := json.Marshal(got)
			require.NoError(t, err)
			require.Equal(t, tt.jsonCompact, string(b))
		})
	}
}

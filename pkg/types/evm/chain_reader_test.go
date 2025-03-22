package evm_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/types/evm"
)

const (
	jsonCompact = `{"contracts":{"MinimalConsumer":{"contractABI":"[{\"inputs\":[],\"name\":\"getFactor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]","contractPollingFilter":{"genericEventNames":null,"pollingFilter":{"topic2":null,"topic3":null,"topic4":null,"retention":"0s","maxLogsKept":0,"logsPerBlock":0}},"configs":{"getFactor":{"chainSpecificName":"getFactor"}}}}}`
	jsonPretty  = `{
  "contracts": {
    "MinimalConsumer": {
      "configs": {
        "getFactor": {"chainSpecificName": "getFactor"}
      },
      "contractABI": "[{\"inputs\":[],\"name\":\"getFactor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
      "contractPollingFilter": {
        "genericEventNames": null,
        "pollingFilter": {
          "topic2": null,
          "topic3": null,
          "topic4": null,
          "retention": "0s",
          "maxLogsKept": 0,
          "logsPerBlock": 0
        }
      }
    }
  }
}`
)

func TestContractReaderConfig(t *testing.T) {
	var got evm.ContractReaderConfig
	require.NoError(t, json.Unmarshal([]byte(jsonPretty), &got))
	t.Logf("%#v", got)

	//TODO fill in nils
	want := evm.ContractReaderConfig{
		Contracts: map[string]evm.ChainContractReader{
			"MinimalConsumer": {
				ContractABI: `[{"inputs":[],"name":"getFactor","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`,
				ContractPollingFilter: evm.ContractPollingFilter{
					GenericEventNames: []string(nil),
					PollingFilter: evm.PollingFilter{
						Topic2:       []string(nil),
						Topic3:       []string(nil),
						Topic4:       []string(nil),
						Retention:    *config.MustNewDuration(0),
						MaxLogsKept:  0x0,
						LogsPerBlock: 0x0,
					},
				},
				Configs: map[string]*evm.ChainReaderDefinition{
					"getFactor": {ChainSpecificName: "getFactor"},
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}

	b, err := json.Marshal(got)
	require.NoError(t, err)
	require.Equal(t, jsonCompact, string(b))
}

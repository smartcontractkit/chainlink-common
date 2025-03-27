package evm

import (
	"github.com/smartcontractkit/chainlink-common/pkg/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

type ContractReaderConfig struct {
	// Contracts key is contract name
	Contracts map[string]ChainContractReader `json:"contracts" toml:"contracts"`
}

type ChainContractReader struct {
	ContractABI           string                `json:"contractABI" toml:"contractABI"`
	ContractPollingFilter ContractPollingFilter `json:"contractPollingFilter,omitempty,omitzero" toml:"contractPollingFilter,omitempty"`
	// key is genericName from config
	Configs map[string]*ChainReaderDefinition `json:"configs" toml:"configs"`
}

type EventDefinitions struct {
	// GenericTopicNames helps QueryingKeys not rely on EVM specific topic names. Key is chain specific name, value is generic name.
	// This helps us translate chain agnostic querying key "transfer-value" to EVM specific "evmTransferEvent-weiAmountTopic".
	GenericTopicNames map[string]string `json:"genericTopicNames,omitempty"`
	// GenericDataWordDetails key is generic name for evm log event data word that maps to chain details.
	// For e.g. first evm data word(32bytes) of USDC log event is value so the key can be called value.
	GenericDataWordDetails map[string]DataWordDetail `json:"genericDataWordDetails,omitempty"`
	// PollingFilter should be defined on a contract level in ContractPollingFilter,
	// unless event needs to override the contract level filter options.
	// This will create a separate log poller filter for this event.
	PollingFilter *PollingFilter `json:"pollingFilter,omitempty"`
}

type ChainReaderDefinition struct {
	CacheEnabled bool `json:"cacheEnabled,omitempty"`
	// chain specific contract method name or event type.
	ChainSpecificName   string                `json:"chainSpecificName"`
	ReadType            string                `json:"readType,omitempty"`
	InputModifications  codec.ModifiersConfig `json:"inputModifications,omitempty"`
	OutputModifications codec.ModifiersConfig `json:"outputModifications,omitempty"`
	EventDefinitions    *EventDefinitions     `json:"eventDefinitions,omitempty" toml:"eventDefinitions,omitempty"`
	// ConfidenceConfirmations is a mapping between a ConfidenceLevel and the confirmations associated. Confidence levels
	// should be valid float values.
	ConfidenceConfirmations map[string]int `json:"confidenceConfirmations,omitempty"`
}

type ContractPollingFilter struct {
	GenericEventNames []string      `json:"genericEventNames"`
	PollingFilter     PollingFilter `json:"pollingFilter"`
}

type PollingFilter struct {
	Topic2       []string        `json:"topic2"`       // list of possible values for topic2
	Topic3       []string        `json:"topic3"`       // list of possible values for topic3
	Topic4       []string        `json:"topic4"`       // list of possible values for topic4
	Retention    config.Duration `json:"retention"`    // maximum amount of time to retain logs
	MaxLogsKept  uint64          `json:"maxLogsKept"`  // maximum number of logs to retain ( 0 = unlimited )
	LogsPerBlock uint64          `json:"logsPerBlock"` // rate limit ( maximum # of logs per block, 0 = unlimited )
}
type DataWordDetail struct {
	Name string `json:"name"`
	// Index is indexed from 0. Index should only be used as an override in specific edge case scenarios where the index can't be programmatically calculated, otherwise leave this as nil.
	Index *int `json:"index,omitempty"`
	// Type should follow the geth ABI types naming convention
	Type string `json:"type,omitempty"`
}

package solana

import "github.com/smartcontractkit/chainlink-common/pkg/codec"

// nolint // ignoring naming suggestion
type ContractWriterConfig struct {
	Programs map[string]ProgramConfig `json:"programs"`
}

type ProgramConfig struct {
	Methods map[string]MethodConfig `json:"methods"`
	IDL     string                  `json:"idl"`
}

type MethodConfig struct {
	FromAddress        string                `json:"fromAddress"`
	InputModifications codec.ModifiersConfig `json:"inputModifications,omitempty"`
	ChainSpecificName  string                `json:"chainSpecificName"`
	LookupTables       LookupTables          `json:"lookupTables"`
	Accounts           []Lookup              `json:"accounts"`
	ATAs               []ATALookup           `json:"atas,omitempty"`
	// Location in the args where the debug ID is stored
	DebugIDLocation string `json:"debugIDLocation,omitempty"`
	ArgsTransform   string `json:"argsTransform,omitempty"`
	// Overhead added to calculated compute units in the args transform
	ComputeUnitLimitOverhead uint32 `json:"ComputeUnitLimitOverhead,omitempty"`
	// Configs for buffering payloads to support larger transaction sizes for this method
	BufferPayloadMethod string `json:"bufferPayloadMethod,omitempty"`
}

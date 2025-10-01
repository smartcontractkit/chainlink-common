package ccipocr3

import (
	"fmt"

	chainsel "github.com/smartcontractkit/chain-selectors"
)

// Codec is an interface that defines the methods for chain family specific encoding and decoding various types of data used in CCIP OCR3
type Codec struct {
	ChainSpecificAddressCodec
	CommitPluginCodec
	ExecutePluginCodec
	TokenDataEncoder
	SourceChainExtraDataCodec
}

// ChainSpecificAddressCodec is an interface that defines the methods for encoding and decoding addresses for a specific chain
type ChainSpecificAddressCodec interface {
	// AddressBytesToString converts an address from bytes to string
	AddressBytesToString([]byte) (string, error)
	// AddressStringToBytes converts an address from string to bytes
	AddressStringToBytes(string) ([]byte, error)
	// OracleIDAsAddressBytes returns a valid address for this chain family with the bytes set to the given oracle ID.
	OracleIDAsAddressBytes(oracleID uint8) ([]byte, error)
	// TransmitterBytesToString converts a transmitter account from bytes to string
	TransmitterBytesToString([]byte) (string, error)
}

// SourceChainExtraDataCodec is an interface for decoding source chain specific extra args and dest exec data into a map[string]any representation for a specific chain
// For chain A to chain B message, this interface will be the chain A specific codec
type SourceChainExtraDataCodec interface {
	// DecodeExtraArgsToMap reformat bytes into a chain agnostic map[string]any representation for extra args
	DecodeExtraArgsToMap(extraArgs Bytes) (map[string]any, error)
	// DecodeDestExecDataToMap reformat bytes into a chain agnostic map[string]interface{} representation for dest exec data
	DecodeDestExecDataToMap(destExecData Bytes) (map[string]any, error)
}

// ExtraDataCodecBundle is an interface that defines methods for decoding extra args and dest exec data.
type ExtraDataCodecBundle interface {
	DecodeExtraArgs(extraArgs Bytes, sourceChainSelector ChainSelector) (map[string]any, error)
	DecodeTokenAmountDestExecData(destExecData Bytes, sourceChainSelector ChainSelector) (map[string]any, error)
}

// ExtraDataCodecMap is a map of chain family to SourceChainExtraDataCodec
type ExtraDataCodecMap map[string]SourceChainExtraDataCodec

var _ ExtraDataCodecBundle = ExtraDataCodecMap{}

// DecodeExtraArgs reformats bytes into a chain agnostic map[string]any representation for extra args
func (c ExtraDataCodecMap) DecodeExtraArgs(extraArgs Bytes, sourceChainSelector ChainSelector) (map[string]any, error) {
	if len(extraArgs) == 0 {
		// return empty map if extraArgs is empty
		return nil, nil
	}

	family, err := chainsel.GetSelectorFamily(uint64(sourceChainSelector))
	if err != nil {
		return nil, fmt.Errorf("failed to get chain family for selector %d: %w", sourceChainSelector, err)
	}

	codec, exist := c[family]
	if !exist {
		return nil, fmt.Errorf("unsupported family for extra args type %s", family)
	}

	return codec.DecodeExtraArgsToMap(extraArgs)
}

// DecodeTokenAmountDestExecData reformats bytes to chain-agnostic map[string]any for tokenAmount DestExecData field
func (c ExtraDataCodecMap) DecodeTokenAmountDestExecData(destExecData Bytes, sourceChainSelector ChainSelector) (map[string]any, error) {
	if len(destExecData) == 0 {
		// return empty map if destExecData is empty
		return nil, nil
	}

	family, err := chainsel.GetSelectorFamily(uint64(sourceChainSelector))
	if err != nil {
		return nil, fmt.Errorf("failed to get chain family for selector %d: %w", sourceChainSelector, err)
	}

	codec, exist := c[family]
	if !exist {
		return nil, fmt.Errorf("unsupported family for dest exec data type %s", family)
	}

	return codec.DecodeDestExecDataToMap(destExecData)
}

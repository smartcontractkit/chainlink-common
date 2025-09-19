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

// AddressCodecMap is a map of chain family to ChainSpecificAddressCodec
type AddressCodecMap map[string]ChainSpecificAddressCodec

var _ AddressCodecBundle = AddressCodecMap{}

// AddressBytesToString converts an address from bytes to string
func (ac AddressCodecMap) AddressBytesToString(addr UnknownAddress, chainSelector ChainSelector) (string, error) {
	family, err := chainsel.GetSelectorFamily(uint64(chainSelector))
	if err != nil {
		return "", fmt.Errorf("failed to get chain family for selector %d: %w", chainSelector, err)
	}

	codec, exist := ac[family]
	if !exist {
		return "", fmt.Errorf("unsupported family for address decode type %s", family)
	}

	return codec.AddressBytesToString(addr)
}

// TransmitterBytesToString converts a transmitter account from bytes to string
func (ac AddressCodecMap) TransmitterBytesToString(addr UnknownAddress, chainSelector ChainSelector) (string, error) {
	family, err := chainsel.GetSelectorFamily(uint64(chainSelector))
	if err != nil {
		return "", fmt.Errorf("failed to get chain family for selector %d: %w", chainSelector, err)
	}

	codec, exist := ac[family]
	if !exist {
		return "", fmt.Errorf("unsupported family for transmitter decode type %s", family)
	}

	return codec.TransmitterBytesToString(addr)
}

// AddressStringToBytes converts an address from string to bytes
func (ac AddressCodecMap) AddressStringToBytes(addr string, chainSelector ChainSelector) (UnknownAddress, error) {
	family, err := chainsel.GetSelectorFamily(uint64(chainSelector))
	if err != nil {
		return nil, fmt.Errorf("failed to get chain family for selector %d: %w", chainSelector, err)
	}
	codec, exist := ac[family]
	if !exist {
		return nil, fmt.Errorf("unsupported family for address decode type %s", family)
	}

	return codec.AddressStringToBytes(addr)
}

// OracleIDAsAddressBytes returns valid address bytes for a given chain selector and oracle ID.
// Used for making nil transmitters in the OCR config valid, it just means that this oracle does not support the destination chain.
func (ac AddressCodecMap) OracleIDAsAddressBytes(oracleID uint8, chainSelector ChainSelector) ([]byte, error) {
	family, err := chainsel.GetSelectorFamily(uint64(chainSelector))
	if err != nil {
		return nil, fmt.Errorf("failed to get chain family for selector %d: %w", chainSelector, err)
	}
	codec, exist := ac[family]
	if !exist {
		return nil, fmt.Errorf("unsupported family for address decode type %s", family)
	}

	return codec.OracleIDAsAddressBytes(oracleID)
}

package ccipocr3

import (
	"context"
)

// TODO: Consolidate CommitPluginCodec, ExecutePluginCodec, ExtraDataCodec into a single Codec interface.

type CommitPluginCodec interface {
	Encode(context.Context, CommitPluginReport) ([]byte, error)
	Decode(context.Context, []byte) (CommitPluginReport, error)
}

type ExecutePluginCodec interface {
	Encode(context.Context, ExecutePluginReport) ([]byte, error)
	Decode(context.Context, []byte) (ExecutePluginReport, error)
}

type MessageHasher interface {
	Hash(context.Context, Message) (Bytes32, error)
}

type AddressCodec interface {
	AddressBytesToString(UnknownAddress, ChainSelector) (string, error)
	AddressStringToBytes(string, ChainSelector) (UnknownAddress, error)
}

type AddressCodecBundle interface {
	AddressBytesToString(addr UnknownAddress, chainSelector ChainSelector) (string, error)
	AddressStringToBytes(addr string, chainSelector ChainSelector) (UnknownAddress, error)
	TransmitterBytesToString(addr UnknownAddress, chainSelector ChainSelector) (string, error)
	OracleIDAsAddressBytes(oracleID uint8, chainSelector ChainSelector) ([]byte, error)
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

// ExtraDataCodecBundle is an interface that defines methods for decoding extra args and dest exec data.
type ExtraDataCodecBundle interface {
	DecodeExtraArgs(extraArgs Bytes, sourceChainSelector ChainSelector) (map[string]any, error)
	DecodeTokenAmountDestExecData(destExecData Bytes, sourceChainSelector ChainSelector) (map[string]any, error)
}

// SourceChainExtraDataCodec is an interface for decoding source chain specific extra args and dest exec data into a map[string]any representation for a specific chain
// For chain A to chain B message, this interface will be the chain A specific codec
type SourceChainExtraDataCodec interface {
	// DecodeExtraArgsToMap reformat bytes into a chain agnostic map[string]any representation for extra args
	DecodeExtraArgsToMap(extraArgs Bytes) (map[string]any, error)
	// DecodeDestExecDataToMap reformat bytes into a chain agnostic map[string]interface{} representation for dest exec data
	DecodeDestExecDataToMap(destExecData Bytes) (map[string]any, error)
}

// RMNCrypto provides a chain-agnostic interface for verifying RMN signatures.
// For example, on EVM, RMN reports are abi-encoded prior to being signed.
// On Solana, they would be borsh encoded instead, etc.
type RMNCrypto interface {
	// VerifyReportSignatures verifies each provided signature against the provided report and the signer addresses.
	// If any signature is invalid (no matching signer address is found), an error is returned immediately.
	VerifyReportSignatures(
		ctx context.Context,
		sigs []RMNECDSASignature,
		report RMNReport,
		signerAddresses []UnknownAddress,
	) error
}

// TokenDataEncoder is a generic interface for encoding offchain token data for different chain families.
// Right now it supports only USDC/CCTP, but every new token that requires offchain data processsing
// should be added to that interface and implemented in the downstream repositories (e.g. chainlink-ccip, chainlink).
type TokenDataEncoder interface {
	EncodeUSDC(ctx context.Context, message Bytes, attestation Bytes) (Bytes, error)
}

// EstimateProvider is used to estimate the gas cost of a message or a merkle tree.
type EstimateProvider interface {
	CalculateMerkleTreeGas(numRequests int) uint64
	CalculateMessageMaxGas(msg Message) uint64
}

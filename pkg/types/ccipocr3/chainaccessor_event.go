package ccipocr3

import (
	"math/big"
)

// ---------------------------------------------------
// The following types match the structs defined in the EVM contracts are used to decode these
// on-chain events.

// SendRequestedEvent represents the contents of the event emitted by the CCIP OnRamp when a
// message is sent.
type SendRequestedEvent struct {
	DestChainSelector ChainSelector
	SequenceNumber    SeqNum
	Message           Message
}

// CommitReportAcceptedEvent represents the contents of the event emitted by the CCIP OffRamp when a
// commit report is accepted.
type CommitReportAcceptedEvent struct {
	BlessedMerkleRoots   []MerkleRoot
	UnblessedMerkleRoots []MerkleRoot
	PriceUpdates         AccessorPriceUpdates
}

// ExecutionStateChangedEvent represents the contents of the event emitted by the CCIP OffRamp
type ExecutionStateChangedEvent struct {
	SourceChainSelector ChainSelector
	SequenceNumber      SeqNum
	MessageID           Bytes32
	MessageHash         Bytes32
	State               uint8
	ReturnData          Bytes
	GasUsed             big.Int
}

type MerkleRoot struct {
	SourceChainSelector uint64
	OnRampAddress       UnknownAddress
	MinSeqNr            uint64
	MaxSeqNr            uint64
	MerkleRoot          Bytes32
}

type TokenPriceUpdate struct {
	SourceToken UnknownAddress
	UsdPerToken *big.Int
}

type GasPriceUpdate struct {
	// DestChainSelector is the chain that the gas price is for (some plugin source chain).
	// Not the chain that the gas price is stored on.
	DestChainSelector uint64
	UsdPerUnitGas     *big.Int
}

type AccessorPriceUpdates struct {
	TokenPriceUpdates []TokenPriceUpdate
	GasPriceUpdates   []GasPriceUpdate
}

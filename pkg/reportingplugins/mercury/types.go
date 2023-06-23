package mercury

import (
	"context"
	"math/big"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// todo: group items by function in separate interfaces
type ParsedObservation interface {
	GetTimestamp() uint32
	GetObserver() commontypes.OracleID
	GetBenchmarkPrice() *big.Int
	GetBid() *big.Int
	GetAsk() *big.Int
	GetPricesValid() bool
	GetCurrentBlockNum() int64
	GetCurrentBlockHash() []byte
	GetCurrentBlockTimestamp() uint64
	GetCurrentBlockValid() bool
	GetMaxFinalizedBlockNumber() int64
	GetMaxFinalizedBlockNumberValid() bool
	GetMaxFinalizedTimestamp() int64
	GetMaxFinalizedTimestampValid() bool
}

type ObsResult[T any] struct {
	Val T
	Err error
}

// ReportCodec All functions on ReportCodec should be pure and thread-safe.
// Be careful validating and parsing any data passed.
type ReportCodec interface {
	// BuildReport Implementers may assume that there is at most one
	// ParsedAttributedObservation per observer, and that all observers are
	// valid. However, observation values, timestamps, etc... should all be
	// treated as untrusted.
	BuildReport(paos []ParsedObservation, f int, validFromBlockNum int64) (ocrtypes.Report, error)

	// MaxReportLength Returns the maximum length of a report based on n, the number of oracles.
	// The output of BuildReport must respect this maximum length.
	MaxReportLength(n int) (int, error)

	// CurrentBlockNumFromReport returns the median current block number from a report
	CurrentBlockNumFromReport(types.Report) (int64, error)
}

type OnchainConfigCodec interface {
	Encode(OnchainConfig) ([]byte, error)
	Decode([]byte) (OnchainConfig, error)
}

type Fetcher interface {
	// FetchInitialMaxFinalizedBlockNumber should fetch the initial max
	// finalized block number from the mercury server.
	FetchInitialMaxFinalizedBlockNumber(context.Context) (*int64, error)
}

type Transmitter interface {
	Fetcher
	// NOTE: Mercury doesn't actually transmit on-chain, so there is no
	// "contract" involved with the transmitter.
	// - Transmit should be implemented and send to Mercury server
	// - LatestConfigDigestAndEpoch is a stub method, does not need to do anything
	// - FromAccount() should return CSA public key
	ocrtypes.ContractTransmitter
}

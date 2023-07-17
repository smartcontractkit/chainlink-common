package mercury

import (
	"context"
	"math/big"

	"github.com/smartcontractkit/libocr/commontypes"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type IParsedAttributedObservation interface {
	GetTimestamp() uint32
	GetObserver() commontypes.OracleID
	GetBenchmarkPrice() (*big.Int, bool)
	GetBid() (*big.Int, bool)
	GetAsk() (*big.Int, bool)

	GetLinkFee() (*big.Int, bool)
	GetNativeFee() (*big.Int, bool)
}

type ObsResult[T any] struct {
	Val T
	Err error
}

type OnchainConfigCodec interface {
	Encode(OnchainConfig) ([]byte, error)
	Decode([]byte) (OnchainConfig, error)
}

type Fetcher interface {
	// FetchInitialMaxFinalizedBlockNumber should fetch the initial max
	// finalized block number from the mercury server.
	FetchInitialMaxFinalizedBlockNumber(context.Context) (*int64, error)
	LatestPrice(ctx context.Context, feedID [32]byte) (*big.Int, error)
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

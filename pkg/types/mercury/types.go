package mercury

import (
	"context"
	"fmt"
	"math/big"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type ObsResult[T any] struct {
	Val T
	Err error
}

type OnchainConfig struct {
	// applies to all values: price, bid and ask
	Min *big.Int
	Max *big.Int
}

type OnchainConfigCodec interface {
	Encode(context.Context, OnchainConfig) ([]byte, error)
	Decode(context.Context, []byte) (OnchainConfig, error)
}

type ServerFetcher interface {
	// FetchInitialMaxFinalizedBlockNumber should fetch the initial max finalized block number
	FetchInitialMaxFinalizedBlockNumber(context.Context) (*int64, error)
	LatestPrice(ctx context.Context, feedID [32]byte) (*big.Int, error)
	LatestTimestamp(context.Context) (int64, error)
}

type Transmitter interface {
	ServerFetcher
	// NOTE: Mercury doesn't actually transmit on-chain, so there is no
	// "contract" involved with the transmitter.
	// - Transmit should be implemented and send to Mercury server
	// - LatestConfigDigestAndEpoch is a stub method, does not need to do anything
	// - FromAccount() should return CSA public key
	ocrtypes.ContractTransmitter
}

type ChainReader interface {
	// LatestHeads returns an ordered list of the latest specified number of heads
	LatestHeads(context.Context, int) ([]Head, error)
}

type Head struct {
	Number    uint64
	Hash      []byte
	Timestamp uint64
}

type MercuryTransmitterProtocol string

const (
	MercuryTransmitterProtocolWSRPC MercuryTransmitterProtocol = "wsrpc"
	MercuryTransmitterProtocolGRPC  MercuryTransmitterProtocol = "grpc"
)

func (m MercuryTransmitterProtocol) String() string {
	return string(m)
}

func (m *MercuryTransmitterProtocol) UnmarshalText(text []byte) error {
	switch string(text) {
	case "wsrpc":
		*m = MercuryTransmitterProtocolWSRPC
	case "grpc":
		*m = MercuryTransmitterProtocolGRPC
	default:
		return fmt.Errorf("unknown mercury transmitter protocol: %s", text)
	}
	return nil
}

package mercury_v0

import (
	"math/big"

	"github.com/smartcontractkit/chainlink-relay/pkg/reportingplugins/mercury"
	"github.com/smartcontractkit/libocr/commontypes"
)

var _ mercury.ParsedObservation = ParsedAttributedObservation{}

type ParsedAttributedObservation struct {
	Timestamp uint32
	Observer  commontypes.OracleID

	BenchmarkPrice *big.Int
	Bid            *big.Int
	Ask            *big.Int
	// All three prices must be valid, or none are (they all should come from one API query and hold invariant bid <= bm <= ask)
	PricesValid bool

	CurrentBlockNum       int64 // inclusive; current block
	CurrentBlockHash      []byte
	CurrentBlockTimestamp uint64
	// All three block observations must be valid, or none are (they all come from the same block)
	CurrentBlockValid bool

	// MaxFinalizedBlockNumber comes from previous report when present and is
	// only observed from mercury server when previous report is nil
	//
	// MaxFinalizedBlockNumber will be -1 if there is none
	MaxFinalizedBlockNumber      int64
	MaxFinalizedBlockNumberValid bool
}

func (pao ParsedAttributedObservation) GetTimestamp() uint32 {
	return pao.Timestamp
}

func (pao ParsedAttributedObservation) GetObserver() commontypes.OracleID {
	return pao.Observer
}

func (pao ParsedAttributedObservation) GetBenchmarkPrice() *big.Int {
	return pao.BenchmarkPrice
}

func (pao ParsedAttributedObservation) GetBid() *big.Int {
	return pao.Bid
}

func (pao ParsedAttributedObservation) GetAsk() *big.Int {
	return pao.Ask
}

func (pao ParsedAttributedObservation) GetPricesValid() bool {
	return pao.PricesValid
}

func (pao ParsedAttributedObservation) GetCurrentBlockNum() int64 {
	return pao.CurrentBlockNum
}

func (pao ParsedAttributedObservation) GetCurrentBlockHash() []byte {
	return pao.CurrentBlockHash
}

func (pao ParsedAttributedObservation) GetCurrentBlockTimestamp() uint64 {
	return pao.CurrentBlockTimestamp
}

func (pao ParsedAttributedObservation) GetCurrentBlockValid() bool {
	return pao.CurrentBlockValid
}

func (pao ParsedAttributedObservation) GetMaxFinalizedTimestamp() uint32 {
	panic("current observation doesn't contain the field")
}

func (pao ParsedAttributedObservation) GetMaxFinalizedTimestampValid() bool {
	panic("current observation doesn't contain the field")
}

func (pao ParsedAttributedObservation) GetMaxFinalizedBlockNumber() int64 {
	return pao.MaxFinalizedBlockNumber
}

func (pao ParsedAttributedObservation) GetMaxFinalizedBlockNumberValid() bool {
	return pao.MaxFinalizedBlockNumberValid
}

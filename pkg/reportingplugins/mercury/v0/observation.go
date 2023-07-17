package mercury_v0

import (
	"math/big"

	"github.com/smartcontractkit/libocr/commontypes"
)

var _ IParsedAttributedObservation = ParsedAttributedObservation{}

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

func (pao ParsedAttributedObservation) GetBenchmarkPrice() (*big.Int, bool) {
	return pao.BenchmarkPrice, pao.PricesValid
}

func (pao ParsedAttributedObservation) GetBid() (*big.Int, bool) {
	return pao.Bid, pao.PricesValid
}

func (pao ParsedAttributedObservation) GetAsk() (*big.Int, bool) {
	return pao.Ask, pao.PricesValid
}

func (pao ParsedAttributedObservation) GetCurrentBlockNum() (int64, bool) {
	return pao.CurrentBlockNum, pao.CurrentBlockValid
}

func (pao ParsedAttributedObservation) GetCurrentBlockHash() ([]byte, bool) {
	return pao.CurrentBlockHash, pao.CurrentBlockValid
}

func (pao ParsedAttributedObservation) GetCurrentBlockTimestamp() (uint64, bool) {
	return pao.CurrentBlockTimestamp, pao.CurrentBlockValid
}

func (pao ParsedAttributedObservation) GetMaxFinalizedBlockNumber() (int64, bool) {
	return pao.MaxFinalizedBlockNumber, pao.MaxFinalizedBlockNumberValid
}

func (pao ParsedAttributedObservation) GetLinkFee() (*big.Int, bool) {
	panic("current observation doesn't contain the field")
}

func (pao ParsedAttributedObservation) GetNativeFee() (*big.Int, bool) {
	panic("current observation doesn't contain the field")
}

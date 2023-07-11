package mercury_v1

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
	PricesValid    bool

	MaxFinalizedTimestamp      uint32
	MaxFinalizedTimestampValid bool
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

func (pao ParsedAttributedObservation) GetMaxFinalizedTimestamp() uint32 {
	return pao.MaxFinalizedTimestamp
}

func (pao ParsedAttributedObservation) GetMaxFinalizedTimestampValid() bool {
	return pao.MaxFinalizedTimestampValid
}

func (pao ParsedAttributedObservation) GetCurrentBlockNum() int64 {
	panic("current observation doesn't contain the field")
}

func (pao ParsedAttributedObservation) GetCurrentBlockHash() []byte {
	panic("current observation doesn't contain the field")
}

func (pao ParsedAttributedObservation) GetCurrentBlockTimestamp() uint64 {
	panic("current observation doesn't contain the field")
}

func (pao ParsedAttributedObservation) GetCurrentBlockValid() bool {
	panic("current observation doesn't contain the field")
}

func (pao ParsedAttributedObservation) GetMaxFinalizedBlockNumber() int64 {
	panic("current observation doesn't contain the field")
}

func (pao ParsedAttributedObservation) GetMaxFinalizedBlockNumberValid() bool {
	panic("current observation doesn't contain the field")
}

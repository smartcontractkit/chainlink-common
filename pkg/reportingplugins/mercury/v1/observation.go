package mercury_v1

import (
	"math/big"

	"github.com/smartcontractkit/libocr/commontypes"
)

var _ ParsedAttributedObservation = parsedAttributedObservation{}

type parsedAttributedObservation struct {
	Timestamp uint32
	Observer  commontypes.OracleID

	BenchmarkPrice *big.Int
	Bid            *big.Int
	Ask            *big.Int
	PricesValid    bool

	MaxFinalizedTimestamp      uint32
	MaxFinalizedTimestmapValid bool

	LinkFee      *big.Int
	LinkFeeValid bool

	NativeFee      *big.Int
	NativeFeeValid bool
}

func (pao parsedAttributedObservation) GetTimestamp() uint32 {
	return pao.Timestamp
}

func (pao parsedAttributedObservation) GetObserver() commontypes.OracleID {
	return pao.Observer
}

func (pao parsedAttributedObservation) GetBenchmarkPrice() (*big.Int, bool) {
	return pao.BenchmarkPrice, pao.PricesValid
}

func (pao parsedAttributedObservation) GetBid() (*big.Int, bool) {
	return pao.Bid, pao.PricesValid
}

func (pao parsedAttributedObservation) GetAsk() (*big.Int, bool) {
	return pao.Ask, pao.PricesValid
}

func (pao parsedAttributedObservation) GetMaxFinalizedTimestamp() (uint32, bool) {
	return pao.MaxFinalizedTimestamp, pao.MaxFinalizedTimestmapValid
}

func (pao parsedAttributedObservation) GetLinkFee() (*big.Int, bool) {
	return pao.LinkFee, pao.LinkFeeValid
}

func (pao parsedAttributedObservation) GetNativeFee() (*big.Int, bool) {
	return pao.NativeFee, pao.NativeFeeValid
}

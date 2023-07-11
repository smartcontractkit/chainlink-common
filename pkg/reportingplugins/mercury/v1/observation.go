package mercury_v1

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
	PricesValid    bool

	LinkFee      *big.Int
	LinkFeeValid bool

	NativeFee      *big.Int
	NativeFeeValid bool
}

func (pao ParsedAttributedObservation) GetTimestamp() uint32 {
	return pao.Timestamp
}

func (pao ParsedAttributedObservation) GetObserver() commontypes.OracleID {
	return pao.Observer
}

// TODO: Change these to return (val, bool)
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

func (pao ParsedAttributedObservation) GetLinkFee() (*big.Int, bool) {
	return pao.LinkFee, pao.LinkFeeValid
}

func (pao ParsedAttributedObservation) GetNativeFee() (*big.Int, bool) {
	return pao.NativeFee, pao.NativeFeeValid
}

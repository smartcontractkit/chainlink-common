package mercury

import (
	"math/big"
	"sort"

	"github.com/pkg/errors"
)

// NOTE: All aggregate functions assume at least one element in the passed slice
// The passed slice might be mutated (sorted)

// GetConsensusTimestamp gets the median timestamp
func GetConsensusTimestamp(paos []IParsedAttributedObservation) uint32 {
	sort.Slice(paos, func(i, j int) bool {
		return paos[i].GetTimestamp() < paos[j].GetTimestamp()
	})
	return paos[len(paos)/2].GetTimestamp()
}

// GetConsensusBenchmarkPrice gets the median benchmark price
func GetConsensusBenchmarkPrice(paos []IParsedAttributedObservation, f int) (*big.Int, error) {
	var validBenchmarkPrices []*big.Int
	for _, pao := range paos {
		if pao.GetPricesValid() {
			validBenchmarkPrices = append(validBenchmarkPrices, pao.GetBenchmarkPrice())
		}
	}
	if len(validBenchmarkPrices) < f+1 {
		return nil, errors.New("fewer than f+1 observations have a valid price")
	}
	sort.Slice(validBenchmarkPrices, func(i, j int) bool {
		return validBenchmarkPrices[i].Cmp(validBenchmarkPrices[j]) < 0
	})

	return validBenchmarkPrices[len(validBenchmarkPrices)/2], nil
}

// GetConsensusBid gets the median bid
func GetConsensusBid(paos []IParsedAttributedObservation, f int) (*big.Int, error) {
	var validBids []*big.Int
	for _, pao := range paos {
		if pao.GetPricesValid() {
			validBids = append(validBids, pao.GetBid())
		}
	}
	if len(validBids) < f+1 {
		return nil, errors.New("fewer than f+1 observations have a valid price")
	}
	sort.Slice(validBids, func(i, j int) bool {
		return validBids[i].Cmp(validBids[j]) < 0
	})

	return validBids[len(validBids)/2], nil
}

// GetConsensusAsk gets the median ask
func GetConsensusAsk(paos []IParsedAttributedObservation, f int) (*big.Int, error) {
	var validAsks []*big.Int
	for _, pao := range paos {
		if pao.GetPricesValid() {
			validAsks = append(validAsks, pao.GetAsk())
		}
	}
	if len(validAsks) < f+1 {
		return nil, errors.New("fewer than f+1 observations have a valid price")
	}
	sort.Slice(validAsks, func(i, j int) bool {
		return validAsks[i].Cmp(validAsks[j]) < 0
	})

	return validAsks[len(validAsks)/2], nil
}

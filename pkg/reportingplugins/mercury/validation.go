package mercury

import (
	"math/big"

	pkgerrors "github.com/pkg/errors"
)

// NOTE: hardcoded for now, this may need to change if we support block range on chains other than eth
const EvmHashLen = 32

func ValidateBenchmarkPrice(paos []IParsedAttributedObservation, f int, min, max *big.Int) error {
	answer, err := GetConsensusBenchmarkPrice(paos, f)
	if err != nil {
		return err
	}

	if !(min.Cmp(answer) <= 0 && answer.Cmp(max) <= 0) {
		return pkgerrors.Errorf("median benchmark price %s is outside of allowable range (Min: %s, Max: %s)", answer, min, max)
	}

	return nil
}

func ValidateBid(paos []IParsedAttributedObservation, f int, min, max *big.Int) error {
	answer, err := GetConsensusBid(paos, f)
	if err != nil {
		return err
	}

	if !(min.Cmp(answer) <= 0 && answer.Cmp(max) <= 0) {
		return pkgerrors.Errorf("median bid price %s is outside of allowable range (Min: %s, Max: %s)", answer, min, max)
	}

	return nil
}

func ValidateAsk(paos []IParsedAttributedObservation, f int, min, max *big.Int) error {
	answer, err := GetConsensusAsk(paos, f)
	if err != nil {
		return err
	}

	if !(min.Cmp(answer) <= 0 && answer.Cmp(max) <= 0) {
		return pkgerrors.Errorf("median ask price %s is outside of allowable range (Min: %s, Max: %s)", answer, min, max)
	}

	return nil
}

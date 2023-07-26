package mercury_v1

import (
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-relay/pkg/reportingplugins/mercury"
)

func ValidateValidFromTimestamp(paos []ParsedAttributedObservation) error {
	observationTimestamp := mercury.GetConsensusTimestamp(Convert(paos))
	validFromTimestamp := GetConsensusMaxFinalizedTimestamp(paos)

	if observationTimestamp <= validFromTimestamp {
		return errors.Errorf("observationTimestamp (%d) must be > validFromTimestamp (%d)", observationTimestamp, validFromTimestamp)
	}

	return nil
}

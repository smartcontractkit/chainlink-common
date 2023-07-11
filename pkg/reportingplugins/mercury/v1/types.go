package mercury_v1

import (
	"github.com/smartcontractkit/chainlink-relay/pkg/reportingplugins/mercury"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type IParsedAttributedObservation interface {
	mercury.IParsedAttributedObservation
}

func Convert(pao []IParsedAttributedObservation) []mercury.IParsedAttributedObservation {
	var ret []mercury.IParsedAttributedObservation
	for _, v := range pao {
		ret = append(ret, v)
	}
	return ret
}

// ReportCodec All functions on ReportCodec should be pure and thread-safe.
// Be careful validating and parsing any data passed.
type ReportCodec interface {
	// BuildReport Implementers may assume that there is at most one
	// ParsedAttributedObservation per observer, and that all observers are
	// valid. However, observation values, timestamps, etc... should all be
	// treated as untrusted.
	BuildReport(paos []IParsedAttributedObservation, f int, validFromTimestamp, expiresAt uint32) (ocrtypes.Report, error)

	// MaxReportLength Returns the maximum length of a report based on n, the number of oracles.
	// The output of BuildReport must respect this maximum length.
	MaxReportLength(n int) (int, error)

	ObservationTimestampFromReport(types.Report) (uint32, error)
}

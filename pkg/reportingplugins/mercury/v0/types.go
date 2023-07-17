package mercury_v0

import (
	"github.com/smartcontractkit/chainlink-relay/pkg/reportingplugins/mercury"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type IParsedAttributedObservation interface {
	mercury.IParsedAttributedObservation

	GetCurrentBlockNum() (int64, bool)
	GetCurrentBlockHash() ([]byte, bool)
	GetCurrentBlockTimestamp() (uint64, bool)
	GetMaxFinalizedBlockNumber() (int64, bool)
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
	BuildReport(paos []IParsedAttributedObservation, f int, validFromBlockNum int64) (ocrtypes.Report, error)

	// MaxReportLength Returns the maximum length of a report based on n, the number of oracles.
	// The output of BuildReport must respect this maximum length.
	MaxReportLength(n int) (int, error)

	// CurrentBlockNumFromReport returns the median current block number from a report
	CurrentBlockNumFromReport(types.Report) (int64, error)
}

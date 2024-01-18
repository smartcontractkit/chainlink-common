package v1_test

import (
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2/types"

	mercury_v1_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v1"
)

type StaticReportCodec struct{}

var _ mercury_v1_types.ReportCodec = StaticReportCodec{}

type StaticReportCodecValues struct {
	Report             ocrtypes.Report
	MaxReportLength    int
	CurrentBlockNumber int64
}

var StaticReportCodecFixtures = StaticReportCodecValues{
	Report:             ocrtypes.Report([]byte("mercury v1 report")),
	MaxReportLength:    20,
	CurrentBlockNumber: 23,
}

func (s StaticReportCodec) BuildReport(fields mercury_v1_types.ReportFields) (ocrtypes.Report, error) {
	return StaticReportCodecFixtures.Report, nil
}

// MaxReportLength Returns the maximum length of a report based on n, the number of oracles.
// The output of BuildReport must respect this maximum length.
func (s StaticReportCodec) MaxReportLength(n int) (int, error) {
	return StaticReportCodecFixtures.MaxReportLength, nil
}

// CurrentBlockNumFromReport returns the median current block number from a report
func (s StaticReportCodec) CurrentBlockNumFromReport(ocrtypes.Report) (int64, error) {
	return StaticReportCodecFixtures.CurrentBlockNumber, nil
}

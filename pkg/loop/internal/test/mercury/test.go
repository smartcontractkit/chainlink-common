package mercury_test

import (
	"github.com/smartcontractkit/libocr/commontypes"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

var (
	configDigest = libocr.ConfigDigest([32]byte{2: 10, 12: 16})
	obs          = []libocr.AttributedObservation{{Observation: []byte{21: 19}, Observer: commontypes.OracleID(99)}}

	previousReport = libocr.Report([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	report         = libocr.Report{42: 101}
	reportContext  = libocr.ReportContext{
		ReportTimestamp: libocr.ReportTimestamp{
			ConfigDigest: libocr.ConfigDigest([32]byte{1: 7, 31: 3}),
			Epoch:        79,
			Round:        17,
		},
		ExtraHash: [32]byte{1: 2, 3: 4, 5: 6},
	}
)

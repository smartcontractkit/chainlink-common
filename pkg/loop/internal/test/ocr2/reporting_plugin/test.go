package reportingplugin_test

import (
	"time"

	"github.com/smartcontractkit/libocr/commontypes"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

const (
	blockHeight    = uint64(1337)
	changedInBlock = uint64(14)
	epoch          = uint32(88)
	round          = uint8(74)
	shouldAccept   = true
	shouldReport   = true
	shouldTransmit = true
)

var (
	ConfigDigest       = libocr.ConfigDigest([32]byte{2: 10, 12: 16})
	configDigestPrefix = libocr.ConfigDigestPrefix(99)

	contractConfig = libocr.ContractConfig{
		ConfigDigest:          ConfigDigest,
		ConfigCount:           42,
		Signers:               []libocr.OnchainPublicKey{[]byte{15: 1}},
		Transmitters:          []libocr.Account{"foo", "bar"},
		F:                     11,
		OnchainConfig:         []byte{2: 11, 14: 22, 31: 1},
		OffchainConfigVersion: 2,
		OffchainConfig:        []byte{1: 99, 12: 55},
	}

	observation = libocr.Observation([]byte{21: 19})
	obs         = []libocr.AttributedObservation{{Observation: []byte{21: 19}, Observer: commontypes.OracleID(99)}}

	query = []byte{42: 42}

	report        = libocr.Report{42: 101}
	reportContext = libocr.ReportContext{
		ReportTimestamp: libocr.ReportTimestamp{
			ConfigDigest: ConfigDigest,
			Epoch:        epoch,
			Round:        round,
		},
		ExtraHash: [32]byte{1: 2, 3: 4, 5: 6},
	}

	reportingPluginConfig = libocr.ReportingPluginConfig{
		ConfigDigest:                            libocr.ConfigDigest{}, // pluginprovider_test.ConfigDigest,
		OracleID:                                commontypes.OracleID(10),
		N:                                       12,
		F:                                       42,
		OnchainConfig:                           []byte{17: 11},
		OffchainConfig:                          []byte{32: 64},
		EstimatedRoundInterval:                  time.Second,
		MaxDurationQuery:                        time.Hour,
		MaxDurationObservation:                  time.Millisecond,
		MaxDurationReport:                       time.Microsecond,
		MaxDurationShouldAcceptFinalizedReport:  10 * time.Second,
		MaxDurationShouldTransmitAcceptedReport: time.Minute,
	}

	rpi = libocr.ReportingPluginInfo{
		Name:          "test",
		UniqueReports: true,
		Limits: libocr.ReportingPluginLimits{
			MaxQueryLength:       42,
			MaxObservationLength: 13,
			MaxReportLength:      17,
		},
	}

	DefaultReportingPluginTestConfig = ReportingPluginTestConfig{
		ReportContext:          reportContext,
		Query:                  query,
		Observation:            observation,
		AttributedObservations: obs,
		Report:                 report,
		//	Sigs:           sigs,
		ShouldAccept:   shouldAccept,
		ShouldReport:   shouldReport,
		ShouldTransmit: shouldTransmit,
	}

	DefaultStaticReportingPlugin = StaticReportingPlugin{
		DefaultReportingPluginTestConfig,
	}
)

package median_test

import (
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	pluginprovider_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/plugin_provider"
)

const ConfigTOML = `[Foo]
Bar = "Baz"
`

const (
	errMsg           = "test error"
	lookbackDuration = time.Minute + 4*time.Second
	max              = 101
	n                = 12
)

var (
	MedianFactoryGeneratorImpl = staticMedianFactoryGenerator{
		staticPluginMedianConfig: staticPluginMedianConfig{
			provider:                  MedianProviderImpl,
			dataSource:                DataSourceImpl,
			juelsPerFeeCoinDataSource: JuelsPerFeeCoinDataSourceImpl,
			errorLog:                  StaticErrorLog{},
		},
	}

	MedianProviderImpl = staticMedianProvider{
		staticMedianProviderConfig: staticMedianProviderConfig{
			offchainDigester:    pluginprovider_test.OffchainConfigDigesterImpl,
			contractTracker:     pluginprovider_test.ContractConfigTrackerImpl,
			contractTransmitter: pluginprovider_test.ContractTransmitterImpl,
			reportCodec:         staticReportCodec{},
			medianContract: staticMedianContract{
				staticMedianContractConfig: staticMedianContractConfig{
					configDigest:     libocr.ConfigDigest([32]byte{1: 1, 11: 8}),
					epoch:            7,
					round:            11,
					latestAnswer:     big.NewInt(123),
					latestTimestamp:  time.Unix(1234567890, 987654321).UTC(),
					lookbackDuration: lookbackDuration,
				},
			},
			onchainConfigCodec: staticOnchainConfigCodec{},
			chainReader:        pluginprovider_test.ChainReaderImpl,
		},
	}
)

var (
	encodedOnchainConfig = []byte{5: 11}
	juelsPerFeeCoin      = big.NewInt(1234)
	onchainConfig        = median.OnchainConfig{Min: big.NewInt(-12), Max: big.NewInt(1234567890987654321)}
	medianValue          = big.NewInt(-1042)

	pobs = []median.ParsedAttributedObservation{{Timestamp: 123, Value: big.NewInt(31), JuelsPerFeeCoin: big.NewInt(54), Observer: commontypes.OracleID(99)}}

	report        = libocr.Report{42: 101}
	reportContext = libocr.ReportContext{
		ReportTimestamp: libocr.ReportTimestamp{
			ConfigDigest: libocr.ConfigDigest([32]byte{1: 7, 31: 3}),
			Epoch:        79,
			Round:        17,
		},
		ExtraHash: [32]byte{1: 2, 3: 4, 5: 6},
	}

	reportingPluginConfig = libocr.ReportingPluginConfig{
		ConfigDigest:                            libocr.ConfigDigest{}, //pluginprovider_test.ConfigDigest,
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
	value = big.NewInt(999)
)

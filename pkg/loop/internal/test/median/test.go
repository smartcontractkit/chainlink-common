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
	encodedOnchainConfig = []byte{5: 11}
	juelsPerFeeCoin      = big.NewInt(1234)
	onchainConfig        = median.OnchainConfig{Min: big.NewInt(-12), Max: big.NewInt(1234567890987654321)}
	latestAnswer         = big.NewInt(-66)
	observation          = libocr.Observation([]byte{21: 19})
	medianValue          = big.NewInt(-1042)

	obs  = []libocr.AttributedObservation{{Observation: []byte{21: 19}, Observer: commontypes.OracleID(99)}}
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
	value                       = big.NewInt(999)
	payload                     = []byte("oops")
	medianContractGenericMethod = "LatestTransmissionDetails"
	getLatestValueParams        = map[string]string{"param1": "value1", "param2": "value2"}
	contractName                = "my median contract"
	latestValue                 = map[string]int{"ret1": 1, "ret2": 2}

	/*
		//OCR3
		ocr3reportingPluginConfig = ocr3types.ReportingPluginConfig{
			ConfigDigest:                            pluginprovider_test.ConfigDigest,
			OracleID:                                commontypes.OracleID(10),
			N:                                       12,
			F:                                       42,
			OnchainConfig:                           []byte{17: 11},
			OffchainConfig:                          []byte{32: 64},
			EstimatedRoundInterval:                  time.Second,
			MaxDurationQuery:                        time.Hour,
			MaxDurationObservation:                  time.Millisecond,
			MaxDurationShouldAcceptAttestedReport:   10 * time.Second,
			MaxDurationShouldTransmitAcceptedReport: time.Minute,
		}

		ocr3rpi = ocr3types.ReportingPluginInfo{
			Name: "test",
			Limits: ocr3types.ReportingPluginLimits{
				MaxQueryLength:       42,
				MaxObservationLength: 13,
				MaxOutcomeLength:     33,
				MaxReportLength:      17,
				MaxReportCount:       41,
			},
		}

		outcomeContext = ocr3types.OutcomeContext{
			SeqNr:           1,
			PreviousOutcome: []byte("previous-outcome"),
			Epoch:           2,
			Round:           3,
		}
	*/
	/*
		ao      = libocr.AttributedObservation{Observation: []byte{21: 19}, Observer: commontypes.OracleID(99)}
		quorum  = ocr3types.Quorum(7)
		outcome = ocr3types.Outcome("outcome")
		seqNr   = uint64(43)
		RI      = ocr3types.ReportWithInfo[[]byte]{
			Report: []byte("report"),
			Info:   []byte("info"),
		}
		RIs = []ocr3types.ReportWithInfo[[]byte]{{
			Report: []byte("report1"),
			Info:   []byte("info1"),
		}, {
			Report: []byte("report2"),
			Info:   []byte("info2"),
		}}

		//CapabilitiesRegistry
		GetID          = "get-id"
		GetTriggerID   = "get-trigger-id"
		GetActionID    = "get-action-id"
		GetConsensusID = "get-consensus-id"
		GetTargetID    = "get-target-id"
		CapabilityInfo = capabilities.CapabilityInfo{
			ID:             "capability-info-id",
			CapabilityType: 2,
			Description:    "capability-info-description",
			Version:        "capability-info-version",
		}
	*/
)

/*
var _ capabilities.BaseCapability = (*baseCapability)(nil)

type baseCapability struct {
}

func (e baseCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	return CapabilityInfo, nil
}
*/

var DefaultPluginMedian = StaticPluginMedian{
	StaticPluginMedianConfig: StaticPluginMedianConfig{
		Provider:                  TestStaticMedianProvider,
		DataSource:                DefaultTestDataSource(),
		JuelsPerFeeCoinDataSource: DefaultTestJuelsPerFeeCoinDataSource(),
		ErrorLog:                  StaticErrorLog{},
	},
}

var TestStaticMedianProvider = StaticMedianProvider{
	StaticMedianProviderConfig: StaticMedianProviderConfig{
		OffchainDigester:    pluginprovider_test.TestOffchainConfigDigester,
		ContractTracker:     pluginprovider_test.TestContractConfigTracker,
		ContractTransmitter: pluginprovider_test.TestContractTransmitter,
	},
	rc: staticReportCodec{},
	mc: staticMedianContract{
		staticMedianContractConfig: staticMedianContractConfig{
			configDigest:     libocr.ConfigDigest([32]byte{1: 1, 11: 8}),
			epoch:            7,
			round:            11,
			latestAnswer:     big.NewInt(123),
			latestTimestamp:  time.Unix(1234567890, 987654321).UTC(),
			lookbackDuration: lookbackDuration,
		},
	},
	ooc: staticOnchainConfigCodec{},
	cr:  pluginprovider_test.TestChainReader,
}

package pluginprovider_test

import (
	reportingplugin_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/reporting_plugin"
	keystore_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources/keystore"
	"github.com/smartcontractkit/libocr/commontypes"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

const (
	//	account        = libocr.Account("testaccount")
	blockHeight    = uint64(1337)
	changedInBlock = uint64(14)
	epoch          = uint32(88)
	round          = uint8(74)
)

var (
	configDigest       = libocr.ConfigDigest([32]byte{2: 10, 16: 3})
	configDigestPrefix = libocr.ConfigDigestPrefix(99)

	contractConfig = libocr.ContractConfig{
		ConfigDigest:          configDigest,
		ConfigCount:           42,
		Signers:               []libocr.OnchainPublicKey{[]byte{15: 1}},
		Transmitters:          []libocr.Account{"foo", "bar"},
		F:                     11,
		OnchainConfig:         []byte{2: 11, 14: 22, 31: 1},
		OffchainConfigVersion: 2,
		OffchainConfig:        []byte{1: 99, 12: 55},
	}

	sigs = []libocr.AttributedOnchainSignature{{Signature: []byte{9: 8, 7: 6}, Signer: commontypes.OracleID(54)}}

	ReportTimestamp = libocr.ReportTimestamp{
		ConfigDigest: configDigest,
		Epoch:        epoch,
		Round:        round,
	}

	DefaultContractTransmitterTestConfig = ContractTransmitterTestConfig{
		ConfigDigest:  configDigest,
		Account:       libocr.Account(keystore_test.DefaultKeystoreTestConfig.Account),
		Epoch:         epoch,
		ReportContext: reportingplugin_test.DefaultReportingPluginTestConfig.ReportContext,
		Report:        reportingplugin_test.DefaultReportingPluginTestConfig.Report,
		Sigs:          sigs,
	}

	TestContractTransmitter = StaticContractTransmitter{
		ContractTransmitterTestConfig: DefaultContractTransmitterTestConfig,
	}

	offchainConfigDigesterTestConfig = OffchainConfigDigesterTestConfig{
		ContractConfig:     contractConfig,
		ConfigDigest:       configDigest,
		ConfigDigestPrefix: configDigestPrefix,
	}

	TestOffchainConfigDigester = StaticOffchainConfigDigester{
		OffchainConfigDigesterTestConfig: offchainConfigDigesterTestConfig,
	}

	contractConfigTrackerTestConfig = ContractConfigTrackerTestConfig{
		ContractConfig: contractConfig,
		ConfigDigest:   configDigest,
		ChangedInBlock: changedInBlock,
		BlockHeight:    blockHeight,
	}

	TestContractConfigTracker = StaticContractConfigTracker{
		ContractConfigTrackerTestConfig: contractConfigTrackerTestConfig,
	}

	configProviderTestConfig = ConfigProviderTestConfig{
		OffchainConfigDigesterTestConfig: offchainConfigDigesterTestConfig,
		ContractConfigTrackerTestConfig:  contractConfigTrackerTestConfig,
	}

	TestStaticConfigProvider = StaticConfigProvider{
		ConfigProviderTestConfig: configProviderTestConfig,
	}
)

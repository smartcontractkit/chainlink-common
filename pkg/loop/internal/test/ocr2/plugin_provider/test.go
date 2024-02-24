package pluginprovider_test

import (
	keystore_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources/keystore"
	"github.com/smartcontractkit/libocr/commontypes"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

const (
	blockHeight    = uint64(1337)
	changedInBlock = uint64(14)
	epoch          = uint32(88)
	round          = uint8(74)
)

var (
	// ContractTransmitterImpl is a static implementation of the ContractTransmitterTester interface for testing
	ContractTransmitterImpl = staticContractTransmitter{
		contractTransmitterTestConfig: contractTransmitterTestConfig{
			ConfigDigest:  configDigest,
			Account:       libocr.Account(keystore_test.DefaultKeystoreTestConfig.Account),
			Epoch:         epoch,
			ReportContext: libocr.ReportContext{ReportTimestamp: reportTimestamp, ExtraHash: [32]byte{1: 3, 3: 5, 7: 11}},
			Report:        libocr.Report{41: 131},
			Sigs:          sigs,
		},
	}

	// OffchainConfigDigesterImpl is a static implementation of the OffchainConfigDigesterTester interface for testing
	OffchainConfigDigesterImpl = staticOffchainConfigDigester{
		staticOffchainConfigDigesterConfig: staticOffchainConfigDigesterConfig{
			contractConfig:     contractConfig,
			configDigest:       configDigest,
			configDigestPrefix: configDigestPrefix,
		},
	}

	// ContractConfigTrackerImpl is a static implementation of the ContractConfigTrackerTester interface for testing
	ContractConfigTrackerImpl = staticContractConfigTracker{
		staticConfigTrackerConfig: staticConfigTrackerConfig{
			contractConfig: contractConfig,
			configDigest:   configDigest,
			changedInBlock: changedInBlock,
			blockHeight:    blockHeight,
		},
	}

	// ChainReaderImpl is a static implementation of the ChainReaderTester interface for testing
	ConfigProviderImpl = staticConfigProvider{
		staticConfigProviderConfig: staticConfigProviderConfig{
			offchainDigester:      OffchainConfigDigesterImpl,
			contractConfigTracker: ContractConfigTrackerImpl,
		},
	}

	// AgnosticPluginProviderImpl is a static implementation of the PluginProviderTester interface for testing
	AgnosticPluginProviderImpl = staticPluginProvider{}

	configDigest       = libocr.ConfigDigest([32]byte{1: 7, 13: 11, 31: 23})
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

	reportTimestamp = libocr.ReportTimestamp{
		ConfigDigest: configDigest,
		Epoch:        epoch,
		Round:        round,
	}
)

package test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	keystoretest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keystore/test"
	chaincomponentstest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader/test"
	cciptest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccip/test"
	ccipocr3test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccipocr3/test"
	mediantest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/median/test"
	mercurytest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/mercury/test"
	ocr3capabilitytest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ocr3capability/test"
	ocr2test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr2/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	looptypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/types"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var chainStatus = types.ChainStatus{
	ID:      "some_chain",
	Enabled: true,
}

var chainInfo = types.ChainInfo{
	FamilyName:      "someFamily",
	ChainID:         "123",
	NetworkName:     "someNetwork",
	NetworkNameFull: "someNetwork-test",
}

// testExtraDataCodecBundle is a dummy implementation of ExtraDataCodecBundle for testing
type testExtraDataCodecBundle struct{}

func newTestExtraDataCodecBundle() ccipocr3.ExtraDataCodecBundle {
	return testExtraDataCodecBundle{}
}

func (t testExtraDataCodecBundle) DecodeExtraArgs(extraArgs ccipocr3.Bytes, sourceChainSelector ccipocr3.ChainSelector) (map[string]any, error) {
	return map[string]any{
		"gasLimit": uint64(100000),
		"test":     "extraArgs",
	}, nil
}

func (t testExtraDataCodecBundle) DecodeTokenAmountDestExecData(destExecData ccipocr3.Bytes, sourceChainSelector ccipocr3.ChainSelector) (map[string]any, error) {
	return map[string]any{
		"data": "test-dest-exec-data",
		"test": "destExecData",
	}, nil
}

type transactionRequest struct {
	from         string
	to           string
	amount       *big.Int
	balanceCheck bool
}

type replayRequest struct {
	fromBlock string
	args      map[string]any
}

type nodeRequest struct {
	pageSize  int32
	pageToken string
}

type nodeResponse struct {
	nodes    []types.NodeStatus
	nextPage string
	total    int
}
type staticRelayerConfig struct {
	StaticChecks           bool
	relayArgs              types.RelayArgs
	pluginArgs             types.PluginArgs
	contractReaderConfig   []byte
	chainWriterConfig      []byte
	offRampAddress         ccipocr3.UnknownAddress
	pluginType             ccipocr3.PluginType
	transmitterAddress     ccipocr3.UnknownEncodedAddress
	extraDataCodecBundle   ccipocr3.ExtraDataCodecBundle
	medianProvider         testtypes.MedianProviderTester
	agnosticProvider       testtypes.PluginProviderTester
	mercuryProvider        mercurytest.MercuryProviderTester
	executionProvider      cciptest.ExecProviderTester
	commitProvider         cciptest.CommitProviderTester
	configProvider         ocr2test.ConfigProviderTester
	ocr3CapabilityProvider testtypes.OCR3CapabilityProviderTester
	contractReaderProvider testtypes.ContractReaderTester
	// Note: add other Provider testers here when we implement them
	// eg Functions, Automation, etc
	nodeRequest        nodeRequest
	nodeResponse       nodeResponse
	transactionRequest transactionRequest
	replayRequest      replayRequest
	chainStatus        types.ChainStatus
	chainInfo          types.ChainInfo
}

func newStaticRelayerConfig(lggr logger.Logger, staticChecks bool) staticRelayerConfig {
	return staticRelayerConfig{
		StaticChecks:           staticChecks,
		relayArgs:              RelayArgs,
		pluginArgs:             PluginArgs,
		contractReaderConfig:   []byte("test"),
		chainWriterConfig:      []byte("chainwriterconfig"),
		offRampAddress:         []byte("fakeAddress"),
		pluginType:             0,
		transmitterAddress:     "fakeAddress",
		extraDataCodecBundle:   newTestExtraDataCodecBundle(),
		medianProvider:         mediantest.MedianProvider(lggr),
		mercuryProvider:        mercurytest.MercuryProvider(lggr),
		executionProvider:      cciptest.ExecutionProvider(lggr),
		agnosticProvider:       ocr2test.AgnosticPluginProvider(lggr),
		configProvider:         ocr2test.ConfigProvider(lggr),
		ocr3CapabilityProvider: ocr3capabilitytest.OCR3CapabilityProvider(lggr),
		contractReaderProvider: chaincomponentstest.ContractReader,
		commitProvider:         cciptest.CommitProvider(lggr),
		nodeRequest: nodeRequest{
			pageSize:  137,
			pageToken: "",
		},
		nodeResponse: nodeResponse{
			nodes:    nodes,
			nextPage: "",
			total:    len(nodes),
		},
		transactionRequest: transactionRequest{
			from:         "me",
			to:           "you",
			amount:       big.NewInt(97),
			balanceCheck: true,
		},
		chainStatus: chainStatus,
		chainInfo:   chainInfo,
	}
}

func NewPluginRelayer(lggr logger.Logger, staticChecks bool) looptypes.PluginRelayer {
	return newStaticPluginRelayer(lggr, staticChecks)
}

func NewRelayerTester(lggr logger.Logger, staticChecks bool) testtypes.RelayerTester {
	return newStaticRelayer(lggr, staticChecks)
}

type staticPluginRelayer struct {
	services.Service
	relayer staticRelayer
}

func newStaticPluginRelayer(lggr logger.Logger, staticChecks bool) staticPluginRelayer {
	lggr = logger.Named(lggr, "staticPluginRelayer")
	return staticPluginRelayer{
		Service: test.NewStaticService(lggr),
		relayer: newStaticRelayer(lggr, staticChecks),
	}
}

func (s staticPluginRelayer) HealthReport() map[string]error {
	hp := s.Service.HealthReport()
	services.CopyHealth(hp, s.relayer.HealthReport())
	return hp
}

func (s staticPluginRelayer) NewRelayer(ctx context.Context, config string, keystore, csaKeystore core.Keystore, capabilityRegistry core.CapabilitiesRegistry) (looptypes.Relayer, error) {
	if s.relayer.StaticChecks && config != ConfigTOML {
		return nil, fmt.Errorf("expected config %q but got %q", ConfigTOML, config)
	}
	keys, err := keystore.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("expected at least one key but got none")
	}
	keys, err = csaKeystore.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("expected at least one CSA key but got none")
	}

	return s.relayer, nil
}

type staticRelayer struct {
	services.Service
	staticRelayerConfig
}

func newStaticRelayer(lggr logger.Logger, staticChecks bool) staticRelayer {
	lggr = logger.Named(lggr, "staticRelayer")
	cfg := newStaticRelayerConfig(lggr, staticChecks)
	return staticRelayer{
		Service:             test.NewStaticService(lggr),
		staticRelayerConfig: cfg,
	}
}

func (s staticRelayer) HealthReport() map[string]error {
	hp := s.Service.HealthReport()
	services.CopyHealth(hp, s.contractReaderProvider.HealthReport())
	services.CopyHealth(hp, s.configProvider.HealthReport())
	services.CopyHealth(hp, s.medianProvider.HealthReport())
	services.CopyHealth(hp, s.agnosticProvider.HealthReport())
	services.CopyHealth(hp, s.ocr3CapabilityProvider.HealthReport())
	services.CopyHealth(hp, s.mercuryProvider.HealthReport())
	services.CopyHealth(hp, s.executionProvider.HealthReport())
	services.CopyHealth(hp, s.commitProvider.HealthReport())
	return hp
}

func (s staticRelayer) NewContractWriter(_ context.Context, _ []byte) (types.ContractWriter, error) {
	return nil, errors.New("not implemented")
}

func (s staticRelayer) EVM() (types.EVMService, error) {
	return nil, nil
}

func (s staticRelayer) TON() (types.TONService, error) {
	return nil, nil
}

func (s staticRelayer) Solana() (types.SolanaService, error) {
	return nil, nil
}

func (s staticRelayer) Aptos() (types.AptosService, error) {
	return nil, nil
}

func (s staticRelayer) NewContractReader(_ context.Context, contractReaderConfig []byte) (types.ContractReader, error) {
	if s.StaticChecks && !(bytes.Equal(s.contractReaderConfig, contractReaderConfig)) {
		return nil, fmt.Errorf("expected contractReaderConfig:\n\t%v\nbut got:\n\t%v", string(s.contractReaderConfig), string(contractReaderConfig))
	}
	return s.contractReaderProvider, nil
}

func (s staticRelayer) NewConfigProvider(ctx context.Context, r types.RelayArgs) (types.ConfigProvider, error) {
	if s.StaticChecks && !equalRelayArgs(r, s.relayArgs) {
		return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", s.relayArgs, r)
	}
	return s.configProvider, nil
}

func (s staticRelayer) NewMedianProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.MedianProvider, error) {
	if s.StaticChecks {
		ra := newRelayArgsWithProviderType(types.Median)
		if !equalRelayArgs(r, ra) {
			return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", RelayArgs, r)
		}
		if !reflect.DeepEqual(PluginArgs, p) {
			return nil, fmt.Errorf("expected plugin args %v but got %v", PluginArgs, p)
		}
	}

	return s.medianProvider, nil
}

func (s staticRelayer) NewPluginProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.PluginProvider, error) {
	if s.StaticChecks {
		ra := newRelayArgsWithProviderType(types.Median)
		if !equalRelayArgs(r, ra) {
			return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", RelayArgs, r)
		}
		if !reflect.DeepEqual(PluginArgs, p) {
			return nil, fmt.Errorf("expected plugin args %v but got %v", PluginArgs, p)
		}
	}
	return s.agnosticProvider, nil
}

func (s staticRelayer) NewOCR3CapabilityProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.OCR3CapabilityProvider, error) {
	if s.StaticChecks {
		ra := newRelayArgsWithProviderType(types.OCR3Capability)
		if !equalRelayArgs(r, ra) {
			return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", RelayArgs, r)
		}
		if !reflect.DeepEqual(PluginArgs, p) {
			return nil, fmt.Errorf("expected plugin args %v but got %v", PluginArgs, p)
		}
	}
	return s.ocr3CapabilityProvider, nil
}

func (s staticRelayer) NewMercuryProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.MercuryProvider, error) {
	if s.StaticChecks {
		if !equalRelayArgs(r, mercurytest.RelayArgs) {
			return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", mercurytest.RelayArgs, r)
		}
		if !reflect.DeepEqual(mercurytest.PluginArgs, p) {
			return nil, fmt.Errorf("expected plugin args %v but got %v", mercurytest.PluginArgs, p)
		}
	}
	return s.mercuryProvider, nil
}

func (s staticRelayer) NewExecutionProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.CCIPExecProvider, error) {
	if s.StaticChecks {
		if !equalRelayArgs(r, cciptest.ExecutionRelayArgs) {
			return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", cciptest.ExecutionRelayArgs, r)
		}
		if !reflect.DeepEqual(cciptest.ExecutionPluginArgs, p) {
			return nil, fmt.Errorf("expected plugin args %v but got %v", cciptest.ExecutionPluginArgs, p)
		}
	}
	return s.executionProvider, nil
}

func (s staticRelayer) NewCommitProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.CCIPCommitProvider, error) {
	if s.StaticChecks {
		if !equalRelayArgs(r, cciptest.CommitRelayArgs) {
			return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", cciptest.CommitRelayArgs, r)
		}
		if !reflect.DeepEqual(cciptest.CommitPluginArgs, p) {
			return nil, fmt.Errorf("expected plugin args %v but got %v", cciptest.CommitPluginArgs, p)
		}
	}
	return s.commitProvider, nil
}

func (s staticRelayer) NewLLOProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.LLOProvider, error) {
	return nil, errors.New("not implemented")
}

func (s staticRelayer) NewCCIPProvider(ctx context.Context, r types.CCIPProviderArgs) (types.CCIPProvider, error) {
	ccipProviderArgs := types.CCIPProviderArgs{
		ExternalJobID:        s.relayArgs.ExternalJobID,
		ContractReaderConfig: s.contractReaderConfig,
		ChainWriterConfig:    s.chainWriterConfig,
		OffRampAddress:       s.offRampAddress,
		PluginType:           s.pluginType,
		TransmitterAddress:   s.transmitterAddress,
		ExtraDataCodecBundle: s.extraDataCodecBundle,
	}
	if s.StaticChecks && !equalCCIPProviderArgs(r, ccipProviderArgs) {
		return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", s.relayArgs, r)
	}
	return ccipocr3test.CCIPProvider(logger.Nop()), nil
}

func (s staticRelayer) LatestHead(ctx context.Context) (types.Head, error) {
	return types.Head{}, errors.New("not implemented")
}

func (s staticRelayer) GetChainStatus(ctx context.Context) (types.ChainStatus, error) {
	return s.chainStatus, nil
}

func (s staticRelayer) GetChainInfo(_ context.Context) (types.ChainInfo, error) {
	return s.chainInfo, nil
}

func (s staticRelayer) ListNodeStatuses(ctx context.Context, pageSize int32, pageToken string) ([]types.NodeStatus, string, int, error) {
	if s.StaticChecks && s.nodeRequest.pageSize != pageSize {
		return nil, "", -1, fmt.Errorf("expected page_size %d but got %d", s.nodeRequest.pageSize, pageSize)
	}
	if pageToken != "" {
		return nil, "", -1, fmt.Errorf("expected empty page_token but got %q", pageToken)
	}
	return s.nodeResponse.nodes, s.nodeResponse.nextPage, s.nodeResponse.total, nil
}

func (s staticRelayer) Transact(ctx context.Context, f, t string, a *big.Int, b bool) error {
	if s.StaticChecks {
		if f != s.transactionRequest.from {
			return fmt.Errorf("expected from %s but got %s", s.transactionRequest.from, f)
		}
		if t != s.transactionRequest.to {
			return fmt.Errorf("expected to %s but got %s", s.transactionRequest.to, t)
		}
		if s.transactionRequest.amount.Cmp(a) != 0 {
			return fmt.Errorf("expected amount %s but got %s", s.transactionRequest.amount, a)
		}
		if b != s.transactionRequest.balanceCheck { //nolint:gosimple
			return fmt.Errorf("expected balance check %t but got %t", s.transactionRequest.balanceCheck, b)
		}
	}

	return nil
}

func (s staticRelayer) Replay(ctx context.Context, fromBlock string, args map[string]any) error {
	if s.StaticChecks {
		if fromBlock != s.replayRequest.fromBlock {
			return fmt.Errorf("expected from %s but got %s", s.replayRequest.fromBlock, fromBlock)
		}
	}
	return nil
}

func (s staticRelayer) AssertEqual(_ context.Context, t *testing.T, relayer looptypes.Relayer) {
	t.Run("ContractReader", func(t *testing.T) {
		//t.Parallel()
		ctx := t.Context()
		contractReader, err := relayer.NewContractReader(ctx, []byte("test"))
		require.NoError(t, err)
		servicetest.Run(t, contractReader)
	})

	t.Run("ConfigProvider", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		configProvider, err := relayer.NewConfigProvider(ctx, RelayArgs)
		require.NoError(t, err)
		servicetest.Run(t, configProvider)

		s.configProvider.AssertEqual(ctx, t, configProvider)
	})

	t.Run("MedianProvider", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		ra := newRelayArgsWithProviderType(types.Median)
		p, err := relayer.NewPluginProvider(ctx, ra, PluginArgs)
		require.NoError(t, err)
		require.NotNil(t, p)
		provider := p.(types.MedianProvider)
		servicetest.Run(t, provider)

		t.Run("ReportingPluginProvider", func(t *testing.T) {
			t.Parallel()
			s.medianProvider.AssertEqual(ctx, t, provider)
		})
	})

	t.Run("PluginProvider", func(t *testing.T) {
		t.Parallel()
		ra := newRelayArgsWithProviderType(types.GenericPlugin)
		provider, err := relayer.NewPluginProvider(t.Context(), ra, PluginArgs)
		require.NoError(t, err)
		servicetest.Run(t, provider)
		t.Run("ReportingPluginProvider", func(t *testing.T) {
			t.Parallel()
			s.agnosticProvider.AssertEqual(t.Context(), t, provider)
		})
	})

	t.Run("GetChainStatus", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		gotChain, err := relayer.GetChainStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, s.chainStatus, gotChain)
	})

	t.Run("GetChainInfo", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		chainInfoReply, err := relayer.GetChainInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, s.chainInfo, chainInfoReply)
	})

	t.Run("ListNodeStatuses", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		gotNodes, gotNextToken, gotCount, err := relayer.ListNodeStatuses(ctx, s.nodeRequest.pageSize, s.nodeRequest.pageToken)
		require.NoError(t, err)
		assert.Equal(t, s.nodeResponse.nodes, gotNodes)
		assert.Equal(t, s.nodeResponse.total, gotCount)
		assert.Empty(t, s.nodeResponse.nextPage, gotNextToken)
	})

	t.Run("Transact", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		err := relayer.Transact(ctx, s.transactionRequest.from, s.transactionRequest.to, s.transactionRequest.amount, s.transactionRequest.balanceCheck)
		require.NoError(t, err)
	})

	t.Run("Replay", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		err := relayer.Replay(ctx, s.replayRequest.fromBlock, s.replayRequest.args)
		require.NoError(t, err)
	})
}

func equalRelayArgs(a, b types.RelayArgs) bool {
	return a.ExternalJobID == b.ExternalJobID &&
		a.JobID == b.JobID &&
		a.ContractID == b.ContractID &&
		a.New == b.New &&
		bytes.Equal(a.RelayConfig, b.RelayConfig)
}

func equalCCIPProviderArgs(a, b types.CCIPProviderArgs) bool {
	return a.ExternalJobID == b.ExternalJobID &&
		slices.Equal(a.ContractReaderConfig, b.ContractReaderConfig) &&
		slices.Equal(a.ChainWriterConfig, b.ChainWriterConfig) &&
		slices.Equal(a.OffRampAddress, b.OffRampAddress) &&
		a.PluginType == b.PluginType &&
		a.TransmitterAddress == b.TransmitterAddress &&
		a.ExtraDataCodecBundle == b.ExtraDataCodecBundle
}

func newRelayArgsWithProviderType(_type types.OCR2PluginType) types.RelayArgs {
	return types.RelayArgs{
		ExternalJobID: RelayArgs.ExternalJobID,
		JobID:         RelayArgs.JobID,
		ContractID:    RelayArgs.ContractID,
		New:           RelayArgs.New,
		RelayConfig:   RelayArgs.RelayConfig,
		ProviderType:  string(_type),
	}
}

func RunPlugin(t *testing.T, p looptypes.PluginRelayer) {
	t.Run("Relayer", func(t *testing.T) {
		ctx := t.Context()
		relayer, err := p.NewRelayer(ctx, ConfigTOML, keystoretest.Keystore, keystoretest.Keystore, nil)
		require.NoError(t, err)
		servicetest.Run(t, relayer)
		Run(t, relayer)
		servicetest.AssertHealthReportNames(t, relayer.HealthReport(),
			"PluginRelayerClient.RelayerClient",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticCommitProvider",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticConfigProvider",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticExecProvider",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticMedianProvider",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticMercuryProvider",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticPluginProvider",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticPluginProvider.staticPluginProvider",
		)
	})
}

func Run(t *testing.T, relayer looptypes.Relayer) {
	ctx := t.Context()
	expectedRelayer := NewRelayerTester(logger.Test(t), false)
	expectedRelayer.AssertEqual(ctx, t, relayer)
}

func RunFuzzPluginRelayer(f *testing.F, relayerFunc func(*testing.T) looptypes.PluginRelayer) {
	var (
		account = "testaccount"
		signed  = []byte{5: 11}
	)
	f.Add("ABC\xa8\x8c\xb3G\xfc", "", true, []byte{}, true, true, "")
	f.Add(ConfigTOML, account, false, signed, false, false, "")

	// nolint: gocognit
	f.Fuzz(func(
		t *testing.T, fConfig string, fAccts string, fAcctErr bool,
		fSigned []byte, fSignErr bool, fValsWErr bool, fErr string,
	) {
		keystore := fuzzerKeystore{
			accounts:      []string{fAccts}, // fuzzer does not support []string type
			acctErr:       fAcctErr,
			signed:        fSigned,
			signErr:       fSignErr,
			valuesWithErr: fValsWErr,
			errStr:        fErr,
		}

		ctx := t.Context()
		_, err := relayerFunc(t).NewRelayer(ctx, fConfig, keystore, keystore, nil)

		grpcUnavailableErr(t, err)
	})
}

func RunFuzzRelayer(f *testing.F, relayerFunc func(*testing.T) looptypes.Relayer) {
	validRaw := [16]byte(RelayArgs.ExternalJobID)
	validRawBytes := make([]byte, 16)

	copy(validRawBytes, validRaw[:])

	f.Add([]byte{}, int32(-1), "ABC\xa8\x8c\xb3G\xfc", false, []byte{}, "", "", []byte{})
	f.Add(validRawBytes, int32(123), "testcontract", true, []byte(ConfigTOML), string(types.Median), "testtransmitter", []byte{100: 88})

	f.Fuzz(func(
		t *testing.T, fExtJobID []byte, fJobID int32, fContractID string, fNew bool,
		fConfig []byte, fType string, fTransmID string, fPlugConf []byte,
	) {
		var rawBytes [16]byte

		copy(rawBytes[:], fExtJobID)

		relayer := relayerFunc(t)
		ctx := t.Context()
		fRelayArgs := types.RelayArgs{
			ExternalJobID: uuid.UUID(rawBytes),
			JobID:         fJobID,
			ContractID:    fContractID,
			New:           fNew,
			RelayConfig:   fConfig,
			ProviderType:  fType,
		}

		_, err := relayer.NewConfigProvider(ctx, fRelayArgs)

		grpcUnavailableErr(t, err)

		pArgs := types.PluginArgs{
			TransmitterID: fTransmID,
			PluginConfig:  fPlugConf,
		}

		provider, err := relayer.NewPluginProvider(ctx, fRelayArgs, pArgs)
		// require.NoError(t, provider.Start(ctx))
		t.Log("provider created")
		t.Cleanup(func() {
			t.Log("cleanup called")
			if provider != nil {
				assert.NoError(t, provider.Close())
			}
		})

		grpcUnavailableErr(t, err)
		t.Logf("error tested: %s", err)
	})
}

type FuzzableProvider[K any] func(context.Context, types.RelayArgs, types.PluginArgs) (K, error)

func RunFuzzProvider[K any](f *testing.F, providerFunc func(*testing.T) FuzzableProvider[K]) {
	validRaw := [16]byte(RelayArgs.ExternalJobID)
	validRawBytes := make([]byte, 16)

	copy(validRawBytes, validRaw[:])

	f.Add([]byte{}, int32(-1), "ABC\xa8\x8c\xb3G\xfc", false, []byte{}, "", "", []byte{})                                                    // bad inputs
	f.Add(validRawBytes, int32(123), "testcontract", true, []byte(ConfigTOML), string(types.Median), "testtransmitter", []byte{100: 88})     // valid for MedianProvider
	f.Add(validRawBytes, int32(123), "testcontract", true, []byte(ConfigTOML), string(types.Mercury), "testtransmitter", []byte{100: 88})    // valid for MercuryProvider
	f.Add(validRawBytes, int32(123), "testcontract", true, []byte(ConfigTOML), string(types.Functions), "testtransmitter", []byte{100: 88})  // valid for FunctionsProvider
	f.Add(validRawBytes, int32(123), "testcontract", true, []byte(ConfigTOML), string(types.OCR2Keeper), "testtransmitter", []byte{100: 88}) // valid for AutomationProvider

	f.Fuzz(func(
		t *testing.T, fExtJobID []byte, fJobID int32, fContractID string, fNew bool,
		fConfig []byte, fType string, fTransmID string, fPlugConf []byte,
	) {
		var rawBytes [16]byte

		copy(rawBytes[:], fExtJobID)

		provider := providerFunc(t)
		ctx := t.Context()
		fRelayArgs := types.RelayArgs{
			ExternalJobID: uuid.UUID(rawBytes),
			JobID:         fJobID,
			ContractID:    fContractID,
			New:           fNew,
			RelayConfig:   fConfig,
			ProviderType:  fType,
		}

		pArgs := types.PluginArgs{
			TransmitterID: fTransmID,
			PluginConfig:  fPlugConf,
		}

		_, err := provider(ctx, fRelayArgs, pArgs)

		grpcUnavailableErr(t, err)
	})
}

func grpcUnavailableErr(t *testing.T, err error) {
	t.Helper()

	if code := status.Code(err); code == codes.Unavailable {
		t.FailNow()
	}
}

type fuzzerKeystore struct {
	accounts      []string
	acctErr       bool
	signed        []byte
	signErr       bool
	decrypted     []byte
	decryptErr    bool
	valuesWithErr bool
	errStr        string
}

func (k fuzzerKeystore) Accounts(ctx context.Context) ([]string, error) {
	if k.acctErr {
		err := errors.New(k.errStr)

		if k.valuesWithErr {
			return k.accounts, err
		}

		return nil, err
	}

	return k.accounts, nil
}

// Sign returns data signed by account.
// nil data can be used as a no-op to check for account existence.
func (k fuzzerKeystore) Sign(ctx context.Context, account string, data []byte) ([]byte, error) {
	if k.signErr {
		err := errors.New(k.errStr)

		if k.valuesWithErr {
			return k.signed, err
		}

		return nil, err
	}

	return k.signed, nil
}

func (k fuzzerKeystore) Decrypt(ctx context.Context, account string, encrypted []byte) ([]byte, error) {
	if k.decryptErr {
		err := errors.New(k.errStr)

		if k.valuesWithErr {
			return k.decrypted, err
		}

		return nil, err
	}

	if len(k.decrypted) == 0 {
		return nil, fmt.Errorf("no decrypted data for account %s", account)
	}

	return k.decrypted, nil
}

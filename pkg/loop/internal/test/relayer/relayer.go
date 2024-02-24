package relayer_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	mercury_common_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/common/test"
	median_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/median"
	pluginprovider_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/plugin_provider"
	keystore_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

var chainStatus = types.ChainStatus{
	ID:      "some_chain",
	Enabled: true,
}

type transactionRequest struct {
	from         string
	to           string
	amount       *big.Int
	balanceCheck bool
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
type StaticPluginRelayerConfig struct {
	StaticChecks bool
	RelayArgs    types.RelayArgs
	PluginArgs   types.PluginArgs
	nodeRequest
	nodeResponse
	transactionRequest
	chainStatus types.ChainStatus
}

type RelayerTester interface {
	loop.PluginRelayer
	loop.Relayer
	mustEmbed()
}

func NewRelayerTester(staticChecks bool) RelayerTester {
	return staticPluginRelayer{
		StaticPluginRelayerConfig: StaticPluginRelayerConfig{
			StaticChecks: staticChecks,
			RelayArgs:    relayArgs,
			PluginArgs:   pluginArgs,
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
		},
	}
}

var _ RelayerTester = staticPluginRelayer{}

type staticPluginRelayer struct {
	StaticPluginRelayerConfig
}

func (s staticPluginRelayer) mustEmbed() {}

func (s staticPluginRelayer) NewRelayer(ctx context.Context, config string, keystore types.Keystore) (internal.Relayer, error) {
	if s.StaticChecks && config != ConfigTOML {
		return nil, fmt.Errorf("expected config %q but got %q", ConfigTOML, config)
	}
	keys, err := keystore.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("expected at least one key but got none")
	}
	/*
		if s.StaticChecks && !reflect.DeepEqual([]string{string(account)}, keys) {
			return nil, fmt.Errorf("expected keys %v but got %v", []string{string(account)}, keys)
		}
		gotSigned, err := keystore.Sign(ctx, string(account), encoded)
		if err != nil {
			return nil, err
		}
		if s.StaticChecks && !bytes.Equal(signed, gotSigned) {
			return nil, fmt.Errorf("expected signed bytes %x but got %x", signed, gotSigned)
		}
	*/
	return s, nil
}

func (s staticPluginRelayer) Start(ctx context.Context) error {

	/*	// lazy initialization
		s.StaticPluginRelayerConfig.RelayArgs = relayArgs
		s.StaticPluginRelayerConfig.PluginArgs = pluginArgs
		s.nodeRequest = nodeRequest{
			pageSize:  137,
			pageToken: "",
		}
		s.nodeResponse = nodeResponse{
			nodes:    nodes,
			nextPage: "",
			total:    len(nodes),
		}

		//s.nodeStatus = nodes
		//s.limit = 137
		s.transactionRequest = transactionRequest{
			from:         "me",
			to:           "you",
			amount:       big.NewInt(97),
			balanceCheck: true,
		}
	*/
	return nil
}

func (s staticPluginRelayer) Close() error { return nil }

func (s staticPluginRelayer) Ready() error { panic("unimplemented") }

func (s staticPluginRelayer) Name() string { panic("unimplemented") }

func (s staticPluginRelayer) HealthReport() map[string]error { panic("unimplemented") }

func (s staticPluginRelayer) NewConfigProvider(ctx context.Context, r types.RelayArgs) (types.ConfigProvider, error) {
	if s.StaticChecks && !equalRelayArgs(r, s.RelayArgs) {
		return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", s.RelayArgs, r)
	}
	return pluginprovider_test.ConfigProviderImpl, nil
}

func (s staticPluginRelayer) NewMedianProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.MedianProvider, error) {

	if s.StaticChecks {
		ra := newRelayArgsWithProviderType(types.Median)
		if !equalRelayArgs(r, ra) {
			return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", relayArgs, r)
		}
		if !reflect.DeepEqual(pluginArgs, p) {
			return nil, fmt.Errorf("expected plugin args %v but got %v", pluginArgs, p)
		}
	}

	return median_test.MedianProviderImpl, nil
}

func (s staticPluginRelayer) NewPluginProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.PluginProvider, error) {
	if s.StaticChecks {
		ra := newRelayArgsWithProviderType(types.Median)
		if !equalRelayArgs(r, ra) {
			return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", relayArgs, r)
		}
		if !reflect.DeepEqual(pluginArgs, p) {
			return nil, fmt.Errorf("expected plugin args %v but got %v", pluginArgs, p)
		}
	}

	return pluginprovider_test.TestPluginProvider, nil
}

func (s staticPluginRelayer) NewMercuryProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.MercuryProvider, error) {
	if s.StaticChecks {
		if !equalRelayArgs(r, mercury_common_test.RelayArgs) {
			return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", mercury_common_test.RelayArgs, r)
		}
		if !reflect.DeepEqual(mercury_common_test.PluginArgs, p) {
			return nil, fmt.Errorf("expected plugin args %v but got %v", mercury_common_test.PluginArgs, p)
		}
	}
	return nil, nil
	//return StaticMercuryProvider{}, nil
}

func (s staticPluginRelayer) NewLLOProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.LLOProvider, error) {
	return nil, errors.New("not implemented")
}

func (s staticPluginRelayer) GetChainStatus(ctx context.Context) (types.ChainStatus, error) {
	return s.chainStatus, nil
}

func (s staticPluginRelayer) ListNodeStatuses(ctx context.Context, pageSize int32, pageToken string) ([]types.NodeStatus, string, int, error) {
	if s.StaticChecks && s.nodeRequest.pageSize != pageSize {
		return nil, "", -1, fmt.Errorf("expected page_size %d but got %d", s.nodeRequest.pageSize, pageSize)
	}
	if pageToken != "" {
		return nil, "", -1, fmt.Errorf("expected empty page_token but got %q", pageToken)
	}
	return s.nodeResponse.nodes, s.nodeResponse.nextPage, s.nodeResponse.total, nil
}

func (s staticPluginRelayer) Transact(ctx context.Context, f, t string, a *big.Int, b bool) error {
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

func equalRelayArgs(a, b types.RelayArgs) bool {
	return a.ExternalJobID == b.ExternalJobID &&
		a.JobID == b.JobID &&
		a.ContractID == b.ContractID &&
		a.New == b.New &&
		bytes.Equal(a.RelayConfig, b.RelayConfig)
}

func newRelayArgsWithProviderType(_type types.OCR2PluginType) types.RelayArgs {
	return types.RelayArgs{
		ExternalJobID: relayArgs.ExternalJobID,
		JobID:         relayArgs.JobID,
		ContractID:    relayArgs.ContractID,
		New:           relayArgs.New,
		RelayConfig:   relayArgs.RelayConfig,
		ProviderType:  string(_type),
	}
}

func RunPluginRelayer(t *testing.T, p internal.PluginRelayer) {
	ctx := tests.Context(t)

	t.Run("Relayer", func(t *testing.T) {
		relayer, err := p.NewRelayer(ctx, ConfigTOML, keystore_test.StaticKeystore{})
		require.NoError(t, err)
		require.NoError(t, relayer.Start(ctx))
		t.Cleanup(func() { assert.NoError(t, relayer.Close()) })
		RunRelayer(t, relayer)
	})
}

func RunRelayer(t *testing.T, relayer internal.Relayer) {
	ctx := tests.Context(t)
	var (
		expectedConfigProvider   = pluginprovider_test.ConfigProviderImpl
		expectedMedianProvider   = median_test.MedianProviderImpl
		expectedAgnosticProvider = pluginprovider_test.ConfigProviderImpl
		expectedRelayer          = staticPluginRelayer{}
	)
	// TODO: fix lazy init?
	assert.NoError(t, expectedRelayer.Start(ctx))
	t.Run("ConfigProvider", func(t *testing.T) {
		t.Parallel()
		configProvider, err := relayer.NewConfigProvider(ctx, relayArgs)
		require.NoError(t, err)
		require.NoError(t, configProvider.Start(ctx))
		t.Cleanup(func() { assert.NoError(t, configProvider.Close()) })

		expectedConfigProvider.AssertEqual(t, ctx, configProvider)
	})

	t.Run("MedianProvider", func(t *testing.T) {
		t.Parallel()
		ra := newRelayArgsWithProviderType(types.Median)
		p, err := relayer.NewPluginProvider(ctx, ra, pluginArgs)
		provider := p.(types.MedianProvider)
		require.NoError(t, err)
		require.NoError(t, provider.Start(ctx))
		t.Cleanup(func() { assert.NoError(t, provider.Close()) })

		t.Run("ReportingPluginProvider", func(t *testing.T) {
			t.Parallel()

			expectedMedianProvider.AssertEqual(t, ctx, provider)
		})
	})

	t.Run("PluginProvider", func(t *testing.T) {
		t.Parallel()
		ra := newRelayArgsWithProviderType(types.GenericPlugin)
		provider, err := relayer.NewPluginProvider(ctx, ra, pluginArgs)
		require.NoError(t, err)
		require.NoError(t, provider.Start(ctx))
		t.Cleanup(func() { assert.NoError(t, provider.Close()) })

		t.Run("ReportingPluginProvider", func(t *testing.T) {
			t.Parallel()

			expectedAgnosticProvider.AssertEqual(t, ctx, provider)
		})
	})

	// TODO add this back

	t.Fatalf("don't forget to add this back")
	/*
		t.Run("GetChainStatus", func(t *testing.T) {
			t.Parallel()
			gotChain, err := relayer.GetChainStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, chain, gotChain)
		})

		t.Run("ListNodeStatuses", func(t *testing.T) {
			t.Parallel()
			gotNodes, gotNextToken, gotCount, err := relayer.ListNodeStatuses(ctx, limit, "")
			require.NoError(t, err)
			assert.Equal(t, nodes, gotNodes)
			assert.Equal(t, total, gotCount)
			assert.Empty(t, gotNextToken)
		})

		t.Run("Transact", func(t *testing.T) {
			t.Parallel()
			err := relayer.Transact(ctx, from, to, amount, balanceCheck)
			require.NoError(t, err)
		})
	*/
}

func RunFuzzPluginRelayer(f *testing.F, relayerFunc func(*testing.T) internal.PluginRelayer) {
	var (
		account = "testaccount"
		signed  = []byte{5: 11}
	)
	f.Add("ABC\xa8\x8c\xb3G\xfc", "", true, []byte{}, true, true, "")
	f.Add(ConfigTOML, string(account), false, signed, false, false, "")

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

		ctx := tests.Context(t)
		_, err := relayerFunc(t).NewRelayer(ctx, fConfig, keystore)

		grpcUnavailableErr(t, err)
	})
}

func RunFuzzRelayer(f *testing.F, relayerFunc func(*testing.T) internal.Relayer) {
	validRaw := [16]byte(relayArgs.ExternalJobID)
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
		ctx := tests.Context(t)
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
	validRaw := [16]byte(relayArgs.ExternalJobID)
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
		ctx := tests.Context(t)
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
	valuesWithErr bool
	errStr        string
}

func (k fuzzerKeystore) Accounts(ctx context.Context) ([]string, error) {
	if k.acctErr {
		err := fmt.Errorf(k.errStr)

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
		err := fmt.Errorf(k.errStr)

		if k.valuesWithErr {
			return k.signed, err
		}

		return nil, err
	}

	return k.signed, nil
}

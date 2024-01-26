package test

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

type StaticKeystore struct{}

func (s StaticKeystore) Accounts(ctx context.Context) (accounts []string, err error) {
	return []string{string(account)}, nil
}

func (s StaticKeystore) Sign(ctx context.Context, id string, data []byte) ([]byte, error) {
	if string(account) != id {
		return nil, fmt.Errorf("expected id %q but got %q", account, id)
	}
	if !bytes.Equal(encoded, data) {
		return nil, fmt.Errorf("expected encoded data %x but got %x", encoded, data)
	}
	return signed, nil
}

type staticPluginRelayer struct {
	staticService
}

func NewStaticPluginRelayer(lggr logger.Logger) staticPluginRelayer {
	return staticPluginRelayer{staticService{lggr: logger.Named(lggr, "staticPluginRelayer")}}
}

func (s staticPluginRelayer) HealthReport() map[string]error {
	hp := s.staticService.HealthReport()
	services.CopyHealth(hp, newStaticRelayer(s.lggr).HealthReport())
	return hp
}

func (s staticPluginRelayer) NewRelayer(ctx context.Context, config string, keystore types.Keystore) (internal.Relayer, error) {
	if config != ConfigTOML {
		return nil, fmt.Errorf("expected config %q but got %q", ConfigTOML, config)
	}
	keys, err := keystore.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual([]string{string(account)}, keys) {
		return nil, fmt.Errorf("expected keys %v but got %v", []string{string(account)}, keys)
	}
	gotSigned, err := keystore.Sign(ctx, string(account), encoded)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(signed, gotSigned) {
		return nil, fmt.Errorf("expected signed bytes %x but got %x", signed, gotSigned)
	}
	return newStaticRelayer(s.lggr), nil
}

type staticRelayer struct {
	staticService
}

func newStaticRelayer(lggr logger.Logger) staticRelayer {
	return staticRelayer{staticService{lggr: logger.Named(lggr, "staticRelayer")}}
}

func (s staticRelayer) HealthReport() map[string]error {
	hp := s.staticService.HealthReport()
	services.CopyHealth(hp, newStaticConfigProvider(s.lggr).HealthReport())
	services.CopyHealth(hp, NewStaticMedianProvider(s.lggr).HealthReport())
	services.CopyHealth(hp, NewStaticPluginProvider(s.lggr).HealthReport())
	return hp
}

func (s staticRelayer) NewConfigProvider(ctx context.Context, r types.RelayArgs) (types.ConfigProvider, error) {
	if !equalRelayArgs(r, RelayArgs) {
		return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", RelayArgs, r)
	}
	return newStaticConfigProvider(s.lggr), nil
}

func (s staticRelayer) NewMedianProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.MedianProvider, error) {
	ra := newRelayArgsWithProviderType(types.Median)
	if !equalRelayArgs(r, ra) {
		return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", RelayArgs, r)
	}
	if !reflect.DeepEqual(PluginArgs, p) {
		return nil, fmt.Errorf("expected plugin args %v but got %v", PluginArgs, p)
	}
	return NewStaticMedianProvider(s.lggr), nil
}

func (s staticRelayer) NewPluginProvider(ctx context.Context, r types.RelayArgs, p types.PluginArgs) (types.PluginProvider, error) {
	ra := newRelayArgsWithProviderType(types.Median)
	if !equalRelayArgs(r, ra) {
		return nil, fmt.Errorf("expected relay args:\n\t%v\nbut got:\n\t%v", RelayArgs, r)
	}
	if !reflect.DeepEqual(PluginArgs, p) {
		return nil, fmt.Errorf("expected plugin args %v but got %v", PluginArgs, p)
	}
	return NewStaticPluginProvider(s.lggr), nil
}

func (s staticRelayer) GetChainStatus(ctx context.Context) (types.ChainStatus, error) {
	return chain, nil
}

func (s staticRelayer) ListNodeStatuses(ctx context.Context, pageSize int32, pageToken string) ([]types.NodeStatus, string, int, error) {
	if limit != pageSize {
		return nil, "", -1, fmt.Errorf("expected page_size %d but got %d", limit, pageSize)
	}
	if pageToken != "" {
		return nil, "", -1, fmt.Errorf("expected empty page_token but got %q", pageToken)
	}
	return nodes, "", total, nil
}

func (s staticRelayer) Transact(ctx context.Context, f, t string, a *big.Int, b bool) error {
	if f != from {
		return fmt.Errorf("expected from %s but got %s", from, f)
	}
	if t != to {
		return fmt.Errorf("expected to %s but got %s", to, t)
	}
	if amount.Cmp(a) != 0 {
		return fmt.Errorf("expected amount %s but got %s", amount, a)
	}
	if b != balanceCheck { //nolint:gosimple
		return fmt.Errorf("expected balance check %t but got %t", balanceCheck, b)
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
		ExternalJobID: RelayArgs.ExternalJobID,
		JobID:         RelayArgs.JobID,
		ContractID:    RelayArgs.ContractID,
		New:           RelayArgs.New,
		RelayConfig:   RelayArgs.RelayConfig,
		ProviderType:  string(_type),
	}
}

func RunPluginRelayer(t *testing.T, p internal.PluginRelayer) {
	ctx := tests.Context(t)
	servicetest.Run(t, p)

	t.Run("Relayer", func(t *testing.T) {
		relayer, err := p.NewRelayer(ctx, ConfigTOML, StaticKeystore{})
		require.NoError(t, err)
		servicetest.Run(t, relayer)
		RunRelayer(t, relayer)
		servicetest.AssertHealthReportNames(t, relayer.HealthReport(),
			"PluginRelayerClient", //TODO missing
			"PluginRelayerClient.RelayerClient",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer", //TODO missing
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticMedianProvider",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticPluginProvider",
			"PluginRelayerClient.RelayerClient.staticPluginRelayer.staticRelayer.staticConfigProvider",
		)
	})
}

func RunRelayer(t *testing.T, relayer internal.Relayer) {
	ctx := tests.Context(t)

	configProvider, err0 := relayer.NewConfigProvider(ctx, RelayArgs)
	require.NoError(t, err0)
	servicetest.Run(t, configProvider)

	t.Run("ConfigProvider", func(t *testing.T) {
		t.Parallel()

		t.Run("OffchainConfigDigester", func(t *testing.T) {
			t.Parallel()
			ocd := configProvider.OffchainConfigDigester()
			gotConfigDigestPrefix, err := ocd.ConfigDigestPrefix()
			require.NoError(t, err)
			assert.Equal(t, configDigestPrefix, gotConfigDigestPrefix)
			gotConfigDigest, err := ocd.ConfigDigest(contractConfig)
			require.NoError(t, err)
			assert.Equal(t, configDigest, gotConfigDigest)
		})
		t.Run("ContractConfigTracker", func(t *testing.T) {
			t.Parallel()
			cct := configProvider.ContractConfigTracker()
			gotBlockHeight, err := cct.LatestBlockHeight(ctx)
			require.NoError(t, err)
			assert.Equal(t, blockHeight, gotBlockHeight)
			gotChangedInBlock, gotConfigDigest, err := cct.LatestConfigDetails(ctx)
			require.NoError(t, err)
			assert.Equal(t, changedInBlock, gotChangedInBlock)
			assert.Equal(t, configDigest, gotConfigDigest)
			gotContractConfig, err := cct.LatestConfig(ctx, changedInBlock)
			require.NoError(t, err)
			assert.Equal(t, contractConfig, gotContractConfig)
		})
	})

	pp, err0 := relayer.NewPluginProvider(ctx, newRelayArgsWithProviderType(types.Median), PluginArgs)
	require.NoError(t, err0)
	provider := pp.(types.MedianProvider)
	servicetest.Run(t, provider)

	t.Run("MedianProvider", func(t *testing.T) {
		t.Parallel()

		t.Run("OffchainConfigDigester", func(t *testing.T) {
			t.Parallel()
			ocd := provider.OffchainConfigDigester()
			gotConfigDigestPrefix, err := ocd.ConfigDigestPrefix()
			require.NoError(t, err)
			assert.Equal(t, configDigestPrefix, gotConfigDigestPrefix)
			gotConfigDigest, err := ocd.ConfigDigest(contractConfig)
			require.NoError(t, err)
			assert.Equal(t, configDigest, gotConfigDigest)
		})
		t.Run("ContractConfigTracker", func(t *testing.T) {
			t.Parallel()
			cct := provider.ContractConfigTracker()
			gotBlockHeight, err := cct.LatestBlockHeight(ctx)
			require.NoError(t, err)
			assert.Equal(t, blockHeight, gotBlockHeight)
			gotChangedInBlock, gotConfigDigest, err := cct.LatestConfigDetails(ctx)
			require.NoError(t, err)
			assert.Equal(t, changedInBlock, gotChangedInBlock)
			assert.Equal(t, configDigest, gotConfigDigest)
			gotContractConfig, err := cct.LatestConfig(ctx, changedInBlock)
			require.NoError(t, err)
			assert.Equal(t, contractConfig, gotContractConfig)
		})
		t.Run("ContractTransmitter", func(t *testing.T) {
			t.Parallel()
			ct := provider.ContractTransmitter()
			gotAccount, err := ct.FromAccount()
			require.NoError(t, err)
			assert.Equal(t, account, gotAccount)
			gotConfigDigest, gotEpoch, err := ct.LatestConfigDigestAndEpoch(ctx)
			require.NoError(t, err)
			assert.Equal(t, configDigest, gotConfigDigest)
			assert.Equal(t, epoch, gotEpoch)
			err = ct.Transmit(ctx, reportContext, report, sigs)
			require.NoError(t, err)
		})
		t.Run("ReportCodec", func(t *testing.T) {
			t.Parallel()
			rc := provider.ReportCodec()
			gotReport, err := rc.BuildReport(pobs)
			require.NoError(t, err)
			assert.Equal(t, report, gotReport)
			gotMedianValue, err := rc.MedianFromReport(report)
			require.NoError(t, err)
			assert.Equal(t, medianValue, gotMedianValue)
			gotMax, err := rc.MaxReportLength(n)
			require.NoError(t, err)
			assert.Equal(t, max, gotMax)
		})
		t.Run("MedianContract", func(t *testing.T) {
			t.Parallel()
			mc := provider.MedianContract()
			gotConfigDigest, gotEpoch, gotRound, err := mc.LatestRoundRequested(ctx, lookbackDuration)
			require.NoError(t, err)
			assert.Equal(t, configDigest, gotConfigDigest)
			assert.Equal(t, epoch, gotEpoch)
			assert.Equal(t, round, gotRound)
			gotConfigDigest, gotEpoch, gotRound, gotLatestAnswer, gotLatestTimestamp, err := mc.LatestTransmissionDetails(ctx)
			require.NoError(t, err)
			assert.Equal(t, configDigest, gotConfigDigest)
			assert.Equal(t, epoch, gotEpoch)
			assert.Equal(t, round, gotRound)
			assert.Equal(t, latestAnswer, gotLatestAnswer)
			assert.WithinDuration(t, latestTimestamp, gotLatestTimestamp, time.Second)
		})
		t.Run("OnchainConfigCodec", func(t *testing.T) {
			t.Parallel()
			occ := provider.OnchainConfigCodec()
			gotEncoded, err := occ.Encode(onchainConfig)
			require.NoError(t, err)
			assert.Equal(t, encoded, gotEncoded)
			gotDecoded, err := occ.Decode(encoded)
			require.NoError(t, err)
			assert.Equal(t, onchainConfig, gotDecoded)
		})
	})

	pp, err0 = relayer.NewPluginProvider(ctx, newRelayArgsWithProviderType(types.GenericPlugin), PluginArgs)
	require.NoError(t, err0)
	servicetest.Run(t, pp)

	t.Run("PluginProvider", func(t *testing.T) {
		t.Parallel()

		t.Run("OffchainConfigDigester", func(t *testing.T) {
			t.Parallel()
			ocd := pp.OffchainConfigDigester()
			gotConfigDigestPrefix, err := ocd.ConfigDigestPrefix()
			require.NoError(t, err)
			assert.Equal(t, configDigestPrefix, gotConfigDigestPrefix)
			gotConfigDigest, err := ocd.ConfigDigest(contractConfig)
			require.NoError(t, err)
			assert.Equal(t, configDigest, gotConfigDigest)
		})
		t.Run("ContractConfigTracker", func(t *testing.T) {
			t.Parallel()
			cct := pp.ContractConfigTracker()
			gotBlockHeight, err := cct.LatestBlockHeight(ctx)
			require.NoError(t, err)
			assert.Equal(t, blockHeight, gotBlockHeight)
			gotChangedInBlock, gotConfigDigest, err := cct.LatestConfigDetails(ctx)
			require.NoError(t, err)
			assert.Equal(t, changedInBlock, gotChangedInBlock)
			assert.Equal(t, configDigest, gotConfigDigest)
			gotContractConfig, err := cct.LatestConfig(ctx, changedInBlock)
			require.NoError(t, err)
			assert.Equal(t, contractConfig, gotContractConfig)
		})
		t.Run("ContractTransmitter", func(t *testing.T) {
			t.Parallel()
			ct := pp.ContractTransmitter()
			gotAccount, err := ct.FromAccount()
			require.NoError(t, err)
			assert.Equal(t, account, gotAccount)
			gotConfigDigest, gotEpoch, err := ct.LatestConfigDigestAndEpoch(ctx)
			require.NoError(t, err)
			assert.Equal(t, configDigest, gotConfigDigest)
			assert.Equal(t, epoch, gotEpoch)
			err = ct.Transmit(ctx, reportContext, report, sigs)
			require.NoError(t, err)
		})
	})

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
}

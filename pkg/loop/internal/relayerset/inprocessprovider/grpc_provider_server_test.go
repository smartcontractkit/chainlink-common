package inprocessprovider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestProviderServer(t *testing.T) {
	r := &mockRelayer{}
	mp, _ := r.NewPluginProvider(tests.Context(t), types.RelayArgs{ProviderType: string(types.Median)}, types.PluginArgs{})

	lggr := logger.Test(t)
	_, err := NewProviderServer(mp, "unsupported-type", lggr)
	require.ErrorContains(t, err, "unsupported-type")

	ps, err := NewProviderServer(staticMedianProvider{}, types.Median, lggr)
	require.NoError(t, err)

	_, err = ps.GetConn()
	require.NoError(t, err)
}

type mockRelayer struct {
	types.Relayer
}

func (m *mockRelayer) NewMedianProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.MedianProvider, error) {
	return staticMedianProvider{}, nil
}

func (m *mockRelayer) NewFunctionsProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.FunctionsProvider, error) {
	return staticFunctionsProvider{}, nil
}

func (m *mockRelayer) NewMercuryProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.MercuryProvider, error) {
	return staticMercuryProvider{}, nil
}

func (m *mockRelayer) NewAutomationProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.AutomationProvider, error) {
	return staticAutomationProvider{}, nil
}

func (m *mockRelayer) NewPluginProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.PluginProvider, error) {
	return staticPluginProvider{}, nil
}

type staticMedianProvider struct {
}

var _ types.MedianProvider = staticMedianProvider{}

// ContractConfigTracker implements types.MedianProvider.
func (s staticMedianProvider) ContractConfigTracker() ocr2types.ContractConfigTracker {
	return nil
}

// ContractTransmitter implements types.MedianProvider.
func (s staticMedianProvider) ContractTransmitter() ocr2types.ContractTransmitter {
	return nil
}

// MedianContract implements types.MedianProvider.
func (s staticMedianProvider) MedianContract() median.MedianContract {
	return nil
}

// OffchainConfigDigester implements types.MedianProvider.
func (s staticMedianProvider) OffchainConfigDigester() ocr2types.OffchainConfigDigester {
	return nil
}

// OnchainConfigCodec implements types.MedianProvider.
func (s staticMedianProvider) OnchainConfigCodec() median.OnchainConfigCodec {
	return nil
}

// ReportCodec implements types.MedianProvider.
func (s staticMedianProvider) ReportCodec() median.ReportCodec {
	return nil
}

// ContractReader implements types.MedianProvider.
func (s staticMedianProvider) ContractReader() types.ContractReader {
	return nil
}

// Close implements types.MedianProvider.
func (s staticMedianProvider) Close() error {
	return nil
}

// Codec implements types.MedianProvider.
func (s staticMedianProvider) Codec() types.Codec {
	return nil
}

// HealthReport implements types.MedianProvider.
func (s staticMedianProvider) HealthReport() map[string]error {
	return nil
}

// Name implements types.MedianProvider.
func (s staticMedianProvider) Name() string {
	return ""
}

// Ready implements types.MedianProvider.
func (s staticMedianProvider) Ready() error {
	return nil
}

// Start implements types.MedianProvider.
func (s staticMedianProvider) Start(context.Context) error {
	return nil
}

type staticFunctionsProvider struct {
	types.FunctionsProvider
}

type staticMercuryProvider struct {
	types.MercuryProvider
}

type staticAutomationProvider struct {
	types.AutomationProvider
}

type staticPluginProvider struct {
	types.PluginProvider
}

package test

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	mercury_common_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/common/test"
	mercury_v1_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/v1/test"
	mercury_v2_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/v2/test"
	mercury_v3_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/v3/test"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"

	mercury_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury"
	mercury_v1_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v1"
	mercury_v2_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v2"
	mercury_v3_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v3"
)

func PluginMercuryFunc(t *testing.T, p types.PluginMercury) {
	PluginMercuryTest{&StaticMercuryProvider{}}.TestPluginMercury(t, p)
}

type PluginMercuryTest struct {
	types.MercuryProvider
}

func (m PluginMercuryTest) TestPluginMercury(t *testing.T, p types.PluginMercury) {
	t.Run("PluginMercury", func(t *testing.T) {
		ctx := tests.Context(t)
		factory, err := p.NewMercuryV3Factory(ctx, m.MercuryProvider, mercury_v3_test.StaticDataSource{})
		require.NoError(t, err)

		ReportingPluginFactory(t, factory)
	})
}

type StaticPluginMercury struct{}

func (s StaticPluginMercury) NewMercuryV3Factory(ctx context.Context, provider types.MercuryProvider, dataSource mercury_v3_types.DataSource) (types.ReportingPluginFactory, error) {

	ocd := provider.OffchainConfigDigester()
	gotDigestPrefix, err := ocd.ConfigDigestPrefix()
	if err != nil {
		return nil, fmt.Errorf("failed to get ConfigDigestPrefix: %w", err)
	}
	if gotDigestPrefix != configDigestPrefix {
		return nil, fmt.Errorf("expected ConfigDigestPrefix %x but got %x", configDigestPrefix, gotDigestPrefix)
	}
	gotDigest, err := ocd.ConfigDigest(contractConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get ConfigDigest: %w", err)
	}
	if gotDigest != configDigest {
		return nil, fmt.Errorf("expected ConfigDigest %x but got %x", configDigest, gotDigest)
	}
	cct := provider.ContractConfigTracker()
	gotBlockHeight, err := cct.LatestBlockHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LatestBlockHeight: %w", err)
	}
	if gotBlockHeight != blockHeight {
		return nil, fmt.Errorf("expected LatestBlockHeight %d but got %d", blockHeight, gotBlockHeight)
	}
	gotChangedInBlock, gotConfigDigest, err := cct.LatestConfigDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LatestConfigDetails: %w", err)
	}
	if gotChangedInBlock != changedInBlock {
		return nil, fmt.Errorf("expected changedInBlock %d but got %d", changedInBlock, gotChangedInBlock)
	}
	if gotConfigDigest != configDigest {
		return nil, fmt.Errorf("expected ConfigDigest %s but got %s", configDigest, gotConfigDigest)
	}
	gotContractConfig, err := cct.LatestConfig(ctx, changedInBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get LatestConfig: %w", err)
	}
	if !reflect.DeepEqual(gotContractConfig, contractConfig) {
		return nil, fmt.Errorf("expected ContractConfig %v but got %v", contractConfig, gotContractConfig)
	}
	ct := provider.ContractTransmitter()
	gotAccount, err := ct.FromAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to get FromAccount: %w", err)
	}
	if gotAccount != account {
		return nil, fmt.Errorf("expectd FromAccount %s but got %s", account, gotAccount)
	}
	gotConfigDigest, gotEpoch, err := ct.LatestConfigDigestAndEpoch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LatestConfigDigestAndEpoch: %w", err)
	}
	if gotConfigDigest != configDigest {
		return nil, fmt.Errorf("expected ConfigDigest %s but got %s", configDigest, gotConfigDigest)
	}
	if gotEpoch != epoch {
		return nil, fmt.Errorf("expected Epoch %d but got %d", epoch, gotEpoch)
	}
	err = ct.Transmit(ctx, reportContext, report, sigs)
	if err != nil {
		return nil, fmt.Errorf("failed to Transmit")
	}

	rc := provider.ReportCodecV3()
	gotReport, err := rc.BuildReport(mercury_v3_test.Fixtures.ReportFields)
	if err != nil {
		return nil, fmt.Errorf("failed to BuildReport: %w", err)
	}
	if !bytes.Equal(gotReport, mercury_v3_test.Fixtures.Report) {
		return nil, fmt.Errorf("expected Report %x but got %x", report, gotReport)
	}
	gotMax, err := rc.MaxReportLength(n)
	if err != nil {
		return nil, fmt.Errorf("failed to get MaxReportLength: %w", err)
	}
	if gotMax != mercury_v3_test.Fixtures.MaxReportLength {
		return nil, fmt.Errorf("expected MaxReportLength %d but got %d", max, gotMax)
	}
	gotObservedTimestamp, err := rc.ObservationTimestampFromReport(gotReport)
	if err != nil {
		return nil, fmt.Errorf("failed to get ObservationTimestampFromReport: %w", err)
	}
	if gotObservedTimestamp != mercury_v3_test.Fixtures.ObservationTimestamp {
		return nil, fmt.Errorf("expected ObservationTimestampFromReport %d but got %d", mercury_v3_test.Fixtures.ObservationTimestamp, gotObservedTimestamp)
	}

	occ := provider.OnchainConfigCodec()
	gotEncoded, err := occ.Encode(mercury_common_test.StaticOnChainConfigCodecFixtures.Decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to Encode: %w", err)
	}
	if !bytes.Equal(gotEncoded, mercury_common_test.StaticOnChainConfigCodecFixtures.Encoded) {
		return nil, fmt.Errorf("expected Encoded %s but got %s", encoded, gotEncoded)
	}
	gotDecoded, err := occ.Decode(mercury_common_test.StaticOnChainConfigCodecFixtures.Encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode: %w", err)
	}
	if !reflect.DeepEqual(gotDecoded, mercury_common_test.StaticOnChainConfigCodecFixtures.Decoded) {
		return nil, fmt.Errorf("expected OnchainConfig %s but got %s", onchainConfig, gotDecoded)
	}

	gotVal, err := dataSource.Observe(ctx, mercury_v3_test.Fixtures.ReportTimestamp, false)
	if err != nil {
		return nil, fmt.Errorf("failed to observe dataSource: %w", err)
	}
	if !assert.ObjectsAreEqual(mercury_v3_test.Fixtures.Observation, gotVal) {
		return nil, fmt.Errorf("expected Value %s but got %s", value, gotVal)
	}

	return staticPluginFactory{}, nil
}

type StaticMercuryProvider struct{}

var _ types.MercuryProvider = StaticMercuryProvider{}

func (s StaticMercuryProvider) Start(ctx context.Context) error { return nil }

func (s StaticMercuryProvider) Close() error { return nil }

func (s StaticMercuryProvider) Ready() error { panic("unimplemented") }

func (s StaticMercuryProvider) Name() string { panic("unimplemented") }

func (s StaticMercuryProvider) HealthReport() map[string]error { panic("unimplemented") }

func (s StaticMercuryProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return staticOffchainConfigDigester{}
}

func (s StaticMercuryProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return staticContractConfigTracker{}
}

func (s StaticMercuryProvider) ContractTransmitter() libocr.ContractTransmitter {
	return staticContractTransmitter{}
}

func (s StaticMercuryProvider) ReportCodecV1() mercury_v1_types.ReportCodec {
	return mercury_v1_test.StaticReportCodec{}
}

func (s StaticMercuryProvider) ReportCodecV2() mercury_v2_types.ReportCodec {
	return mercury_v2_test.StaticReportCodec{}
}

func (s StaticMercuryProvider) ReportCodecV3() mercury_v3_types.ReportCodec {
	return mercury_v3_test.StaticReportCodec{}
}

func (s StaticMercuryProvider) OnchainConfigCodec() mercury_types.OnchainConfigCodec {
	return mercury_common_test.StaticOnchainConfigCodec{}
}

func (s StaticMercuryProvider) MercuryChainReader() mercury_types.ChainReader {
	return mercury_common_test.StaticMercuryChainReader{}
}

func (s StaticMercuryProvider) ChainReader() types.ChainReader {
	panic("mercury does not use the general ChainReader interface yet")
}

func (s StaticMercuryProvider) MercuryServerFetcher() mercury_types.ServerFetcher {
	return mercury_common_test.StaticServerFetcher{}
}

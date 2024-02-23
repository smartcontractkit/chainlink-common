package median_test

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

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	pluginprovider_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/plugin_provider"
	reportingplugin_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/reporting_plugin"
	codec_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func PluginMedian(t *testing.T, p types.PluginMedian) {
	PluginMedianTest{&TestStaticMedianProvider}.TestPluginMedian(t, p)
}

type PluginMedianTest struct {
	types.MedianProvider
}

func (m PluginMedianTest) TestPluginMedian(t *testing.T, p types.PluginMedian) {
	t.Run("PluginMedian", func(t *testing.T) {
		ctx := tests.Context(t)
		factory, err := p.NewMedianFactory(ctx, m.MedianProvider, DefaultTestDataSource(), DefaultTestJuelsPerFeeCoinDataSource(), &StaticErrorLog{})
		require.NoError(t, err)

		ReportingPluginFactory(t, factory)
	})
}

func ReportingPluginFactory(t *testing.T, factory types.ReportingPluginFactory) {
	t.Run("ReportingPluginFactory", func(t *testing.T) {
		rp, gotRPI, err := factory.NewReportingPlugin(reportingPluginConfig)
		require.NoError(t, err)
		assert.Equal(t, rpi, gotRPI)
		t.Cleanup(func() { assert.NoError(t, rp.Close()) })
		t.Run("ReportingPlugin", func(t *testing.T) {
			ctx := tests.Context(t)

			reportingplugin_test.TestStaticReportingPlugin.Evaluate(t, ctx, rp)
		})
	})
}

/*
	func OCR3ReportingPluginFactory(t *testing.T, factory types.OCR3ReportingPluginFactory) {
		t.Run("OCR3ReportingPluginFactory", func(t *testing.T) {
			rp, gotRPI, err := factory.NewReportingPlugin(ocr3reportingPluginConfig)
			require.NoError(t, err)
			assert.Equal(t, ocr3rpi, gotRPI)
			t.Cleanup(func() { assert.NoError(t, rp.Close()) })
			t.Run("OCR3ReportingPlugin", func(t *testing.T) {
				ctx := tests.Context(t)

				gotQuery, err := rp.Query(ctx, outcomeContext)
				require.NoError(t, err)
				assert.Equal(t, query, []byte(gotQuery))

				gotObs, err := rp.Observation(ctx, outcomeContext, query)
				require.NoError(t, err)
				assert.Equal(t, observation, gotObs)

				err = rp.ValidateObservation(outcomeContext, query, ao)
				require.NoError(t, err)

				gotQuorum, err := rp.ObservationQuorum(outcomeContext, query)
				require.NoError(t, err)
				assert.Equal(t, quorum, gotQuorum)

				gotOutcome, err := rp.Outcome(outcomeContext, query, obs)
				require.NoError(t, err)
				assert.Equal(t, outcome, gotOutcome)

				gotRI, err := rp.Reports(seqNr, outcome)
				require.NoError(t, err)
				assert.Equal(t, RIs, gotRI)

				gotShouldAccept, err := rp.ShouldAcceptAttestedReport(ctx, seqNr, RI)
				require.NoError(t, err)
				assert.True(t, gotShouldAccept)

				gotShouldTransmit, err := rp.ShouldTransmitAcceptedReport(ctx, seqNr, RI)
				require.NoError(t, err)
				assert.True(t, gotShouldTransmit)
			})
		})
	}
*/
type StaticPluginMedianConfig struct {
	Provider                  types.MedianProvider
	DataSource                median.DataSource
	JuelsPerFeeCoinDataSource median.DataSource
	ErrorLog                  types.ErrorLog
}

type StaticPluginMedian struct {
	StaticPluginMedianConfig
}

func (s StaticPluginMedian) NewMedianFactory(ctx context.Context, provider types.MedianProvider, dataSource, juelsPerFeeCoinDataSource median.DataSource, errorLog types.ErrorLog) (types.ReportingPluginFactory, error) {
	// the provider may be a grpc client, so we can't compare it directly
	// but in all of these static tests, the implementation of the provider is expected
	// to be the same static implementation, so we can compare the expected values

	err := TestStaticMedianProvider.Equal(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("NewMedianFactory: provider does not equal a static median provider implementation: %w", err)
	}

	err = DefaultTestDataSource().Evaluate(ctx, dataSource)
	if err != nil {
		return nil, fmt.Errorf("NewMedianFactory: dataSource does not equal a static test data source implementation: %w", err)
	}

	err = DefaultTestJuelsPerFeeCoinDataSource().Evaluate(ctx, juelsPerFeeCoinDataSource)
	if err != nil {
		return nil, fmt.Errorf("NewMedianFactory: juelsPerFeeCoinDataSource does not equal a static test juels per fee coin data source implementation: %w", err)
	}

	if err := errorLog.SaveError(ctx, errMsg); err != nil {
		return nil, fmt.Errorf("failed to save error: %w", err)
	}
	return staticPluginFactory{}, nil
}

type staticPluginFactory struct {
	libocr.ReportingPluginConfig
}

func (s staticPluginFactory) Name() string { panic("implement me") }

func (s staticPluginFactory) Start(ctx context.Context) error {
	return nil
}

func (s staticPluginFactory) Close() error { return nil }

func (s staticPluginFactory) Ready() error { panic("implement me") }

func (s staticPluginFactory) HealthReport() map[string]error { panic("implement me") }

func (s staticPluginFactory) NewReportingPlugin(config libocr.ReportingPluginConfig) (libocr.ReportingPlugin, libocr.ReportingPluginInfo, error) {
	// lazy initialization
	s.ReportingPluginConfig = reportingPluginConfig

	if config.ConfigDigest != s.ConfigDigest {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected ConfigDigest %x but got %x", s.ConfigDigest, config.ConfigDigest)
	}
	if config.OracleID != s.OracleID {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected OracleID %d but got %d", s.OracleID, config.OracleID)
	}
	if config.F != s.F {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected F %d but got %d", s.F, config.F)
	}
	if config.N != s.N {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected N %d but got %d", s.N, config.N)
	}
	if !bytes.Equal(config.OnchainConfig, s.OnchainConfig) {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected OnchainConfig %x but got %x", s.OnchainConfig, config.OnchainConfig)
	}
	if !bytes.Equal(config.OffchainConfig, s.OffchainConfig) {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected OffchainConfig %x but got %x", s.OffchainConfig, config.OffchainConfig)
	}
	if config.EstimatedRoundInterval != s.EstimatedRoundInterval {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected EstimatedRoundInterval %d but got %d", s.EstimatedRoundInterval, config.EstimatedRoundInterval)
	}
	if config.MaxDurationQuery != s.MaxDurationQuery {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected MaxDurationQuery %d but got %d", s.MaxDurationQuery, config.MaxDurationQuery)
	}
	if config.MaxDurationReport != s.MaxDurationReport {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected MaxDurationReport %d but got %d", s.MaxDurationReport, config.MaxDurationReport)
	}
	if config.MaxDurationObservation != s.MaxDurationObservation {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected MaxDurationObservation %d but got %d", s.MaxDurationObservation, config.MaxDurationObservation)
	}
	if config.MaxDurationShouldAcceptFinalizedReport != s.MaxDurationShouldAcceptFinalizedReport {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected MaxDurationShouldAcceptFinalizedReport %d but got %d", s.MaxDurationShouldAcceptFinalizedReport, config.MaxDurationShouldAcceptFinalizedReport)
	}
	if config.MaxDurationShouldTransmitAcceptedReport != s.MaxDurationShouldTransmitAcceptedReport {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("expected MaxDurationShouldTransmitAcceptedReport %d but got %d", s.MaxDurationShouldTransmitAcceptedReport, config.MaxDurationShouldTransmitAcceptedReport)
	}

	return reportingplugin_test.TestStaticReportingPlugin, rpi, nil
}

type StaticMedianProviderConfig struct {
	// we use the static implementation type not the interface type
	// because we always expect the static implementation to be used
	// and it facilitates testing.
	OffchainDigester    pluginprovider_test.StaticOffchainConfigDigester
	ContractTracker     pluginprovider_test.StaticContractConfigTracker
	ContractTransmitter pluginprovider_test.StaticContractTransmitter
}

type StaticMedianProvider struct {
	StaticMedianProviderConfig
	rc  staticReportCodec
	mc  staticMedianContract
	ooc staticOnchainConfigCodec
}

func (s StaticMedianProvider) Start(ctx context.Context) error {
	// lazy initialization
	s.StaticMedianProviderConfig.OffchainDigester = pluginprovider_test.TestOffchainConfigDigester
	s.StaticMedianProviderConfig.ContractTracker = pluginprovider_test.TestContractConfigTracker
	s.StaticMedianProviderConfig.ContractTransmitter = pluginprovider_test.TestContractTransmitter
	s.rc = staticReportCodec{}
	s.mc = staticMedianContract{}
	s.ooc = staticOnchainConfigCodec{}
	return nil
}

func (s StaticMedianProvider) Close() error { return nil }

func (s StaticMedianProvider) Ready() error { panic("unimplemented") }

func (s StaticMedianProvider) Name() string { panic("unimplemented") }

func (s StaticMedianProvider) HealthReport() map[string]error { panic("unimplemented") }

func (s StaticMedianProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	//return staticOffchainConfigDigester{}
	return s.StaticMedianProviderConfig.OffchainDigester
}

func (s StaticMedianProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	//return staticContractConfigTracker{}
	return s.StaticMedianProviderConfig.ContractTracker
}

func (s StaticMedianProvider) ContractTransmitter() libocr.ContractTransmitter {
	//return staticContractTransmitter{}
	return s.StaticMedianProviderConfig.ContractTransmitter
}

func (s StaticMedianProvider) ReportCodec() median.ReportCodec { return s.rc }

func (s StaticMedianProvider) MedianContract() median.MedianContract {
	return s.mc
}

func (s StaticMedianProvider) OnchainConfigCodec() median.OnchainConfigCodec {
	return s.ooc
}

func (s StaticMedianProvider) ChainReader() types.ChainReader {
	return StaticChainReader{}
}

func (s StaticMedianProvider) Codec() types.Codec {
	return codec_test.StaticCodec{}
}

func (s StaticMedianProvider) Equal(ctx context.Context, provider types.MedianProvider) error {

	cr := provider.ChainReader()
	var gotLatestValue map[string]int

	err := cr.GetLatestValue(ctx, contractName, medianContractGenericMethod, getLatestValueParams, &gotLatestValue)
	if err != nil {
		fmt.Errorf("failed to call GetLatestValue() on median provider: %w", err)
	}

	if !assert.ObjectsAreEqual(gotLatestValue, latestValue) {
		fmt.Errorf("GetLatestValue: expected %v but got %v", gotLatestValue, latestValue)
	}

	ocd := provider.OffchainConfigDigester()
	err = s.OffchainDigester.Equal(ocd)
	if err != nil {
		return fmt.Errorf("providers offchain digester does not equal static offchain digester: %w", err)
	}

	cct := provider.ContractConfigTracker()
	err = s.ContractTracker.Equal(ctx, cct)
	if err != nil {
		return fmt.Errorf("providers contract config tracker does not equal static contract config tracker: %w", err)
	}

	ct := provider.ContractTransmitter()
	err = s.StaticMedianProviderConfig.ContractTransmitter.Equal(ctx, ct)
	if err != nil {
		return fmt.Errorf("providers contract transmitter does not equal static contract transmitter: %w", err)
	}

	rc := provider.ReportCodec()
	err = s.rc.Evaluate(ctx, rc)
	if err != nil {
		return fmt.Errorf("failed to evaluate report codec: %w", err)
	}

	mc := provider.MedianContract()
	err = s.mc.Evaluate(ctx, mc)
	if err != nil {
		return fmt.Errorf("failed to evaluate median contract: %w", err)
	}

	occ := provider.OnchainConfigCodec()
	err = s.ooc.Evaluate(ctx, occ)
	if err != nil {
		return fmt.Errorf("failed to evaluate onchain config codec: %w", err)
	}

	return nil

}

type staticReportCodec struct{}

// TODO remove hard coded values
func (s staticReportCodec) BuildReport(os []median.ParsedAttributedObservation) (libocr.Report, error) {
	if !assert.ObjectsAreEqual(pobs, os) {
		return nil, fmt.Errorf("expected observations %v but got %v", pobs, os)
	}
	return report, nil
}

func (s staticReportCodec) MedianFromReport(r libocr.Report) (*big.Int, error) {
	if !bytes.Equal(report, r) {
		return nil, fmt.Errorf("expected report %x but got %x", report, r)
	}
	return medianValue, nil
}

func (s staticReportCodec) MaxReportLength(n2 int) (int, error) {
	if n != n2 {
		return -1, fmt.Errorf("expected n %d but got %d", n, n2)
	}
	return max, nil
}

func (s staticReportCodec) Evaluate(ctx context.Context, rc median.ReportCodec) error {
	gotReport, err := rc.BuildReport(pobs)
	if err != nil {
		return fmt.Errorf("failed to BuildReport: %w", err)
	}
	if !bytes.Equal(gotReport, report) {
		return fmt.Errorf("expected Report %x but got %x", report, gotReport)
	}
	gotMedianValue, err := rc.MedianFromReport(report)
	if err != nil {
		return fmt.Errorf("failed to get MedianFromReport: %w", err)
	}
	if medianValue.Cmp(gotMedianValue) != 0 {
		return fmt.Errorf("expected MedianValue %s but got %s", medianValue, gotMedianValue)
	}
	gotMax, err := rc.MaxReportLength(n)
	if err != nil {
		return fmt.Errorf("failed to get MaxReportLength: %w", err)
	}
	if gotMax != max {
		return fmt.Errorf("expected MaxReportLength %d but got %d", max, gotMax)
	}
	return nil
}

type staticMedianContractConfig struct {
	configDigest     libocr.ConfigDigest
	epoch            uint32
	round            uint8
	latestAnswer     *big.Int
	latestTimestamp  time.Time
	lookbackDuration time.Duration
}

type staticMedianContract struct {
	staticMedianContractConfig
}

func (s staticMedianContract) LatestTransmissionDetails(ctx context.Context) (libocr.ConfigDigest, uint32, uint8, *big.Int, time.Time, error) {
	return s.configDigest, s.epoch, s.round, s.latestAnswer, s.latestTimestamp, nil
}

func (s staticMedianContract) LatestRoundRequested(ctx context.Context, lookback time.Duration) (libocr.ConfigDigest, uint32, uint8, error) {
	if s.lookbackDuration != lookback {
		return libocr.ConfigDigest{}, 0, 0, fmt.Errorf("expected lookback %s but got %s", s.lookbackDuration, lookback)
	}
	return s.configDigest, s.epoch, s.round, nil
}

func (s staticMedianContract) Evaluate(ctx context.Context, mc median.MedianContract) error {
	gotConfigDigest, gotEpoch, gotRound, err := mc.LatestRoundRequested(ctx, s.lookbackDuration)
	if err != nil {
		return fmt.Errorf("failed to get LatestRoundRequested: %w", err)
	}
	if gotConfigDigest != s.configDigest {
		return fmt.Errorf("expected ConfigDigest %s but got %s", s.configDigest, gotConfigDigest)
	}
	if gotEpoch != s.epoch {
		return fmt.Errorf("expected Epoch %d but got %d", s.epoch, gotEpoch)
	}
	if gotRound != s.round {
		return fmt.Errorf("expected Round %d but got %d", s.round, gotRound)
	}
	gotConfigDigest, gotEpoch, gotRound, gotLatestAnswer, gotLatestTimestamp, err := mc.LatestTransmissionDetails(ctx)
	if err != nil {
		return fmt.Errorf("failed to get LatestTransmissionDetails: %w", err)
	}
	if gotConfigDigest != s.configDigest {
		return fmt.Errorf("expected ConfigDigest %s but got %s", s.configDigest, gotConfigDigest)
	}
	if gotEpoch != s.epoch {
		return fmt.Errorf("expected Epoch %d but got %d", s.epoch, gotEpoch)
	}
	if gotRound != s.round {
		return fmt.Errorf("expected Round %d but got %d", s.round, gotRound)
	}
	if s.latestAnswer.Cmp(gotLatestAnswer) != 0 {
		return fmt.Errorf("expected LatestAnswer %s but got %s", s.latestAnswer, gotLatestAnswer)
	}
	if !gotLatestTimestamp.Equal(s.latestTimestamp) {
		return fmt.Errorf("expected LatestTimestamp %s but got %s", s.latestTimestamp, gotLatestTimestamp)
	}
	return nil
}

// TODO remove hard coded values
type staticOnchainConfigCodec struct{}

func (s staticOnchainConfigCodec) Encode(c median.OnchainConfig) ([]byte, error) {
	if !assert.ObjectsAreEqual(onchainConfig.Max, c.Max) {
		return nil, fmt.Errorf("expected max %s but got %s", onchainConfig.Max, c.Max)
	}
	if !assert.ObjectsAreEqual(onchainConfig.Min, c.Min) {
		return nil, fmt.Errorf("expected min %s but got %s", onchainConfig.Min, c.Min)
	}
	return encodedOnchainConfig, nil
}

func (s staticOnchainConfigCodec) Decode(b []byte) (median.OnchainConfig, error) {
	if !bytes.Equal(encodedOnchainConfig, b) {
		return median.OnchainConfig{}, fmt.Errorf("expected encoded %x but got %x", encodedOnchainConfig, b)
	}
	return onchainConfig, nil
}

func (s staticOnchainConfigCodec) Evaluate(ctx context.Context, occ median.OnchainConfigCodec) error {
	gotEncoded, err := occ.Encode(onchainConfig)
	if err != nil {
		return fmt.Errorf("failed to Encode: %w", err)
	}
	if !bytes.Equal(gotEncoded, encodedOnchainConfig) {
		return fmt.Errorf("expected Encoded %s but got %s", encodedOnchainConfig, gotEncoded)
	}
	gotDecoded, err := occ.Decode(encodedOnchainConfig)
	if err != nil {
		return fmt.Errorf("failed to Decode: %w", err)
	}
	if !reflect.DeepEqual(gotDecoded, onchainConfig) {
		return fmt.Errorf("expected OnchainConfig %s but got %s", onchainConfig, gotDecoded)
	}
	return nil
}

/*
// OCR3
type ocr3staticPluginFactory struct{}

var _ types.OCR3ReportingPluginFactory = (*ocr3staticPluginFactory)(nil)

func (o ocr3staticPluginFactory) Name() string { panic("implement me") }

func (o ocr3staticPluginFactory) Start(ctx context.Context) error { return nil }

func (o ocr3staticPluginFactory) Close() error { return nil }

func (o ocr3staticPluginFactory) Ready() error { panic("implement me") }

func (o ocr3staticPluginFactory) HealthReport() map[string]error { panic("implement me") }

func (o ocr3staticPluginFactory) NewReportingPlugin(config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[[]byte], ocr3types.ReportingPluginInfo, error) {
	if config.ConfigDigest != ocr3reportingPluginConfig.ConfigDigest {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected ConfigDigest %x but got %x", reportingPluginConfig.ConfigDigest, config.ConfigDigest)
	}
	if config.OracleID != ocr3reportingPluginConfig.OracleID {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected OracleID %d but got %d", reportingPluginConfig.OracleID, config.OracleID)
	}
	if config.F != ocr3reportingPluginConfig.F {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected F %d but got %d", reportingPluginConfig.F, config.F)
	}
	if config.N != ocr3reportingPluginConfig.N {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected N %d but got %d", reportingPluginConfig.N, config.N)
	}
	if !bytes.Equal(config.OnchainConfig, ocr3reportingPluginConfig.OnchainConfig) {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected OnchainConfig %x but got %x", ocr3reportingPluginConfig.OnchainConfig, config.OnchainConfig)
	}
	if !bytes.Equal(config.OffchainConfig, ocr3reportingPluginConfig.OffchainConfig) {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected OffchainConfig %x but got %x", ocr3reportingPluginConfig.OffchainConfig, config.OffchainConfig)
	}
	if config.EstimatedRoundInterval != ocr3reportingPluginConfig.EstimatedRoundInterval {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected EstimatedRoundInterval %d but got %d", ocr3reportingPluginConfig.EstimatedRoundInterval, config.EstimatedRoundInterval)
	}
	if config.MaxDurationQuery != ocr3reportingPluginConfig.MaxDurationQuery {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected MaxDurationQuery %d but got %d", ocr3reportingPluginConfig.MaxDurationQuery, config.MaxDurationQuery)
	}
	if config.MaxDurationObservation != ocr3reportingPluginConfig.MaxDurationObservation {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected MaxDurationObservation %d but got %d", ocr3reportingPluginConfig.MaxDurationObservation, config.MaxDurationObservation)
	}
	if config.MaxDurationShouldAcceptAttestedReport != ocr3reportingPluginConfig.MaxDurationShouldAcceptAttestedReport {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected MaxDurationShouldAcceptAttestedReport %d but got %d", ocr3reportingPluginConfig.MaxDurationShouldAcceptAttestedReport, config.MaxDurationShouldAcceptAttestedReport)
	}
	if config.MaxDurationShouldTransmitAcceptedReport != ocr3reportingPluginConfig.MaxDurationShouldTransmitAcceptedReport {
		return nil, ocr3types.ReportingPluginInfo{}, fmt.Errorf("expected MaxDurationShouldTransmitAcceptedReport %d but got %d", ocr3reportingPluginConfig.MaxDurationShouldTransmitAcceptedReport, config.MaxDurationShouldTransmitAcceptedReport)
	}

	return ocr3staticReportingPlugin{}, ocr3rpi, nil
}
*/

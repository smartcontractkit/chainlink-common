package median_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	errorlogtest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/errorlog/test"
	reportingplugintest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func PluginMedian(t *testing.T, p core.PluginMedian) {
	PluginMedianTest{MedianProvider(logger.Test(t))}.TestPluginMedian(t, p)
}

type PluginMedianTest struct {
	types.MedianProvider
}

func (m PluginMedianTest) TestPluginMedian(t *testing.T, p core.PluginMedian) {
	t.Run("PluginMedian", func(t *testing.T) {
		ctx := tests.Context(t)
		factory, err := p.NewMedianFactory(ctx, m.MedianProvider, MedianContractID, DataSource, JuelsPerFeeCoinDataSource, GasPriceSubunitsDataSource, &errorlogtest.ErrorLog)
		require.NoError(t, err)

		ReportingPluginFactory(t, factory)
		servicetest.AssertHealthReportNames(t, p.HealthReport(),
			"PluginMedianClient",
			"PluginMedianClient.staticMedianFactoryServer",
			"PluginMedianClient.staticMedianFactoryServer.staticReportingPluginFactory",
		)
	})

	// when gasPriceSubunitsDataSource is meant to trigger a no-op
	t.Run("PluginMedian (Zero GasPriceSubunitsDataSource)", func(t *testing.T) {
		ctx := tests.Context(t)
		factory, err := p.NewMedianFactory(ctx, m.MedianProvider, MedianContractID, DataSource, JuelsPerFeeCoinDataSource, &ZeroDataSource{}, &errorlogtest.ErrorLog)
		require.NoError(t, err)

		ReportingPluginFactory(t, factory)
		servicetest.AssertHealthReportNames(t, p.HealthReport(),
			"PluginMedianClient",
			"PluginMedianClient.staticMedianFactoryServer",
			"PluginMedianClient.staticMedianFactoryServer.staticReportingPluginFactory",
		)
	})
}

func ReportingPluginFactory(t *testing.T, factory types.ReportingPluginFactory) {
	t.Run("ReportingPluginFactory", func(t *testing.T) {
		// we expect the static implementation to be used under the covers
		// we can't compare the types directly because the returned reporting plugin may be a grpc client
		// that wraps the static implementation
		var expectedReportingPlugin = reportingplugintest.ReportingPlugin

		rp, gotRPI, err := factory.NewReportingPlugin(tests.Context(t), reportingPluginConfig)
		require.NoError(t, err)
		assert.Equal(t, rpi, gotRPI)
		t.Cleanup(func() { assert.NoError(t, rp.Close()) })
		t.Run("ReportingPlugin", func(t *testing.T) {
			ctx := tests.Context(t)

			expectedReportingPlugin.AssertEqual(ctx, t, rp)
		})
	})
}

type staticPluginMedianConfig struct {
	provider                   staticMedianProvider
	contractID                 string
	dataSource                 staticDataSource
	juelsPerFeeCoinDataSource  staticDataSource
	gasPriceSubunitsDataSource staticDataSource
	errorLog                   testtypes.ErrorLogEvaluator
}

type staticMedianFactoryServer struct {
	services.Service
	staticPluginMedianConfig
	reportingPluginFactory staticReportingPluginFactory
}

func newStaticMedianFactoryServer(lggr logger.Logger, cfg staticPluginMedianConfig) staticMedianFactoryServer {
	lggr = logger.Named(lggr, "staticMedianFactoryServer")
	return staticMedianFactoryServer{
		Service:                  test.NewStaticService(lggr),
		staticPluginMedianConfig: cfg,
		reportingPluginFactory:   newStaticReportingPluginFactory(lggr, reportingPluginConfig),
	}
}

var _ core.PluginMedian = staticMedianFactoryServer{}

func (s staticMedianFactoryServer) HealthReport() map[string]error {
	hp := s.Service.HealthReport()
	services.CopyHealth(hp, s.reportingPluginFactory.HealthReport())
	return hp
}

func (s staticMedianFactoryServer) NewMedianFactory(ctx context.Context, provider types.MedianProvider, contractID string, dataSource, juelsPerFeeCoinDataSource, gasPriceSubunitsDataSource median.DataSource, errorLog core.ErrorLog) (types.ReportingPluginFactory, error) {
	// the provider may be a grpc client, so we can't compare it directly
	// but in all of these static tests, the implementation of the provider is expected
	// to be the same static implementation, so we can compare the expected values

	err := s.provider.Evaluate(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("NewMedianFactory: provider does not equal a static median provider implementation: %w", err)
	}

	if s.contractID != contractID {
		return nil, errors.New("NewMedianFactory: contract address does not equal static median implementation")
	}

	err = s.dataSource.Evaluate(ctx, dataSource)
	if err != nil {
		return nil, fmt.Errorf("NewMedianFactory: dataSource does not equal a static test data source implementation: %w", err)
	}

	err = s.juelsPerFeeCoinDataSource.Evaluate(ctx, juelsPerFeeCoinDataSource)
	if err != nil {
		return nil, fmt.Errorf("NewMedianFactory: juelsPerFeeCoinDataSource does not equal a static test juels per fee coin data source implementation: %w", err)
	}

	err = s.gasPriceSubunitsDataSource.Evaluate(ctx, gasPriceSubunitsDataSource)

	if err != nil {
		var compareError *CompareError
		isCompareError := errors.As(err, &compareError)
		// allow 0 as valid data source value with the same staticMedianFactoryServer (because it is only defined once as a global var for all tests)
		if !(isCompareError && compareError.GotZero()) {
			return nil, fmt.Errorf("NewMedianFactory: gasPriceSubunitsDataSource does not equal a static gas price subunits data source implementation: %w", err)
		}
	}

	if err := errorLog.SaveError(ctx, "an error"); err != nil {
		return nil, fmt.Errorf("failed to save error: %w", err)
	}
	return s.reportingPluginFactory, nil
}

type staticReportingPluginFactory struct {
	services.Service
	libocr.ReportingPluginConfig
}

func newStaticReportingPluginFactory(lggr logger.Logger, cfg libocr.ReportingPluginConfig) staticReportingPluginFactory {
	lggr = logger.Named(lggr, "staticReportingPluginFactory")
	return staticReportingPluginFactory{
		Service:               test.NewStaticService(lggr),
		ReportingPluginConfig: cfg,
	}
}

func (s staticReportingPluginFactory) NewReportingPlugin(ctx context.Context, config libocr.ReportingPluginConfig) (libocr.ReportingPlugin, libocr.ReportingPluginInfo, error) {
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

	return reportingplugintest.ReportingPlugin, rpi, nil
}

type staticMedianProviderConfig struct {
	// we use the static implementation type not the interface type
	// because we always expect the static implementation to be used
	// and it facilitates testing.
	offchainDigester    testtypes.OffchainConfigDigesterEvaluator
	contractTracker     testtypes.ContractConfigTrackerEvaluator
	contractTransmitter testtypes.ContractTransmitterEvaluator
	reportCodec         staticReportCodec
	medianContract      staticMedianContract
	onchainConfigCodec  staticOnchainConfigCodec
	contractReader      testtypes.ContractReaderTester
	codec               testtypes.CodecEvaluator
}

// implements types.MedianProvider and testtypes.Evaluator[types.MedianProvider]
type staticMedianProvider struct {
	services.Service
	staticMedianProviderConfig
}

func newStaticMedianProvider(lggr logger.Logger, cfg staticMedianProviderConfig) staticMedianProvider {
	lggr = logger.Named(lggr, "staticMedianProvider")
	return staticMedianProvider{
		Service:                    test.NewStaticService(lggr),
		staticMedianProviderConfig: cfg,
	}
}

var _ testtypes.MedianProviderTester = staticMedianProvider{}

func (s staticMedianProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return s.offchainDigester
}

func (s staticMedianProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return s.contractTracker
}

func (s staticMedianProvider) ContractTransmitter() libocr.ContractTransmitter {
	return s.contractTransmitter
}

func (s staticMedianProvider) ReportCodec() median.ReportCodec { return s.reportCodec }

func (s staticMedianProvider) MedianContract() median.MedianContract {
	return s.medianContract
}

func (s staticMedianProvider) OnchainConfigCodec() median.OnchainConfigCodec {
	return s.onchainConfigCodec
}

func (s staticMedianProvider) ContractReader() types.ContractReader {
	return s.contractReader
}

func (s staticMedianProvider) Codec() types.Codec {
	return s.codec
}

func (s staticMedianProvider) AssertEqual(ctx context.Context, t *testing.T, provider types.MedianProvider) {
	t.Run("OffchainConfigDigester", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.offchainDigester.Evaluate(ctx, provider.OffchainConfigDigester()))
	})

	t.Run("ContractConfigTracker", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.staticMedianProviderConfig.contractTracker.Evaluate(ctx, provider.ContractConfigTracker()))
	})

	t.Run("ContractTransmitter", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.staticMedianProviderConfig.contractTransmitter.Evaluate(ctx, provider.ContractTransmitter()))
	})

	t.Run("ReportCodec", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.reportCodec.Evaluate(ctx, provider.ReportCodec()))
	})

	t.Run("MedianContract", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.medianContract.Evaluate(ctx, provider.MedianContract()))
	})

	t.Run("OnchainConfigCodec", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.onchainConfigCodec.Evaluate(ctx, provider.OnchainConfigCodec()))
	})
}

func (s staticMedianProvider) Evaluate(ctx context.Context, provider types.MedianProvider) error {
	ocd := provider.OffchainConfigDigester()
	err := s.offchainDigester.Evaluate(ctx, ocd)
	if err != nil {
		return fmt.Errorf("providers offchain digester does not equal static offchain digester: %w", err)
	}

	cct := provider.ContractConfigTracker()
	err = s.contractTracker.Evaluate(ctx, cct)
	if err != nil {
		return fmt.Errorf("providers contract config tracker does not equal static contract config tracker: %w", err)
	}

	ct := provider.ContractTransmitter()
	err = s.staticMedianProviderConfig.contractTransmitter.Evaluate(ctx, ct)
	if err != nil {
		return fmt.Errorf("providers contract transmitter does not equal static contract transmitter: %w", err)
	}

	rc := provider.ReportCodec()
	err = s.reportCodec.Evaluate(ctx, rc)
	if err != nil {
		return fmt.Errorf("failed to evaluate report codec: %w", err)
	}

	mc := provider.MedianContract()
	err = s.medianContract.Evaluate(ctx, mc)
	if err != nil {
		return fmt.Errorf("failed to evaluate median contract: %w", err)
	}

	occ := provider.OnchainConfigCodec()
	err = s.onchainConfigCodec.Evaluate(ctx, occ)
	if err != nil {
		return fmt.Errorf("failed to evaluate onchain config codec: %w", err)
	}

	cr := provider.ContractReader()
	err = s.contractReader.Evaluate(ctx, cr)
	if err != nil {
		return fmt.Errorf("providers contract reader does not equal static contract reader: %w", err)
	}

	return nil
}

// implements median.ReportCodec and testtypes.Evaluator[median.ReportCodec]
type staticReportCodec struct{}

var _ testtypes.Evaluator[median.ReportCodec] = staticReportCodec{}
var _ median.ReportCodec = staticReportCodec{}

// TODO BCF-3068 remove hard coded values, use the staticXXXConfig pattern elsewhere in the test framework
func (s staticReportCodec) BuildReport(ctx context.Context, os []median.ParsedAttributedObservation) (libocr.Report, error) {
	if !assert.ObjectsAreEqual(pobs, os) {
		return nil, fmt.Errorf("expected observations %v but got %v", pobs, os)
	}
	return report, nil
}

func (s staticReportCodec) MedianFromReport(ctx context.Context, r libocr.Report) (*big.Int, error) {
	if !bytes.Equal(report, r) {
		return nil, fmt.Errorf("expected report %x but got %x", report, r)
	}
	return medianValue, nil
}

func (s staticReportCodec) MaxReportLength(ctx context.Context, n2 int) (int, error) {
	if n != n2 {
		return -1, fmt.Errorf("expected n %d but got %d", n, n2)
	}
	return max, nil
}

func (s staticReportCodec) Evaluate(ctx context.Context, rc median.ReportCodec) error {
	gotReport, err := rc.BuildReport(ctx, pobs)
	if err != nil {
		return fmt.Errorf("failed to BuildReport: %w", err)
	}
	if !bytes.Equal(gotReport, report) {
		return fmt.Errorf("expected Report %x but got %x", report, gotReport)
	}
	gotMedianValue, err := rc.MedianFromReport(ctx, report)
	if err != nil {
		return fmt.Errorf("failed to get MedianFromReport: %w", err)
	}
	if medianValue.Cmp(gotMedianValue) != 0 {
		return fmt.Errorf("expected MedianValue %s but got %s", medianValue, gotMedianValue)
	}
	gotMax, err := rc.MaxReportLength(ctx, n)
	if err != nil {
		return fmt.Errorf("failed to get MaxReportLength: %w", err)
	}
	if gotMax != max {
		return fmt.Errorf("expected MaxReportLength %d but got %d", max, gotMax)
	}
	return nil
}

// configuration for the static median provider
type staticMedianContractConfig struct {
	configDigest     libocr.ConfigDigest
	epoch            uint32
	round            uint8
	latestAnswer     *big.Int
	latestTimestamp  time.Time
	lookbackDuration time.Duration
}

// implements median.MedianContract and testtypes.Evaluator[median.MedianContract]
type staticMedianContract struct {
	staticMedianContractConfig
}

var _ testtypes.Evaluator[median.MedianContract] = (*staticMedianContract)(nil)
var _ median.MedianContract = (*staticMedianContract)(nil)

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

// implements median.OnchainConfigCodec and testtypes.Evaluator[median.OnchainConfigCodec]
type staticOnchainConfigCodec struct{}

var _ testtypes.Evaluator[median.OnchainConfigCodec] = staticOnchainConfigCodec{}
var _ median.OnchainConfigCodec = staticOnchainConfigCodec{}

func (s staticOnchainConfigCodec) Encode(ctx context.Context, c median.OnchainConfig) ([]byte, error) {
	if !assert.ObjectsAreEqual(onchainConfig.Max, c.Max) {
		return nil, fmt.Errorf("expected max %s but got %s", onchainConfig.Max, c.Max)
	}
	if !assert.ObjectsAreEqual(onchainConfig.Min, c.Min) {
		return nil, fmt.Errorf("expected min %s but got %s", onchainConfig.Min, c.Min)
	}
	return encodedOnchainConfig, nil
}

func (s staticOnchainConfigCodec) Decode(ctx context.Context, b []byte) (median.OnchainConfig, error) {
	if !bytes.Equal(encodedOnchainConfig, b) {
		return median.OnchainConfig{}, fmt.Errorf("expected encoded %x but got %x", encodedOnchainConfig, b)
	}
	return onchainConfig, nil
}

func (s staticOnchainConfigCodec) Evaluate(ctx context.Context, occ median.OnchainConfigCodec) error {
	gotEncoded, err := occ.Encode(ctx, onchainConfig)
	if err != nil {
		return fmt.Errorf("failed to Encode: %w", err)
	}
	if !bytes.Equal(gotEncoded, encodedOnchainConfig) {
		return fmt.Errorf("expected Encoded %s but got %s", encodedOnchainConfig, gotEncoded)
	}
	gotDecoded, err := occ.Decode(ctx, encodedOnchainConfig)
	if err != nil {
		return fmt.Errorf("failed to Decode: %w", err)
	}
	if !reflect.DeepEqual(gotDecoded, onchainConfig) {
		return fmt.Errorf("expected OnchainConfig %s but got %s", onchainConfig, gotDecoded)
	}
	return nil
}

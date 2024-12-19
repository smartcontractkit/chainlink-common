package test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccip"
	looptest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	cciptypes "github.com/smartcontractkit/chainlink-common/pkg/types/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestStaticCommitStore(t *testing.T) {
	t.Parallel()

	// static test implementation is self consistent
	ctx := context.Background()
	csr := CommitStoreReader(logger.Test(t))
	assert.NoError(t, csr.Evaluate(ctx, csr))

	// error when the test implementation is evaluates something that differs from the static implementation
	botched := CommitStoreReader(logger.Test(t))
	botched.changeConfigResponse = "not the right conifg"
	err := csr.Evaluate(ctx, botched)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not the right conifg")
}

func TestCommitStoreGRPC(t *testing.T) {
	t.Parallel()

	scaffold := looptest.NewGRPCScaffold(t, setupCommitStoreServer, ccip.NewCommitStoreReaderGRPCClient)
	roundTripCommitStoreTests(t, scaffold.Client())
	// commit store implements dependency management, test that it closes properly
	t.Run("Dependency management", func(t *testing.T) {
		d := &looptest.MockDep{}
		scaffold.Server().AddDep(d)
		assert.False(t, d.IsClosed())
		scaffold.Client().Close()
		assert.True(t, d.IsClosed())
	})
}

// roundTripCommitStoreTests tests the round trip of the client<->server.
// it should exercise all the methods of the client.
// do not add client.Close to this test, test that from the driver test
func roundTripCommitStoreTests(t *testing.T, client cciptypes.CommitStoreReader) {
	t.Run("ChangeConfig", func(t *testing.T) {
		csr := CommitStoreReader(logger.Test(t))
		gotAddr, err := client.ChangeConfig(tests.Context(t), csr.changeConfigRequest.onchainConfig, csr.changeConfigRequest.offchainConfig)
		require.NoError(t, err)
		assert.Equal(t, csr.changeConfigResponse, gotAddr)
	})

	t.Run("DecodeCommitReport", func(t *testing.T) {
		csr := CommitStoreReader(logger.Test(t))
		report, err := client.DecodeCommitReport(tests.Context(t), csr.decodeCommitReportRequest)
		require.NoError(t, err)
		if !reflect.DeepEqual(csr.decodeCommitReportResponse, report) {
			t.Errorf("expected %v, got %v", csr.decodeCommitReportResponse, report)
		}
	})

	// reuse the test data for the encode method
	t.Run("EncodeCommtReport", func(t *testing.T) {
		csr := CommitStoreReader(logger.Test(t))
		report, err := client.EncodeCommitReport(tests.Context(t), csr.decodeCommitReportResponse)
		require.NoError(t, err)
		assert.Equal(t, csr.decodeCommitReportRequest, report)
	})

	// exercise all the gas price estimator methods
	t.Run("GasPriceEstimator", func(t *testing.T) {
		estimator, err := client.GasPriceEstimator(tests.Context(t))
		require.NoError(t, err)

		t.Run("GetGasPrice", func(t *testing.T) {
			price, err := estimator.GetGasPrice(tests.Context(t))
			require.NoError(t, err)
			assert.Equal(t, GasPriceEstimatorCommit.getGasPriceResponse, price)
		})

		t.Run("DenoteInUSD", func(t *testing.T) {
			ctx := tests.Context(t)
			usd, err := estimator.DenoteInUSD(ctx, GasPriceEstimatorCommit.denoteInUSDRequest.p, GasPriceEstimatorCommit.denoteInUSDRequest.wrappedNativePrice)
			require.NoError(t, err)
			assert.Equal(t, GasPriceEstimatorCommit.denoteInUSDResponse.result, usd)
		})

		t.Run("Deviates", func(t *testing.T) {
			ctx := tests.Context(t)
			deviates, err := estimator.Deviates(ctx, GasPriceEstimatorCommit.deviatesRequest.p1, GasPriceEstimatorCommit.deviatesRequest.p2)
			require.NoError(t, err)
			assert.Equal(t, GasPriceEstimatorCommit.deviatesResponse, deviates)
		})

		t.Run("Median", func(t *testing.T) {
			ctx := tests.Context(t)
			median, err := estimator.Median(ctx, GasPriceEstimatorCommit.medianRequest.gasPrices)
			require.NoError(t, err)
			assert.Equal(t, GasPriceEstimatorCommit.medianResponse, median)
		})
	})

	t.Run("GetAcceptedCommitReportGteTimestamp", func(t *testing.T) {
		csr := CommitStoreReader(logger.Test(t))
		report, err := client.GetAcceptedCommitReportsGteTimestamp(tests.Context(t),
			csr.getAcceptedCommitReportsGteTimestampRequest.timestamp,
			csr.getAcceptedCommitReportsGteTimestampRequest.confirmations)
		require.NoError(t, err)
		if !reflect.DeepEqual(csr.getAcceptedCommitReportsGteTimestampResponse, report) {
			t.Errorf("expected %v, got %v", csr.getAcceptedCommitReportsGteTimestampResponse, report)
		}
	})

	t.Run("GetCommitReportMatchingSeqNum", func(t *testing.T) {
		csr := CommitStoreReader(logger.Test(t))
		report, err := client.GetCommitReportMatchingSeqNum(tests.Context(t),
			csr.getCommitReportMatchingSeqNumRequest.seqNum,
			csr.getCommitReportMatchingSeqNumRequest.confirmations)
		require.NoError(t, err)
		// use the same response as the reportsGteTimestamp for simplicity
		if !reflect.DeepEqual(csr.getAcceptedCommitReportsGteTimestampResponse, report) {
			t.Errorf("expected %v, got %v", csr.getAcceptedCommitReportsGteTimestampRequest, report)
		}
	})

	t.Run("GetCommitStoreStaticConfig", func(t *testing.T) {
		config, err := client.GetCommitStoreStaticConfig(tests.Context(t))
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader(logger.Test(t)).getCommitStoreStaticConfigResponse, config)
	})

	t.Run("GetExpectedNextSequenceNumber", func(t *testing.T) {
		seq, err := client.GetExpectedNextSequenceNumber(tests.Context(t))
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader(logger.Test(t)).getExpectedNextSequenceNumberResponse, seq)
	})

	t.Run("GetLatestPriceEpochAndRound", func(t *testing.T) {
		got, err := client.GetLatestPriceEpochAndRound(tests.Context(t))
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader(logger.Test(t)).getLatestPriceEpochAndRoundResponse, got)
	})

	t.Run("IsBlessed", func(t *testing.T) {
		csr := CommitStoreReader(logger.Test(t))
		got, err := client.IsBlessed(tests.Context(t), csr.isBlessedRequest)
		require.NoError(t, err)
		assert.Equal(t, csr.isBlessedResponse, got)
	})

	t.Run("IsDestChainHealthy", func(t *testing.T) {
		got, err := client.IsDestChainHealthy(tests.Context(t))
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader(logger.Test(t)).isDestChainHealthyResponse, got)
	})

	t.Run("IsDown", func(t *testing.T) {
		got, err := client.IsDown(tests.Context(t))
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader(logger.Test(t)).isDownResponse, got)
	})

	t.Run("OffchainConfig", func(t *testing.T) {
		config, err := client.OffchainConfig(tests.Context(t))
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader(logger.Test(t)).offchainConfigResponse, config)
	})

	t.Run("VerifyExecutionReport", func(t *testing.T) {
		csr := CommitStoreReader(logger.Test(t))
		got, err := client.VerifyExecutionReport(tests.Context(t), csr.verifyExecutionReportRequest)
		require.NoError(t, err)
		assert.Equal(t, csr.verifyExecutionReportResponse, got)
	})
}

func setupCommitStoreServer(t *testing.T, s *grpc.Server, b *loopnet.BrokerExt) *ccip.CommitStoreGRPCServer {
	ctx := tests.Context(t)
	commitProvider, err := ccip.NewCommitStoreReaderGRPCServer(ctx, CommitStoreReader(logger.Test(t)), b)
	require.NoError(t, err)
	ccippb.RegisterCommitStoreReaderServer(s, commitProvider)
	return commitProvider
}

var _ looptest.SetupGRPCServer[*ccip.CommitStoreGRPCServer] = setupCommitStoreServer
var _ looptest.SetupGRPCClient[*ccip.CommitStoreGRPCClient] = ccip.NewCommitStoreReaderGRPCClient
